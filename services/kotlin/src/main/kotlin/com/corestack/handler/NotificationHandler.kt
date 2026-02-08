package com.corestack.handler

import com.corestack.connect.ConnectAdapter
import com.corestack.repository.*
import com.google.protobuf.Struct
import com.google.protobuf.util.JsonFormat
import notification.v1.NotificationOuterClass as NF
import common.v1.Pagination

private val jsonParser = JsonFormat.parser().ignoringUnknownFields()

fun registerNotificationHandlers(adapter: ConnectAdapter, repo: NotificationRepo) {
    adapter.registerUnary(
        "notification.v1.NotificationService", "ListNotifications",
        NF.ListNotificationsRequest.getDefaultInstance()
    ) { req: NF.ListNotificationsRequest ->
        try {
            val workspaceId = parseUUID(req.workspaceId)
            val params = ListNotificationsParams(
                workspaceId = workspaceId,
                status = req.status,
                pageSize = if (req.hasPagination()) req.pagination.pageSize else 0,
                pageToken = if (req.hasPagination()) req.pagination.pageToken else ""
            )
            val list = repo.list(params)
            NF.ListNotificationsResponse.newBuilder()
                .addAllNotifications(list.notifications.map { it.toProto() })
                .setPagination(Pagination.PaginationResponse.newBuilder()
                    .setNextPageToken(list.nextPageToken)
                    .setTotalCount(list.totalCount)
                    .build())
                .build()
        } catch (e: RepositoryError) { throw toConnectError(e) }
    }

    adapter.registerUnary(
        "notification.v1.NotificationService", "MarkNotificationRead",
        NF.MarkNotificationReadRequest.getDefaultInstance()
    ) { req: NF.MarkNotificationReadRequest ->
        try {
            val n = repo.markProcessed(req.id)
            NF.MarkNotificationReadResponse.newBuilder()
                .setNotification(n.toProto())
                .build()
        } catch (e: RepositoryError) { throw toConnectError(e) }
    }
}

private fun Notification.toProto(): NF.Notification {
    val builder = NF.Notification.newBuilder()
        .setId(id)
        .setWorkspaceId(workspaceId.toString())
        .setEventType(eventType)
        .setStatus(status)
        .setRetryCount(retryCount)
        .setCreatedAt(createdAt.toTimestamp())

    processedAt?.let { builder.setProcessedAt(it.toTimestamp()) }

    if (payload.isNotEmpty() && payload != "{}") {
        val structBuilder = Struct.newBuilder()
        jsonParser.merge(payload, structBuilder)
        builder.setPayload(structBuilder.build())
    }

    return builder.build()
}
