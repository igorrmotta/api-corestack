package com.corestack.repository

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import org.jdbi.v3.core.Jdbi
import java.time.Instant
import java.util.UUID

data class Comment(
    val id: UUID,
    val taskId: UUID,
    val authorId: String,
    val content: String,
    val createdAt: Instant,
    val updatedAt: Instant
)

data class CreateCommentParams(val taskId: UUID, val authorId: String, val content: String)
data class ListCommentsParams(val taskId: UUID, val pageSize: Int = 0, val pageToken: String = "")
data class CommentList(val comments: List<Comment>, val nextPageToken: String, val totalCount: Int)

class CommentRepo(private val jdbi: Jdbi) {

    suspend fun create(params: CreateCommentParams): Comment = withContext(Dispatchers.IO) {
        jdbi.withHandle<Comment, Exception> { handle ->
            handle.createQuery(
                """INSERT INTO task_comments (id, task_id, author_id, content, created_at, updated_at)
                   VALUES (gen_random_uuid(), :taskId, :authorId, :content, NOW(), NOW())
                   RETURNING id, task_id, author_id, content, created_at, updated_at"""
            )
                .bind("taskId", params.taskId)
                .bind("authorId", params.authorId)
                .bind("content", params.content)
                .map { rs, _ ->
                    Comment(
                        id = rs.getObject("id", UUID::class.java),
                        taskId = rs.getObject("task_id", UUID::class.java),
                        authorId = rs.getString("author_id"),
                        content = rs.getString("content"),
                        createdAt = rs.getTimestamp("created_at").toInstant(),
                        updatedAt = rs.getTimestamp("updated_at").toInstant()
                    )
                }
                .first()
        }
    }

    suspend fun list(params: ListCommentsParams): CommentList = withContext(Dispatchers.IO) {
        val pageSize = if (params.pageSize in 1..100) params.pageSize else 20

        jdbi.withHandle<CommentList, Exception> { handle ->
            val totalCount = handle.createQuery("SELECT COUNT(*)::int FROM task_comments WHERE task_id = :taskId")
                .bind("taskId", params.taskId)
                .mapTo(Int::class.java)
                .first()

            val query = if (params.pageToken.isNotEmpty()) {
                val cursorId = UUID.fromString(params.pageToken)
                handle.createQuery(
                    """SELECT id, task_id, author_id, content, created_at, updated_at
                       FROM task_comments WHERE task_id = :taskId AND id < :cursorId
                       ORDER BY created_at DESC, id DESC LIMIT :limit"""
                )
                    .bind("taskId", params.taskId)
                    .bind("cursorId", cursorId)
                    .bind("limit", pageSize + 1)
            } else {
                handle.createQuery(
                    """SELECT id, task_id, author_id, content, created_at, updated_at
                       FROM task_comments WHERE task_id = :taskId
                       ORDER BY created_at DESC, id DESC LIMIT :limit"""
                )
                    .bind("taskId", params.taskId)
                    .bind("limit", pageSize + 1)
            }

            val comments = query.map { rs, _ ->
                Comment(
                    id = rs.getObject("id", UUID::class.java),
                    taskId = rs.getObject("task_id", UUID::class.java),
                    authorId = rs.getString("author_id"),
                    content = rs.getString("content"),
                    createdAt = rs.getTimestamp("created_at").toInstant(),
                    updatedAt = rs.getTimestamp("updated_at").toInstant()
                )
            }.toMutableList()

            val nextPageToken = if (comments.size > pageSize) {
                val token = comments[pageSize].id.toString()
                comments.removeAt(pageSize)
                token
            } else ""

            CommentList(comments = comments, nextPageToken = nextPageToken, totalCount = totalCount)
        }
    }

    suspend fun delete(id: UUID): Unit = withContext(Dispatchers.IO) {
        val rowsAffected = jdbi.withHandle<Int, Exception> { handle ->
            handle.createUpdate("DELETE FROM task_comments WHERE id = :id")
                .bind("id", id)
                .execute()
        }
        if (rowsAffected == 0) throw NotFoundError("comment not found")
    }
}
