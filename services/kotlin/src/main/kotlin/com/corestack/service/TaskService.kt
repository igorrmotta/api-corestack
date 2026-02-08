package com.corestack.service

import com.corestack.repository.*
import io.github.oshai.kotlinlogging.KotlinLogging
import java.util.UUID

private val logger = KotlinLogging.logger {}

class TaskService(
    private val repo: TaskRepo,
    private val notifRepo: NotificationRepo?
) {

    suspend fun create(params: CreateTaskParams): Task {
        if (params.title.isBlank()) throw InvalidInputError("invalid input: title is required")
        if (params.workspaceId == UUID(0, 0)) throw InvalidInputError("invalid input: workspace_id is required")
        if (params.projectId == UUID(0, 0)) throw InvalidInputError("invalid input: project_id is required")

        val validPriorities = setOf("low", "medium", "high", "critical", "")
        if (params.priority !in validPriorities) {
            throw InvalidInputError("invalid input: invalid priority: ${params.priority}")
        }

        logger.debug { "creating task title=${params.title} project_id=${params.projectId}" }
        val task = repo.create(params)

        // Fire-and-forget notification
        notifRepo?.let {
            try {
                it.create(CreateNotificationParams(
                    workspaceId = task.workspaceId,
                    eventType = "task.created",
                    payload = """{"task_id":"${task.id}","title":"${task.title}"}"""
                ))
            } catch (_: Exception) { }
        }

        return task
    }

    suspend fun getById(id: UUID): Task = repo.getById(id)

    suspend fun list(params: ListTasksParams): TaskList {
        val validStatuses = setOf("todo", "in_progress", "review", "done", "")
        if (params.status !in validStatuses) {
            throw InvalidInputError("invalid input: invalid status filter: ${params.status}")
        }
        return repo.list(params)
    }

    suspend fun update(params: UpdateTaskParams): Task {
        if (params.title.isBlank()) throw InvalidInputError("invalid input: title is required")

        val validStatuses = setOf("todo", "in_progress", "review", "done", "")
        if (params.status !in validStatuses) {
            throw InvalidInputError("invalid input: invalid status: ${params.status}")
        }

        return repo.update(params)
    }

    suspend fun delete(id: UUID) = repo.delete(id)
}
