import { loadConfig, createSql, createLogger } from "./config.js";
import { run, type TaskList } from "graphile-worker";
import { NotificationRepo } from "./repository/notification-repo.js";
import { TaskRepo } from "./repository/task-repo.js";

async function main() {
  const cfg = loadConfig();
  const logger = createLogger(cfg.logLevel);
  const sql = createSql(cfg.databaseUrl);

  // Test database connection
  try {
    await sql`SELECT 1`;
    logger.info("worker connected to database");
  } catch (err) {
    logger.error({ error: err }, "failed to connect to database");
    process.exit(1);
  }

  const taskRepo = new TaskRepo(sql);
  const notifRepo = new NotificationRepo(sql);

  // Define graphile-worker task handlers
  const taskList: TaskList = {
    notification_process: async (payload) => {
      const { notification_id } = payload as { notification_id: number };
      logger.info({ notification_id }, "processing notification");
      // Simulate sending notification
      await new Promise((resolve) => setTimeout(resolve, 100));
      logger.info({ notification_id }, "notification processed successfully");
    },

    notification_batch: async (payload) => {
      let { batch_size } = payload as { batch_size: number };
      if (!batch_size || batch_size <= 0) batch_size = 10;

      logger.info({ batch_size }, "fetching pending notifications");

      const notifications = await notifRepo.fetchPending(batch_size);
      if (notifications.length === 0) {
        logger.debug("no pending notifications");
        return;
      }

      logger.info({ count: notifications.length }, "processing notification batch");

      for (const n of notifications) {
        logger.info(
          { id: n.id, eventType: n.eventType, workspaceId: n.workspaceId, retryCount: n.retryCount },
          "sending notification",
        );

        // Simulate sending
        await new Promise((resolve) => setTimeout(resolve, 50));

        try {
          await notifRepo.markProcessed(n.id);
          logger.info({ id: n.id }, "notification sent");
        } catch (markErr) {
          logger.error({ id: n.id, error: markErr }, "failed to mark notification processed");
          try {
            await notifRepo.markFailed(n.id, String(markErr));
          } catch {
            // ignore
          }
        }
      }
    },

    bulk_import: async (payload) => {
      const { workspace_id, project_id, tasks: taskInputs } = payload as {
        workspace_id: string;
        project_id: string;
        tasks: Array<{
          title: string;
          description: string;
          priority: string;
          assigned_to: string;
          metadata?: Record<string, unknown>;
        }>;
      };

      logger.info(
        { workspace_id, project_id, total_tasks: taskInputs.length },
        "starting async bulk import",
      );

      let succeeded = 0;
      let failed = 0;

      for (let i = 0; i < taskInputs.length; i++) {
        const input = taskInputs[i];
        try {
          const task = await taskRepo.create({
            workspaceId: workspace_id,
            projectId: project_id,
            title: input.title,
            description: input.description,
            priority: input.priority,
            assignedTo: input.assigned_to,
            dueDate: null,
            metadata: input.metadata ?? null,
          });

          succeeded++;

          // Enqueue notification for imported task
          try {
            await notifRepo.create({
              workspaceId: workspace_id,
              eventType: "task.imported",
              payload: { task_id: task.id, title: task.title },
            });
          } catch {
            // fire-and-forget
          }
        } catch (err) {
          failed++;
          logger.warn({ index: i, error: err }, "import task failed");
        }
      }

      logger.info(
        { total: taskInputs.length, succeeded, failed },
        "async bulk import completed",
      );
    },
  };

  // Start graphile-worker
  const runner = await run({
    connectionString: cfg.databaseUrl,
    concurrency: cfg.workerConcurrency,
    taskList,
  });

  logger.info({ concurrency: cfg.workerConcurrency }, "worker started");

  // Graceful shutdown
  const shutdown = async () => {
    logger.info("worker shutting down");
    await runner.stop();
    await sql.end();
    logger.info("worker stopped");
    process.exit(0);
  };

  process.on("SIGINT", shutdown);
  process.on("SIGTERM", shutdown);
}

main();
