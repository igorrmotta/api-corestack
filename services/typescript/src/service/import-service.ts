import PQueue from "p-queue";
import { NotificationRepo } from "../repository/notification-repo.js";
import {
  TaskRepo,
  type ImportError,
  type ImportResult,
  type TaskInput,
} from "../repository/task-repo.js";
import type { Logger } from "pino";

export class ImportService {
  private concurrency: number;
  private rateLimit: number;

  constructor(
    private taskRepo: TaskRepo,
    private notifRepo: NotificationRepo | null,
    private logger: Logger,
    concurrency = 10,
    rateLimit = 100,
  ) {
    this.concurrency = concurrency > 0 ? concurrency : 10;
    this.rateLimit = rateLimit > 0 ? rateLimit : 100;
  }

  async bulkImport(
    workspaceId: string,
    projectId: string,
    inputs: TaskInput[],
  ): Promise<ImportResult> {
    const result: ImportResult = {
      total: inputs.length,
      succeeded: 0,
      failed: 0,
      errors: [],
    };

    if (inputs.length === 0) {
      return result;
    }

    this.logger.info(
      { workspaceId, projectId, total: inputs.length, concurrency: this.concurrency },
      "starting bulk import",
    );

    const queue = new PQueue({
      concurrency: this.concurrency,
      interval: 1000,
      intervalCap: this.rateLimit,
    });

    const errors: ImportError[] = [];

    const promises = inputs.map((input, i) =>
      queue.add(async () => {
        try {
          const task = await this.taskRepo.create({
            workspaceId,
            projectId,
            title: input.title,
            description: input.description,
            priority: input.priority,
            assignedTo: input.assignedTo,
            dueDate: input.dueDate,
            metadata: input.metadata,
          });

          // Enqueue notification for each imported task
          if (this.notifRepo) {
            try {
              await this.notifRepo.create({
                workspaceId,
                eventType: "task.imported",
                payload: { task_id: task.id, title: task.title },
              });
            } catch {
              // fire-and-forget
            }
          }

          result.succeeded++;
        } catch (err) {
          result.failed++;
          errors.push({
            index: i,
            error: err instanceof Error ? err.message : String(err),
          });
        }
      }),
    );

    await Promise.all(promises);
    result.errors = errors;

    this.logger.info(
      { total: result.total, succeeded: result.succeeded, failed: result.failed },
      "bulk import completed",
    );

    return result;
  }
}
