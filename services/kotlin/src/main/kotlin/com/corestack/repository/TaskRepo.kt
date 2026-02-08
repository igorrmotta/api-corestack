package com.corestack.repository

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import org.jdbi.v3.core.Jdbi
import org.postgresql.util.PGobject
import java.time.Instant
import java.util.UUID

data class Task(
    val id: UUID,
    val workspaceId: UUID,
    val projectId: UUID,
    val title: String,
    val description: String,
    val status: String,
    val priority: String,
    val assignedTo: String,
    val dueDate: Instant?,
    val metadata: String,
    val createdAt: Instant,
    val updatedAt: Instant,
    val deletedAt: Instant? = null
)

data class CreateTaskParams(
    val workspaceId: UUID,
    val projectId: UUID,
    val title: String,
    val description: String = "",
    val priority: String = "",
    val assignedTo: String = "",
    val dueDate: Instant? = null,
    val metadata: String? = null
)

data class UpdateTaskParams(
    val id: UUID,
    val title: String,
    val description: String = "",
    val status: String = "",
    val priority: String = "",
    val assignedTo: String = "",
    val dueDate: Instant? = null,
    val metadata: String? = null
)

data class ListTasksParams(
    val workspaceId: UUID,
    val projectId: UUID? = null,
    val status: String = "",
    val priority: String = "",
    val assignedTo: String = "",
    val pageSize: Int = 0,
    val pageToken: String = ""
)

data class TaskList(val tasks: List<Task>, val nextPageToken: String, val totalCount: Int)

class TaskRepo(private val jdbi: Jdbi) {

    suspend fun create(params: CreateTaskParams): Task = withContext(Dispatchers.IO) {
        val assignedTo: String? = params.assignedTo.ifEmpty { null }
        val metadataJson = if (params.metadata.isNullOrEmpty()) "{}" else params.metadata

        jdbi.withHandle<Task, Exception> { handle ->
            handle.createQuery(
                """INSERT INTO tasks (id, workspace_id, project_id, title, description, status, priority, assigned_to, due_date, metadata, created_at, updated_at)
                   VALUES (gen_random_uuid(), :workspaceId, :projectId, :title, :description,
                           COALESCE(NULLIF(:status, ''), 'todo'),
                           COALESCE(NULLIF(:priority, ''), 'medium'),
                           :assignedTo, :dueDate, :metadata, NOW(), NOW())
                   RETURNING id, workspace_id, project_id, title, description, status, priority,
                             COALESCE(assigned_to, ''), due_date, metadata, created_at, updated_at, deleted_at"""
            )
                .bind("workspaceId", params.workspaceId)
                .bind("projectId", params.projectId)
                .bind("title", params.title)
                .bind("description", params.description)
                .bind("status", "")
                .bind("priority", params.priority)
                .bind("assignedTo", assignedTo)
                .bind("dueDate", params.dueDate?.let { java.sql.Timestamp.from(it) })
                .bind("metadata", PGobject().apply { type = "jsonb"; value = metadataJson })
                .map { rs, _ ->
                    Task(
                        id = rs.getObject("id", UUID::class.java),
                        workspaceId = rs.getObject("workspace_id", UUID::class.java),
                        projectId = rs.getObject("project_id", UUID::class.java),
                        title = rs.getString("title"),
                        description = rs.getString("description"),
                        status = rs.getString("status"),
                        priority = rs.getString("priority"),
                        assignedTo = rs.getString("coalesce") ?: "",
                        dueDate = rs.getTimestamp("due_date")?.toInstant(),
                        metadata = rs.getString("metadata") ?: "{}",
                        createdAt = rs.getTimestamp("created_at").toInstant(),
                        updatedAt = rs.getTimestamp("updated_at").toInstant(),
                        deletedAt = rs.getTimestamp("deleted_at")?.toInstant()
                    )
                }
                .first()
        }
    }

    suspend fun getById(id: UUID): Task = withContext(Dispatchers.IO) {
        jdbi.withHandle<Task?, Exception> { handle ->
            handle.createQuery(
                """SELECT id, workspace_id, project_id, title, description, status, priority,
                          COALESCE(assigned_to, '') AS assigned_to, due_date, metadata, created_at, updated_at, deleted_at
                   FROM tasks WHERE id = :id AND deleted_at IS NULL"""
            )
                .bind("id", id)
                .map { rs, _ ->
                    Task(
                        id = rs.getObject("id", UUID::class.java),
                        workspaceId = rs.getObject("workspace_id", UUID::class.java),
                        projectId = rs.getObject("project_id", UUID::class.java),
                        title = rs.getString("title"),
                        description = rs.getString("description"),
                        status = rs.getString("status"),
                        priority = rs.getString("priority"),
                        assignedTo = rs.getString("assigned_to") ?: "",
                        dueDate = rs.getTimestamp("due_date")?.toInstant(),
                        metadata = rs.getString("metadata") ?: "{}",
                        createdAt = rs.getTimestamp("created_at").toInstant(),
                        updatedAt = rs.getTimestamp("updated_at").toInstant(),
                        deletedAt = rs.getTimestamp("deleted_at")?.toInstant()
                    )
                }
                .firstOrNull()
        } ?: throw NotFoundError("task not found")
    }

    suspend fun list(params: ListTasksParams): TaskList = withContext(Dispatchers.IO) {
        val pageSize = if (params.pageSize in 1..100) params.pageSize else 20

        jdbi.withHandle<TaskList, Exception> { handle ->
            // Build dynamic WHERE clause
            val conditions = mutableListOf("workspace_id = :workspaceId", "deleted_at IS NULL")
            val bindings = mutableMapOf<String, Any>("workspaceId" to params.workspaceId)

            if (params.projectId != null) {
                conditions.add("project_id = :projectId")
                bindings["projectId"] = params.projectId
            }
            if (params.status.isNotEmpty()) {
                conditions.add("status = :status")
                bindings["status"] = params.status
            }
            if (params.priority.isNotEmpty()) {
                conditions.add("priority = :priority")
                bindings["priority"] = params.priority
            }
            if (params.assignedTo.isNotEmpty()) {
                conditions.add("assigned_to = :assignedTo")
                bindings["assignedTo"] = params.assignedTo
            }

            val whereClause = conditions.joinToString(" AND ")

            // Count query
            val countQuery = handle.createQuery("SELECT COUNT(*)::int FROM tasks WHERE $whereClause")
            bindings.forEach { (k, v) -> countQuery.bind(k, v) }
            val totalCount = countQuery.mapTo(Int::class.java).first()

            // Add cursor pagination
            if (params.pageToken.isNotEmpty()) {
                val cursorId = UUID.fromString(params.pageToken)
                conditions.add("id < :cursorId")
                bindings["cursorId"] = cursorId
            }

            val paginatedWhere = conditions.joinToString(" AND ")

            val listQuery = handle.createQuery(
                """SELECT id, workspace_id, project_id, title, description, status, priority,
                          COALESCE(assigned_to, '') AS assigned_to, due_date, metadata, created_at, updated_at, deleted_at
                   FROM tasks WHERE $paginatedWhere
                   ORDER BY created_at DESC, id DESC LIMIT :limit"""
            )
            bindings.forEach { (k, v) -> listQuery.bind(k, v) }
            listQuery.bind("limit", pageSize + 1)

            val tasks = listQuery.map { rs, _ ->
                Task(
                    id = rs.getObject("id", UUID::class.java),
                    workspaceId = rs.getObject("workspace_id", UUID::class.java),
                    projectId = rs.getObject("project_id", UUID::class.java),
                    title = rs.getString("title"),
                    description = rs.getString("description"),
                    status = rs.getString("status"),
                    priority = rs.getString("priority"),
                    assignedTo = rs.getString("assigned_to") ?: "",
                    dueDate = rs.getTimestamp("due_date")?.toInstant(),
                    metadata = rs.getString("metadata") ?: "{}",
                    createdAt = rs.getTimestamp("created_at").toInstant(),
                    updatedAt = rs.getTimestamp("updated_at").toInstant(),
                    deletedAt = rs.getTimestamp("deleted_at")?.toInstant()
                )
            }.toMutableList()

            val nextPageToken = if (tasks.size > pageSize) {
                val token = tasks[pageSize].id.toString()
                tasks.removeAt(pageSize)
                token
            } else ""

            TaskList(tasks = tasks, nextPageToken = nextPageToken, totalCount = totalCount)
        }
    }

    suspend fun update(params: UpdateTaskParams): Task = withContext(Dispatchers.IO) {
        val assignedTo: String? = params.assignedTo.ifEmpty { null }
        val metadataJson = if (params.metadata.isNullOrEmpty()) "{}" else params.metadata

        jdbi.withHandle<Task?, Exception> { handle ->
            handle.createQuery(
                """UPDATE tasks SET title = :title, description = :description, status = :status, priority = :priority,
                          assigned_to = :assignedTo, due_date = :dueDate, metadata = :metadata, updated_at = NOW()
                   WHERE id = :id AND deleted_at IS NULL
                   RETURNING id, workspace_id, project_id, title, description, status, priority,
                             COALESCE(assigned_to, '') AS assigned_to, due_date, metadata, created_at, updated_at, deleted_at"""
            )
                .bind("title", params.title)
                .bind("description", params.description)
                .bind("status", params.status)
                .bind("priority", params.priority)
                .bind("assignedTo", assignedTo)
                .bind("dueDate", params.dueDate?.let { java.sql.Timestamp.from(it) })
                .bind("metadata", PGobject().apply { type = "jsonb"; value = metadataJson })
                .bind("id", params.id)
                .map { rs, _ ->
                    Task(
                        id = rs.getObject("id", UUID::class.java),
                        workspaceId = rs.getObject("workspace_id", UUID::class.java),
                        projectId = rs.getObject("project_id", UUID::class.java),
                        title = rs.getString("title"),
                        description = rs.getString("description"),
                        status = rs.getString("status"),
                        priority = rs.getString("priority"),
                        assignedTo = rs.getString("assigned_to") ?: "",
                        dueDate = rs.getTimestamp("due_date")?.toInstant(),
                        metadata = rs.getString("metadata") ?: "{}",
                        createdAt = rs.getTimestamp("created_at").toInstant(),
                        updatedAt = rs.getTimestamp("updated_at").toInstant(),
                        deletedAt = rs.getTimestamp("deleted_at")?.toInstant()
                    )
                }
                .firstOrNull()
        } ?: throw NotFoundError("task not found")
    }

    suspend fun delete(id: UUID): Unit = withContext(Dispatchers.IO) {
        val rowsAffected = jdbi.withHandle<Int, Exception> { handle ->
            handle.createUpdate(
                """UPDATE tasks SET deleted_at = NOW(), updated_at = NOW()
                   WHERE id = :id AND deleted_at IS NULL"""
            )
                .bind("id", id)
                .execute()
        }
        if (rowsAffected == 0) throw NotFoundError("task not found")
    }
}
