package com.corestack

import com.corestack.connect.ConnectAdapter
import com.corestack.handler.*
import com.corestack.middleware.loggingInterceptor
import com.corestack.middleware.recoveryInterceptor
import com.corestack.repository.*
import com.corestack.service.*
import io.github.oshai.kotlinlogging.KotlinLogging
import io.ktor.http.*
import io.ktor.server.engine.*
import io.ktor.server.netty.*
import io.ktor.server.response.*
import io.ktor.server.routing.*

private val logger = KotlinLogging.logger {}

fun startServer() {
    val config = loadConfig()
    val dataSource = createDataSource(config)
    val jdbi = createJdbi(dataSource)

    logger.info { "connected to database" }

    // Initialize repositories
    val workspaceRepo = WorkspaceRepo(jdbi)
    val projectRepo = ProjectRepo(jdbi)
    val taskRepo = TaskRepo(jdbi)
    val commentRepo = CommentRepo(jdbi)
    val notifRepo = NotificationRepo(jdbi)

    // Initialize services
    val workspaceSvc = WorkspaceService(workspaceRepo)
    val projectSvc = ProjectService(projectRepo)
    val taskSvc = TaskService(taskRepo, notifRepo)
    val commentSvc = CommentService(commentRepo)
    val importSvc = ImportService(taskRepo, notifRepo, config.workerConcurrency)

    // Create Connect adapter
    val adapter = ConnectAdapter()
    adapter.addInterceptor(loggingInterceptor())
    adapter.addInterceptor(recoveryInterceptor())

    // Register handlers
    registerWorkspaceHandlers(adapter, workspaceSvc)
    registerProjectHandlers(adapter, projectSvc)
    registerTaskHandlers(adapter, taskSvc, importSvc)
    registerCommentHandlers(adapter, commentSvc)
    registerNotificationHandlers(adapter, notifRepo)

    // Start Ktor server
    val server = embeddedServer(Netty, port = config.grpcPort) {
        routing {
            adapter.install(this)

            get("/health") {
                call.respondText("""{"status":"ok"}""", ContentType.Application.Json)
            }
        }
    }

    // Graceful shutdown
    Runtime.getRuntime().addShutdownHook(Thread {
        logger.info { "shutting down" }
        server.stop(1000, 5000)
        dataSource.close()
        logger.info { "server stopped" }
    })

    logger.info { "server starting addr=:${config.grpcPort}" }
    server.start(wait = true)
}
