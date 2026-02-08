package com.corestack.handler

import com.corestack.connect.ConnectAdapter
import com.corestack.connect.ConnectCode
import com.corestack.connect.ConnectError
import com.corestack.repository.*
import com.corestack.service.WorkspaceService
import com.google.protobuf.Timestamp
import workspace.v1.WorkspaceOuterClass as WS
import common.v1.Pagination
import java.util.UUID

fun registerWorkspaceHandlers(adapter: ConnectAdapter, svc: WorkspaceService) {
    adapter.registerUnary(
        "workspace.v1.WorkspaceService", "CreateWorkspace",
        WS.CreateWorkspaceRequest.getDefaultInstance()
    ) { req: WS.CreateWorkspaceRequest ->
        try {
            val w = svc.create(CreateWorkspaceParams(name = req.name, slug = req.slug))
            WS.CreateWorkspaceResponse.newBuilder().setWorkspace(w.toProto()).build()
        } catch (e: RepositoryError) { throw toConnectError(e) }
    }

    adapter.registerUnary(
        "workspace.v1.WorkspaceService", "GetWorkspace",
        WS.GetWorkspaceRequest.getDefaultInstance()
    ) { req: WS.GetWorkspaceRequest ->
        try {
            val id = parseUUID(req.id)
            val w = svc.getById(id)
            WS.GetWorkspaceResponse.newBuilder().setWorkspace(w.toProto()).build()
        } catch (e: RepositoryError) { throw toConnectError(e) }
    }

    adapter.registerUnary(
        "workspace.v1.WorkspaceService", "ListWorkspaces",
        WS.ListWorkspacesRequest.getDefaultInstance()
    ) { req: WS.ListWorkspacesRequest ->
        try {
            val params = ListWorkspacesParams(
                pageSize = if (req.hasPagination()) req.pagination.pageSize else 0,
                pageToken = if (req.hasPagination()) req.pagination.pageToken else ""
            )
            val list = svc.list(params)
            WS.ListWorkspacesResponse.newBuilder()
                .addAllWorkspaces(list.workspaces.map { it.toProto() })
                .setPagination(Pagination.PaginationResponse.newBuilder()
                    .setNextPageToken(list.nextPageToken)
                    .setTotalCount(list.totalCount)
                    .build())
                .build()
        } catch (e: RepositoryError) { throw toConnectError(e) }
    }

    adapter.registerUnary(
        "workspace.v1.WorkspaceService", "UpdateWorkspace",
        WS.UpdateWorkspaceRequest.getDefaultInstance()
    ) { req: WS.UpdateWorkspaceRequest ->
        try {
            val id = parseUUID(req.id)
            val w = svc.update(UpdateWorkspaceParams(id = id, name = req.name, slug = req.slug))
            WS.UpdateWorkspaceResponse.newBuilder().setWorkspace(w.toProto()).build()
        } catch (e: RepositoryError) { throw toConnectError(e) }
    }

    adapter.registerUnary(
        "workspace.v1.WorkspaceService", "DeleteWorkspace",
        WS.DeleteWorkspaceRequest.getDefaultInstance()
    ) { req: WS.DeleteWorkspaceRequest ->
        try {
            val id = parseUUID(req.id)
            svc.delete(id)
            WS.DeleteWorkspaceResponse.getDefaultInstance()
        } catch (e: RepositoryError) { throw toConnectError(e) }
    }
}

private fun Workspace.toProto(): WS.Workspace = WS.Workspace.newBuilder()
    .setId(id.toString())
    .setName(name)
    .setSlug(slug)
    .setCreatedAt(createdAt.toTimestamp())
    .setUpdatedAt(updatedAt.toTimestamp())
    .build()

internal fun java.time.Instant.toTimestamp(): Timestamp = Timestamp.newBuilder()
    .setSeconds(epochSecond)
    .setNanos(nano)
    .build()

internal fun Timestamp.toInstant(): java.time.Instant =
    java.time.Instant.ofEpochSecond(seconds, nanos.toLong())

internal fun parseUUID(s: String): UUID = try {
    UUID.fromString(s)
} catch (e: IllegalArgumentException) {
    throw ConnectError(ConnectCode.INVALID_ARGUMENT, "invalid UUID: $s")
}
