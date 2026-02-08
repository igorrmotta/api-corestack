import http from "node:http";
import { loadConfig, createSql, createLogger } from "./config.js";
import { connectNodeAdapter } from "@connectrpc/connect-node";
import { WorkspaceRepo } from "./repository/workspace-repo.js";
import { ProjectRepo } from "./repository/project-repo.js";
import { TaskRepo } from "./repository/task-repo.js";
import { CommentRepo } from "./repository/comment-repo.js";
import { NotificationRepo } from "./repository/notification-repo.js";
import { WorkspaceService } from "./service/workspace-service.js";
import { ProjectService } from "./service/project-service.js";
import { TaskService } from "./service/task-service.js";
import { CommentService } from "./service/comment-service.js";
import { ImportService } from "./service/import-service.js";
import { registerWorkspaceHandler } from "./handler/workspace-handler.js";
import { registerProjectHandler } from "./handler/project-handler.js";
import { registerTaskHandler } from "./handler/task-handler.js";
import { registerCommentHandler } from "./handler/comment-handler.js";
import { registerNotificationHandler } from "./handler/notification-handler.js";
import { loggingInterceptor } from "./middleware/logging.js";
import { recoveryInterceptor } from "./middleware/recovery.js";

async function main() {
  const cfg = loadConfig();
  const logger = createLogger(cfg.logLevel);
  const sql = createSql(cfg.databaseUrl);

  // Test database connection
  try {
    await sql`SELECT 1`;
    logger.info("connected to database");
  } catch (err) {
    logger.error({ error: err }, "failed to connect to database");
    process.exit(1);
  }

  // Initialize repositories
  const workspaceRepo = new WorkspaceRepo(sql);
  const projectRepo = new ProjectRepo(sql);
  const taskRepo = new TaskRepo(sql);
  const commentRepo = new CommentRepo(sql);
  const notifRepo = new NotificationRepo(sql);

  // Initialize services
  const workspaceSvc = new WorkspaceService(workspaceRepo, logger);
  const projectSvc = new ProjectService(projectRepo, logger);
  const taskSvc = new TaskService(taskRepo, notifRepo, logger);
  const commentSvc = new CommentService(commentRepo, logger);
  const importSvc = new ImportService(taskRepo, notifRepo, logger, cfg.workerConcurrency, 100);

  // Create HTTP handler with Connect adapter
  const handler = connectNodeAdapter({
    routes(router) {
      registerWorkspaceHandler(router, workspaceSvc);
      registerProjectHandler(router, projectSvc);
      registerTaskHandler(router, taskSvc, importSvc);
      registerCommentHandler(router, commentSvc);
      registerNotificationHandler(router, notifRepo);
    },
    interceptors: [loggingInterceptor(logger), recoveryInterceptor(logger)],
  });

  // Create HTTP server (supports both HTTP/1.1 and Connect protocol)
  const server = http.createServer((req, res) => {
    // Health check endpoint
    if (req.url === "/health") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end('{"status":"ok"}\n');
      return;
    }
    handler(req, res);
  });

  const addr = cfg.grpcPort;
  server.listen(addr, () => {
    logger.info({ addr }, "server starting");
  });

  // Graceful shutdown
  const shutdown = () => {
    logger.info("shutting down");
    server.close(() => {
      sql.end().then(() => {
        logger.info("server stopped");
        process.exit(0);
      });
    });
  };

  process.on("SIGINT", shutdown);
  process.on("SIGTERM", shutdown);
}

main();
