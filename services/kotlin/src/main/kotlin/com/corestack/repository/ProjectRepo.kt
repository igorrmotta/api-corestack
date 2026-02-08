package com.corestack.repository

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import org.jdbi.v3.core.Jdbi
import java.time.Instant
import java.util.UUID

data class Project(
    val id: UUID,
    val workspaceId: UUID,
    val name: String,
    val description: String,
    val status: String,
    val createdAt: Instant,
    val updatedAt: Instant,
    val deletedAt: Instant? = null
)

data class CreateProjectParams(val workspaceId: UUID, val name: String, val description: String)
data class UpdateProjectParams(val id: UUID, val name: String, val description: String, val status: String)
data class ListProjectsParams(val workspaceId: UUID, val pageSize: Int = 0, val pageToken: String = "")
data class ProjectList(val projects: List<Project>, val nextPageToken: String, val totalCount: Int)

class ProjectRepo(private val jdbi: Jdbi) {

    suspend fun create(params: CreateProjectParams): Project = withContext(Dispatchers.IO) {
        jdbi.withHandle<Project, Exception> { handle ->
            handle.createQuery(
                """INSERT INTO projects (id, workspace_id, name, description, status, created_at, updated_at)
                   VALUES (gen_random_uuid(), :workspaceId, :name, :description, 'active', NOW(), NOW())
                   RETURNING id, workspace_id, name, description, status, created_at, updated_at, deleted_at"""
            )
                .bind("workspaceId", params.workspaceId)
                .bind("name", params.name)
                .bind("description", params.description)
                .map { rs, _ ->
                    Project(
                        id = rs.getObject("id", UUID::class.java),
                        workspaceId = rs.getObject("workspace_id", UUID::class.java),
                        name = rs.getString("name"),
                        description = rs.getString("description"),
                        status = rs.getString("status"),
                        createdAt = rs.getTimestamp("created_at").toInstant(),
                        updatedAt = rs.getTimestamp("updated_at").toInstant(),
                        deletedAt = rs.getTimestamp("deleted_at")?.toInstant()
                    )
                }
                .first()
        }
    }

    suspend fun getById(id: UUID): Project = withContext(Dispatchers.IO) {
        jdbi.withHandle<Project?, Exception> { handle ->
            handle.createQuery(
                """SELECT id, workspace_id, name, description, status, created_at, updated_at, deleted_at
                   FROM projects WHERE id = :id AND deleted_at IS NULL"""
            )
                .bind("id", id)
                .map { rs, _ ->
                    Project(
                        id = rs.getObject("id", UUID::class.java),
                        workspaceId = rs.getObject("workspace_id", UUID::class.java),
                        name = rs.getString("name"),
                        description = rs.getString("description"),
                        status = rs.getString("status"),
                        createdAt = rs.getTimestamp("created_at").toInstant(),
                        updatedAt = rs.getTimestamp("updated_at").toInstant(),
                        deletedAt = rs.getTimestamp("deleted_at")?.toInstant()
                    )
                }
                .firstOrNull()
        } ?: throw NotFoundError("project not found")
    }

    suspend fun list(params: ListProjectsParams): ProjectList = withContext(Dispatchers.IO) {
        val pageSize = if (params.pageSize in 1..100) params.pageSize else 20

        jdbi.withHandle<ProjectList, Exception> { handle ->
            val totalCount = handle.createQuery(
                "SELECT COUNT(*)::int FROM projects WHERE workspace_id = :workspaceId AND deleted_at IS NULL"
            )
                .bind("workspaceId", params.workspaceId)
                .mapTo(Int::class.java)
                .first()

            val query = if (params.pageToken.isNotEmpty()) {
                val cursorId = UUID.fromString(params.pageToken)
                handle.createQuery(
                    """SELECT id, workspace_id, name, description, status, created_at, updated_at, deleted_at
                       FROM projects WHERE workspace_id = :workspaceId AND deleted_at IS NULL AND id < :cursorId
                       ORDER BY created_at DESC, id DESC LIMIT :limit"""
                )
                    .bind("workspaceId", params.workspaceId)
                    .bind("cursorId", cursorId)
                    .bind("limit", pageSize + 1)
            } else {
                handle.createQuery(
                    """SELECT id, workspace_id, name, description, status, created_at, updated_at, deleted_at
                       FROM projects WHERE workspace_id = :workspaceId AND deleted_at IS NULL
                       ORDER BY created_at DESC, id DESC LIMIT :limit"""
                )
                    .bind("workspaceId", params.workspaceId)
                    .bind("limit", pageSize + 1)
            }

            val projects = query.map { rs, _ ->
                Project(
                    id = rs.getObject("id", UUID::class.java),
                    workspaceId = rs.getObject("workspace_id", UUID::class.java),
                    name = rs.getString("name"),
                    description = rs.getString("description"),
                    status = rs.getString("status"),
                    createdAt = rs.getTimestamp("created_at").toInstant(),
                    updatedAt = rs.getTimestamp("updated_at").toInstant(),
                    deletedAt = rs.getTimestamp("deleted_at")?.toInstant()
                )
            }.toMutableList()

            val nextPageToken = if (projects.size > pageSize) {
                val token = projects[pageSize].id.toString()
                projects.removeAt(pageSize)
                token
            } else ""

            ProjectList(projects = projects, nextPageToken = nextPageToken, totalCount = totalCount)
        }
    }

    suspend fun update(params: UpdateProjectParams): Project = withContext(Dispatchers.IO) {
        jdbi.withHandle<Project?, Exception> { handle ->
            handle.createQuery(
                """UPDATE projects SET name = :name, description = :description, status = :status, updated_at = NOW()
                   WHERE id = :id AND deleted_at IS NULL
                   RETURNING id, workspace_id, name, description, status, created_at, updated_at, deleted_at"""
            )
                .bind("name", params.name)
                .bind("description", params.description)
                .bind("status", params.status)
                .bind("id", params.id)
                .map { rs, _ ->
                    Project(
                        id = rs.getObject("id", UUID::class.java),
                        workspaceId = rs.getObject("workspace_id", UUID::class.java),
                        name = rs.getString("name"),
                        description = rs.getString("description"),
                        status = rs.getString("status"),
                        createdAt = rs.getTimestamp("created_at").toInstant(),
                        updatedAt = rs.getTimestamp("updated_at").toInstant(),
                        deletedAt = rs.getTimestamp("deleted_at")?.toInstant()
                    )
                }
                .firstOrNull()
        } ?: throw NotFoundError("project not found")
    }

    suspend fun delete(id: UUID): Unit = withContext(Dispatchers.IO) {
        val rowsAffected = jdbi.withHandle<Int, Exception> { handle ->
            handle.createUpdate(
                """UPDATE projects SET deleted_at = NOW(), updated_at = NOW()
                   WHERE id = :id AND deleted_at IS NULL"""
            )
                .bind("id", id)
                .execute()
        }
        if (rowsAffected == 0) throw NotFoundError("project not found")
    }
}
