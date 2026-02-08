package com.corestack.worker

import com.corestack.repository.NotificationRepo
import io.github.oshai.kotlinlogging.KotlinLogging
import kotlinx.coroutines.delay
import kotlinx.coroutines.isActive
import kotlin.coroutines.coroutineContext

private val logger = KotlinLogging.logger {}

class NotificationProcessor(
    private val repo: NotificationRepo,
    private val batchSize: Int = 10
) {
    @Volatile
    private var running = true

    fun stop() {
        running = false
    }

    suspend fun run() {
        logger.info { "notification processor started batch_size=$batchSize" }
        while (running && coroutineContext.isActive) {
            try {
                val notifications = repo.fetchPending(batchSize)
                if (notifications.isEmpty()) {
                    delay(1000)
                    continue
                }

                logger.info { "processing notification batch count=${notifications.size}" }

                for (n in notifications) {
                    try {
                        logger.info { "sending notification id=${n.id} event_type=${n.eventType} workspace_id=${n.workspaceId}" }

                        // Simulate sending
                        delay(50)

                        repo.markProcessed(n.id)
                        logger.info { "notification sent id=${n.id}" }
                    } catch (e: Exception) {
                        logger.error { "failed to process notification id=${n.id} error=${e.message}" }
                        try {
                            repo.markFailed(n.id, e.message ?: "unknown error")
                        } catch (markErr: Exception) {
                            logger.error { "failed to mark notification failed id=${n.id} error=${markErr.message}" }
                        }
                    }
                }
            } catch (e: Exception) {
                logger.error { "notification processor error: ${e.message}" }
                delay(1000)
            }
        }
        logger.info { "notification processor stopped" }
    }
}
