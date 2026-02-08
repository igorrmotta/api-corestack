package com.corestack.repository

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import org.jdbi.v3.core.Jdbi
import java.time.Instant
import java.util.UUID

data class Workspace(
    val id: UUID,
    val name: String,
    val slug: String,
    val createdAt: Instant,
    val updatedAt: Instant,
    val deletedAt: Instant? = null
)

data class CreateWorkspaceParams(val name: String, val slug: String)
data class UpdateWorkspaceParams(val id: UUID, val name: String, val slug: String)
data class ListWorkspacesParams(val pageSize: Int = 0, val pageToken: String = "")
data class WorkspaceList(val workspaces: List<Workspace>, val nextPageToken: String, val totalCount: Int)

class WorkspaceRepo(private val jdbi: Jdbi) {

    suspend fun create(params: CreateWorkspaceParams): Workspace = withContext(Dispatchers.IO) {
        try {
            jdbi.withHandle<Workspace, Exception> { handle ->
                handle.createQuery(
                    """INSERT INTO workspaces (id, name, slug, created_at, updated_at)
                       VALUES (gen_random_uuid(), :name, :slug, NOW(), NOW())
                       RETURNING id, name, slug, created_at, updated_at, deleted_at"""
                )
                    .bind("name", params.name)
                    .bind("slug", params.slug)
                    .map { rs, _ ->
                        Workspace(
                            id = rs.getObject("id", UUID::class.java),
                            name = rs.getString("name"),
                            slug = rs.getString("slug"),
                            createdAt = rs.getTimestamp("created_at").toInstant(),
                            updatedAt = rs.getTimestamp("updated_at").toInstant(),
                            deletedAt = rs.getTimestamp("deleted_at")?.toInstant()
                        )
                    }
                    .first()
            }
        } catch (e: Exception) {
            if (isPgUniqueViolation(e)) throw ConflictError("create workspace: conflict")
            throw e
        }
    }

    suspend fun getById(id: UUID): Workspace = withContext(Dispatchers.IO) {
        jdbi.withHandle<Workspace?, Exception> { handle ->
            handle.createQuery(
                """SELECT id, name, slug, created_at, updated_at, deleted_at
                   FROM workspaces WHERE id = :id AND deleted_at IS NULL"""
            )
                .bind("id", id)
                .map { rs, _ ->
                    Workspace(
                        id = rs.getObject("id", UUID::class.java),
                        name = rs.getString("name"),
                        slug = rs.getString("slug"),
                        createdAt = rs.getTimestamp("created_at").toInstant(),
                        updatedAt = rs.getTimestamp("updated_at").toInstant(),
                        deletedAt = rs.getTimestamp("deleted_at")?.toInstant()
                    )
                }
                .firstOrNull()
        } ?: throw NotFoundError("workspace not found")
    }

    suspend fun list(params: ListWorkspacesParams): WorkspaceList = withContext(Dispatchers.IO) {
        val pageSize = if (params.pageSize in 1..100) params.pageSize else 20

        jdbi.withHandle<WorkspaceList, Exception> { handle ->
            val totalCount = handle.createQuery("SELECT COUNT(*)::int FROM workspaces WHERE deleted_at IS NULL")
                .mapTo(Int::class.java)
                .first()

            val query = if (params.pageToken.isNotEmpty()) {
                val cursorId = UUID.fromString(params.pageToken)
                handle.createQuery(
                    """SELECT id, name, slug, created_at, updated_at, deleted_at
                       FROM workspaces WHERE deleted_at IS NULL AND id < :cursorId
                       ORDER BY created_at DESC, id DESC LIMIT :limit"""
                )
                    .bind("cursorId", cursorId)
                    .bind("limit", pageSize + 1)
            } else {
                handle.createQuery(
                    """SELECT id, name, slug, created_at, updated_at, deleted_at
                       FROM workspaces WHERE deleted_at IS NULL
                       ORDER BY created_at DESC, id DESC LIMIT :limit"""
                )
                    .bind("limit", pageSize + 1)
            }

            val workspaces = query.map { rs, _ ->
                Workspace(
                    id = rs.getObject("id", UUID::class.java),
                    name = rs.getString("name"),
                    slug = rs.getString("slug"),
                    createdAt = rs.getTimestamp("created_at").toInstant(),
                    updatedAt = rs.getTimestamp("updated_at").toInstant(),
                    deletedAt = rs.getTimestamp("deleted_at")?.toInstant()
                )
            }.toMutableList()

            val nextPageToken = if (workspaces.size > pageSize) {
                val token = workspaces[pageSize].id.toString()
                workspaces.removeAt(pageSize)
                token
            } else ""

            WorkspaceList(workspaces = workspaces, nextPageToken = nextPageToken, totalCount = totalCount)
        }
    }

    suspend fun update(params: UpdateWorkspaceParams): Workspace = withContext(Dispatchers.IO) {
        try {
            jdbi.withHandle<Workspace?, Exception> { handle ->
                handle.createQuery(
                    """UPDATE workspaces SET name = :name, slug = :slug, updated_at = NOW()
                       WHERE id = :id AND deleted_at IS NULL
                       RETURNING id, name, slug, created_at, updated_at, deleted_at"""
                )
                    .bind("name", params.name)
                    .bind("slug", params.slug)
                    .bind("id", params.id)
                    .map { rs, _ ->
                        Workspace(
                            id = rs.getObject("id", UUID::class.java),
                            name = rs.getString("name"),
                            slug = rs.getString("slug"),
                            createdAt = rs.getTimestamp("created_at").toInstant(),
                            updatedAt = rs.getTimestamp("updated_at").toInstant(),
                            deletedAt = rs.getTimestamp("deleted_at")?.toInstant()
                        )
                    }
                    .firstOrNull()
            } ?: throw NotFoundError("workspace not found")
        } catch (e: RepositoryError) {
            throw e
        } catch (e: Exception) {
            if (isPgUniqueViolation(e)) throw ConflictError("update workspace: conflict")
            throw e
        }
    }

    suspend fun delete(id: UUID): Unit = withContext(Dispatchers.IO) {
        val rowsAffected = jdbi.withHandle<Int, Exception> { handle ->
            handle.createUpdate(
                """UPDATE workspaces SET deleted_at = NOW(), updated_at = NOW()
                   WHERE id = :id AND deleted_at IS NULL"""
            )
                .bind("id", id)
                .execute()
        }
        if (rowsAffected == 0) throw NotFoundError("workspace not found")
    }
}
