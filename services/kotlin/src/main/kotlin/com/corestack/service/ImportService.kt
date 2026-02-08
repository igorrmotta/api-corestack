package com.corestack.service

import com.corestack.repository.*
import io.github.oshai.kotlinlogging.KotlinLogging
import kotlinx.coroutines.async
import kotlinx.coroutines.awaitAll
import kotlinx.coroutines.coroutineScope
import kotlinx.coroutines.sync.Semaphore
import kotlinx.coroutines.sync.withPermit
import java.util.UUID
import java.util.concurrent.atomic.AtomicInteger

private val logger = KotlinLogging.logger {}

data class ImportError(val index: Int, val error: String)
data class ImportResult(val total: Int, val succeeded: Int, val failed: Int, val errors: List<ImportError>)

data class TaskInput(
    val title: String,
    val description: String = "",
    val priority: String = "",
    val assignedTo: String = "",
    val dueDate: java.time.Instant? = null,
    val metadata: String? = null
)

class ImportService(
    private val taskRepo: TaskRepo,
    private val notifRepo: NotificationRepo?,
    private val concurrency: Int = 10
) {

    suspend fun bulkImport(
        workspaceId: UUID,
        projectId: UUID,
        inputs: List<TaskInput>
    ): ImportResult {
        if (inputs.isEmpty()) {
            return ImportResult(total = 0, succeeded = 0, failed = 0, errors = emptyList())
        }

        logger.info { "starting bulk import workspace_id=$workspaceId project_id=$projectId total=${inputs.size} concurrency=$concurrency" }

        val semaphore = Semaphore(concurrency)
        val succeeded = AtomicInteger(0)
        val failed = AtomicInteger(0)
        val errors = java.util.concurrent.CopyOnWriteArrayList<ImportError>()

        coroutineScope {
            inputs.mapIndexed { i, input ->
                async {
                    semaphore.withPermit {
                        try {
                            val task = taskRepo.create(CreateTaskParams(
                                workspaceId = workspaceId,
                                projectId = projectId,
                                title = input.title,
                                description = input.description,
                                priority = input.priority,
                                assignedTo = input.assignedTo,
                                dueDate = input.dueDate,
                                metadata = input.metadata
                            ))

                            notifRepo?.let {
                                try {
                                    it.create(CreateNotificationParams(
                                        workspaceId = workspaceId,
                                        eventType = "task.imported",
                                        payload = """{"task_id":"${task.id}","title":"${task.title}"}"""
                                    ))
                                } catch (_: Exception) { }
                            }

                            succeeded.incrementAndGet()
                        } catch (e: Exception) {
                            failed.incrementAndGet()
                            errors.add(ImportError(index = i, error = e.message ?: "unknown error"))
                        }
                    }
                }
            }.awaitAll()
        }

        logger.info { "bulk import completed total=${inputs.size} succeeded=${succeeded.get()} failed=${failed.get()}" }

        return ImportResult(
            total = inputs.size,
            succeeded = succeeded.get(),
            failed = failed.get(),
            errors = errors.toList()
        )
    }
}
