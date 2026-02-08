package com.corestack.handler

import com.corestack.connect.ConnectAdapter
import com.corestack.repository.*
import com.corestack.service.ProjectService
import project.v1.ProjectOuterClass as PJ
import common.v1.Pagination

fun registerProjectHandlers(adapter: ConnectAdapter, svc: ProjectService) {
    adapter.registerUnary(
        "project.v1.ProjectService", "CreateProject",
        PJ.CreateProjectRequest.getDefaultInstance()
    ) { req: PJ.CreateProjectRequest ->
        try {
            val workspaceId = parseUUID(req.workspaceId)
            val p = svc.create(CreateProjectParams(
                workspaceId = workspaceId,
                name = req.name,
                description = req.description
            ))
            PJ.CreateProjectResponse.newBuilder().setProject(p.toProto()).build()
        } catch (e: RepositoryError) { throw toConnectError(e) }
    }

    adapter.registerUnary(
        "project.v1.ProjectService", "GetProject",
        PJ.GetProjectRequest.getDefaultInstance()
    ) { req: PJ.GetProjectRequest ->
        try {
            val id = parseUUID(req.id)
            val p = svc.getById(id)
            PJ.GetProjectResponse.newBuilder().setProject(p.toProto()).build()
        } catch (e: RepositoryError) { throw toConnectError(e) }
    }

    adapter.registerUnary(
        "project.v1.ProjectService", "ListProjects",
        PJ.ListProjectsRequest.getDefaultInstance()
    ) { req: PJ.ListProjectsRequest ->
        try {
            val workspaceId = parseUUID(req.workspaceId)
            val params = ListProjectsParams(
                workspaceId = workspaceId,
                pageSize = if (req.hasPagination()) req.pagination.pageSize else 0,
                pageToken = if (req.hasPagination()) req.pagination.pageToken else ""
            )
            val list = svc.list(params)
            PJ.ListProjectsResponse.newBuilder()
                .addAllProjects(list.projects.map { it.toProto() })
                .setPagination(Pagination.PaginationResponse.newBuilder()
                    .setNextPageToken(list.nextPageToken)
                    .setTotalCount(list.totalCount)
                    .build())
                .build()
        } catch (e: RepositoryError) { throw toConnectError(e) }
    }

    adapter.registerUnary(
        "project.v1.ProjectService", "UpdateProject",
        PJ.UpdateProjectRequest.getDefaultInstance()
    ) { req: PJ.UpdateProjectRequest ->
        try {
            val id = parseUUID(req.id)
            val p = svc.update(UpdateProjectParams(
                id = id,
                name = req.name,
                description = req.description,
                status = req.status
            ))
            PJ.UpdateProjectResponse.newBuilder().setProject(p.toProto()).build()
        } catch (e: RepositoryError) { throw toConnectError(e) }
    }

    adapter.registerUnary(
        "project.v1.ProjectService", "DeleteProject",
        PJ.DeleteProjectRequest.getDefaultInstance()
    ) { req: PJ.DeleteProjectRequest ->
        try {
            val id = parseUUID(req.id)
            svc.delete(id)
            PJ.DeleteProjectResponse.getDefaultInstance()
        } catch (e: RepositoryError) { throw toConnectError(e) }
    }
}

private fun Project.toProto(): PJ.Project = PJ.Project.newBuilder()
    .setId(id.toString())
    .setWorkspaceId(workspaceId.toString())
    .setName(name)
    .setDescription(description)
    .setStatus(status)
    .setCreatedAt(createdAt.toTimestamp())
    .setUpdatedAt(updatedAt.toTimestamp())
    .build()
