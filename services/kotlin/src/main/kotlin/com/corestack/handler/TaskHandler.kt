package com.corestack.handler

import com.corestack.connect.ConnectAdapter
import com.corestack.connect.ConnectCode
import com.corestack.connect.ConnectError
import com.corestack.repository.*
import com.corestack.service.ImportService
import com.corestack.service.TaskInput
import com.corestack.service.TaskService
import com.google.protobuf.Struct
import com.google.protobuf.util.JsonFormat
import task.v1.TaskOuterClass as TK
import common.v1.Pagination
import java.util.UUID

private val jsonPrinter = JsonFormat.printer().alwaysPrintFieldsWithNoPresence()
private val jsonParser = JsonFormat.parser().ignoringUnknownFields()

fun registerTaskHandlers(adapter: ConnectAdapter, svc: TaskService, importSvc: ImportService) {
    adapter.registerUnary(
        "task.v1.TaskService", "CreateTask",
        TK.CreateTaskRequest.getDefaultInstance()
    ) { req: TK.CreateTaskRequest ->
        try {
            val workspaceId = parseUUID(req.workspaceId)
            val projectId = parseUUID(req.projectId)
            val params = CreateTaskParams(
                workspaceId = workspaceId,
                projectId = projectId,
                title = req.title,
                description = req.description,
                priority = req.priority,
                assignedTo = req.assignedTo,
                dueDate = if (req.hasDueDate()) req.dueDate.toInstant() else null,
                metadata = if (req.hasMetadata()) structToJson(req.metadata) else null
            )
            val task = svc.create(params)
            TK.CreateTaskResponse.newBuilder().setTask(task.toProto()).build()
        } catch (e: RepositoryError) { throw toConnectError(e) }
    }

    adapter.registerUnary(
        "task.v1.TaskService", "GetTask",
        TK.GetTaskRequest.getDefaultInstance()
    ) { req: TK.GetTaskRequest ->
        try {
            val id = parseUUID(req.id)
            val task = svc.getById(id)
            TK.GetTaskResponse.newBuilder().setTask(task.toProto()).build()
        } catch (e: RepositoryError) { throw toConnectError(e) }
    }

    adapter.registerUnary(
        "task.v1.TaskService", "ListTasks",
        TK.ListTasksRequest.getDefaultInstance()
    ) { req: TK.ListTasksRequest ->
        try {
            val workspaceId = parseUUID(req.workspaceId)
            val projectId = if (req.projectId.isNotEmpty()) parseUUID(req.projectId) else UUID(0, 0)
            val params = ListTasksParams(
                workspaceId = workspaceId,
                projectId = projectId,
                status = req.status,
                priority = req.priority,
                assignedTo = req.assignedTo,
                pageSize = if (req.hasPagination()) req.pagination.pageSize else 0,
                pageToken = if (req.hasPagination()) req.pagination.pageToken else ""
            )
            val list = svc.list(params)
            TK.ListTasksResponse.newBuilder()
                .addAllTasks(list.tasks.map { it.toProto() })
                .setPagination(Pagination.PaginationResponse.newBuilder()
                    .setNextPageToken(list.nextPageToken)
                    .setTotalCount(list.totalCount)
                    .build())
                .build()
        } catch (e: RepositoryError) { throw toConnectError(e) }
    }

    adapter.registerUnary(
        "task.v1.TaskService", "UpdateTask",
        TK.UpdateTaskRequest.getDefaultInstance()
    ) { req: TK.UpdateTaskRequest ->
        try {
            val id = parseUUID(req.id)
            val params = UpdateTaskParams(
                id = id,
                title = req.title,
                description = req.description,
                status = req.status,
                priority = req.priority,
                assignedTo = req.assignedTo,
                dueDate = if (req.hasDueDate()) req.dueDate.toInstant() else null,
                metadata = if (req.hasMetadata()) structToJson(req.metadata) else null
            )
            val task = svc.update(params)
            TK.UpdateTaskResponse.newBuilder().setTask(task.toProto()).build()
        } catch (e: RepositoryError) { throw toConnectError(e) }
    }

    adapter.registerUnary(
        "task.v1.TaskService", "DeleteTask",
        TK.DeleteTaskRequest.getDefaultInstance()
    ) { req: TK.DeleteTaskRequest ->
        try {
            val id = parseUUID(req.id)
            svc.delete(id)
            TK.DeleteTaskResponse.getDefaultInstance()
        } catch (e: RepositoryError) { throw toConnectError(e) }
    }

    adapter.registerUnary(
        "task.v1.TaskService", "BulkImportTasks",
        TK.BulkImportTasksRequest.getDefaultInstance()
    ) { req: TK.BulkImportTasksRequest ->
        val workspaceId = parseUUID(req.workspaceId)
        val projectId = parseUUID(req.projectId)
        val inputs = req.tasksList.map { t ->
            TaskInput(
                title = t.title,
                description = t.description,
                priority = t.priority,
                assignedTo = t.assignedTo,
                dueDate = if (t.hasDueDate()) t.dueDate.toInstant() else null,
                metadata = if (t.hasMetadata()) structToJson(t.metadata) else null
            )
        }
        val result = importSvc.bulkImport(workspaceId, projectId, inputs)
        TK.BulkImportTasksResponse.newBuilder()
            .setTotal(result.total)
            .setSucceeded(result.succeeded)
            .setFailed(result.failed)
            .addAllErrors(result.errors.map { e ->
                TK.TaskError.newBuilder()
                    .setIndex(e.index)
                    .setError(e.error)
                    .build()
            })
            .build()
    }
}

private fun Task.toProto(): TK.Task {
    val builder = TK.Task.newBuilder()
        .setId(id.toString())
        .setWorkspaceId(workspaceId.toString())
        .setProjectId(projectId.toString())
        .setTitle(title)
        .setDescription(description)
        .setStatus(status)
        .setPriority(priority)
        .setAssignedTo(assignedTo)
        .setCreatedAt(createdAt.toTimestamp())
        .setUpdatedAt(updatedAt.toTimestamp())

    dueDate?.let { builder.setDueDate(it.toTimestamp()) }

    if (metadata.isNotEmpty() && metadata != "{}") {
        val structBuilder = Struct.newBuilder()
        jsonParser.merge(metadata, structBuilder)
        builder.setMetadata(structBuilder.build())
    }

    return builder.build()
}

private fun structToJson(struct: Struct): String {
    return jsonPrinter.print(struct)
}
