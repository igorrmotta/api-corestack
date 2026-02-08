package com.corestack.repository

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import org.jdbi.v3.core.Jdbi
import org.postgresql.util.PGobject
import java.time.Instant
import java.util.UUID

data class Notification(
    val id: Long,
    val workspaceId: UUID,
    val eventType: String,
    val payload: String,
    val status: String,
    val retryCount: Int,
    val maxRetries: Int,
    val nextRetryAt: Instant?,
    val lastError: String,
    val createdAt: Instant,
    val processedAt: Instant?
)

data class CreateNotificationParams(val workspaceId: UUID, val eventType: String, val payload: String)
data class ListNotificationsParams(
    val workspaceId: UUID,
    val status: String = "",
    val pageSize: Int = 0,
    val pageToken: String = ""
)
data class NotificationList(val notifications: List<Notification>, val nextPageToken: String, val totalCount: Int)

class NotificationRepo(private val jdbi: Jdbi) {

    suspend fun create(params: CreateNotificationParams): Notification = withContext(Dispatchers.IO) {
        jdbi.withHandle<Notification, Exception> { handle ->
            handle.createQuery(
                """INSERT INTO notification_queue (workspace_id, event_type, payload, status, created_at)
                   VALUES (:workspaceId, :eventType, :payload, 'pending', NOW())
                   RETURNING id, workspace_id, event_type, payload, status, retry_count, max_retries,
                             next_retry_at, COALESCE(last_error, '') AS last_error, created_at, processed_at"""
            )
                .bind("workspaceId", params.workspaceId)
                .bind("eventType", params.eventType)
                .bind("payload", PGobject().apply { type = "jsonb"; value = params.payload })
                .map { rs, _ ->
                    Notification(
                        id = rs.getLong("id"),
                        workspaceId = rs.getObject("workspace_id", UUID::class.java),
                        eventType = rs.getString("event_type"),
                        payload = rs.getString("payload") ?: "{}",
                        status = rs.getString("status"),
                        retryCount = rs.getInt("retry_count"),
                        maxRetries = rs.getInt("max_retries"),
                        nextRetryAt = rs.getTimestamp("next_retry_at")?.toInstant(),
                        lastError = rs.getString("last_error") ?: "",
                        createdAt = rs.getTimestamp("created_at").toInstant(),
                        processedAt = rs.getTimestamp("processed_at")?.toInstant()
                    )
                }
                .first()
        }
    }

    suspend fun list(params: ListNotificationsParams): NotificationList = withContext(Dispatchers.IO) {
        val pageSize = if (params.pageSize in 1..100) params.pageSize else 20

        // Parse offset from page token (offset-based pagination for notifications)
        val offset = if (params.pageToken.isNotEmpty()) params.pageToken.toInt() else 0

        jdbi.withHandle<NotificationList, Exception> { handle ->
            // Count query
            val totalCount = if (params.status.isNotEmpty()) {
                handle.createQuery(
                    "SELECT COUNT(*)::int FROM notification_queue WHERE workspace_id = :workspaceId AND status = :status"
                )
                    .bind("workspaceId", params.workspaceId)
                    .bind("status", params.status)
                    .mapTo(Int::class.java)
                    .first()
            } else {
                handle.createQuery(
                    "SELECT COUNT(*)::int FROM notification_queue WHERE workspace_id = :workspaceId"
                )
                    .bind("workspaceId", params.workspaceId)
                    .mapTo(Int::class.java)
                    .first()
            }

            // List query
            val query = if (params.status.isNotEmpty()) {
                handle.createQuery(
                    """SELECT id, workspace_id, event_type, payload, status, retry_count, max_retries,
                              next_retry_at, COALESCE(last_error, '') AS last_error, created_at, processed_at
                       FROM notification_queue
                       WHERE workspace_id = :workspaceId AND status = :status
                       ORDER BY created_at DESC
                       LIMIT :limit OFFSET :offset"""
                )
                    .bind("workspaceId", params.workspaceId)
                    .bind("status", params.status)
                    .bind("limit", pageSize)
                    .bind("offset", offset)
            } else {
                handle.createQuery(
                    """SELECT id, workspace_id, event_type, payload, status, retry_count, max_retries,
                              next_retry_at, COALESCE(last_error, '') AS last_error, created_at, processed_at
                       FROM notification_queue
                       WHERE workspace_id = :workspaceId
                       ORDER BY created_at DESC
                       LIMIT :limit OFFSET :offset"""
                )
                    .bind("workspaceId", params.workspaceId)
                    .bind("limit", pageSize)
                    .bind("offset", offset)
            }

            val notifications = query.map { rs, _ ->
                Notification(
                    id = rs.getLong("id"),
                    workspaceId = rs.getObject("workspace_id", UUID::class.java),
                    eventType = rs.getString("event_type"),
                    payload = rs.getString("payload") ?: "{}",
                    status = rs.getString("status"),
                    retryCount = rs.getInt("retry_count"),
                    maxRetries = rs.getInt("max_retries"),
                    nextRetryAt = rs.getTimestamp("next_retry_at")?.toInstant(),
                    lastError = rs.getString("last_error") ?: "",
                    createdAt = rs.getTimestamp("created_at").toInstant(),
                    processedAt = rs.getTimestamp("processed_at")?.toInstant()
                )
            }.list()

            val nextPageToken = run {
                val nextOffset = offset + pageSize
                if (nextOffset < totalCount) nextOffset.toString() else ""
            }

            NotificationList(notifications = notifications, nextPageToken = nextPageToken, totalCount = totalCount)
        }
    }

    suspend fun fetchPending(limit: Int): List<Notification> = withContext(Dispatchers.IO) {
        jdbi.withHandle<List<Notification>, Exception> { handle ->
            handle.createQuery(
                """SELECT id, workspace_id, event_type, payload, status, retry_count, max_retries,
                          next_retry_at, COALESCE(last_error, '') AS last_error, created_at, processed_at
                   FROM notification_queue
                   WHERE status IN ('pending', 'failed')
                     AND (next_retry_at IS NULL OR next_retry_at <= NOW())
                   ORDER BY created_at ASC
                   LIMIT :limit
                   FOR UPDATE SKIP LOCKED"""
            )
                .bind("limit", limit)
                .map { rs, _ ->
                    Notification(
                        id = rs.getLong("id"),
                        workspaceId = rs.getObject("workspace_id", UUID::class.java),
                        eventType = rs.getString("event_type"),
                        payload = rs.getString("payload") ?: "{}",
                        status = rs.getString("status"),
                        retryCount = rs.getInt("retry_count"),
                        maxRetries = rs.getInt("max_retries"),
                        nextRetryAt = rs.getTimestamp("next_retry_at")?.toInstant(),
                        lastError = rs.getString("last_error") ?: "",
                        createdAt = rs.getTimestamp("created_at").toInstant(),
                        processedAt = rs.getTimestamp("processed_at")?.toInstant()
                    )
                }
                .list()
        }
    }

    suspend fun markProcessed(id: Long): Notification = withContext(Dispatchers.IO) {
        jdbi.withHandle<Notification?, Exception> { handle ->
            handle.createQuery(
                """UPDATE notification_queue
                   SET status = 'processed', processed_at = NOW()
                   WHERE id = :id
                   RETURNING id, workspace_id, event_type, payload, status, retry_count, max_retries,
                             next_retry_at, COALESCE(last_error, '') AS last_error, created_at, processed_at"""
            )
                .bind("id", id)
                .map { rs, _ ->
                    Notification(
                        id = rs.getLong("id"),
                        workspaceId = rs.getObject("workspace_id", UUID::class.java),
                        eventType = rs.getString("event_type"),
                        payload = rs.getString("payload") ?: "{}",
                        status = rs.getString("status"),
                        retryCount = rs.getInt("retry_count"),
                        maxRetries = rs.getInt("max_retries"),
                        nextRetryAt = rs.getTimestamp("next_retry_at")?.toInstant(),
                        lastError = rs.getString("last_error") ?: "",
                        createdAt = rs.getTimestamp("created_at").toInstant(),
                        processedAt = rs.getTimestamp("processed_at")?.toInstant()
                    )
                }
                .firstOrNull()
        } ?: throw NotFoundError("notification not found")
    }

    suspend fun markFailed(id: Long, errMsg: String): Unit = withContext(Dispatchers.IO) {
        val rowsAffected = jdbi.withHandle<Int, Exception> { handle ->
            handle.createUpdate(
                """UPDATE notification_queue
                   SET status = 'failed',
                       last_error = :errMsg,
                       retry_count = retry_count + 1,
                       next_retry_at = NOW() + (INTERVAL '1 second' * POWER(2, retry_count + 1))
                   WHERE id = :id"""
            )
                .bind("errMsg", errMsg)
                .bind("id", id)
                .execute()
        }
        if (rowsAffected == 0) throw NotFoundError("notification not found")
    }
}
