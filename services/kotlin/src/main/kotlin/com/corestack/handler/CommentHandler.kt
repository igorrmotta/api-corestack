package com.corestack.handler

import com.corestack.connect.ConnectAdapter
import com.corestack.repository.*
import com.corestack.service.CommentService
import comment.v1.CommentOuterClass as CM
import common.v1.Pagination

fun registerCommentHandlers(adapter: ConnectAdapter, svc: CommentService) {
    adapter.registerUnary(
        "comment.v1.CommentService", "CreateComment",
        CM.CreateCommentRequest.getDefaultInstance()
    ) { req: CM.CreateCommentRequest ->
        try {
            val taskId = parseUUID(req.taskId)
            val c = svc.create(CreateCommentParams(
                taskId = taskId,
                authorId = req.authorId,
                content = req.content
            ))
            CM.CreateCommentResponse.newBuilder().setComment(c.toProto()).build()
        } catch (e: RepositoryError) { throw toConnectError(e) }
    }

    adapter.registerUnary(
        "comment.v1.CommentService", "ListComments",
        CM.ListCommentsRequest.getDefaultInstance()
    ) { req: CM.ListCommentsRequest ->
        try {
            val taskId = parseUUID(req.taskId)
            val params = ListCommentsParams(
                taskId = taskId,
                pageSize = if (req.hasPagination()) req.pagination.pageSize else 0,
                pageToken = if (req.hasPagination()) req.pagination.pageToken else ""
            )
            val list = svc.list(params)
            CM.ListCommentsResponse.newBuilder()
                .addAllComments(list.comments.map { it.toProto() })
                .setPagination(Pagination.PaginationResponse.newBuilder()
                    .setNextPageToken(list.nextPageToken)
                    .setTotalCount(list.totalCount)
                    .build())
                .build()
        } catch (e: RepositoryError) { throw toConnectError(e) }
    }

    adapter.registerUnary(
        "comment.v1.CommentService", "DeleteComment",
        CM.DeleteCommentRequest.getDefaultInstance()
    ) { req: CM.DeleteCommentRequest ->
        try {
            val id = parseUUID(req.id)
            svc.delete(id)
            CM.DeleteCommentResponse.getDefaultInstance()
        } catch (e: RepositoryError) { throw toConnectError(e) }
    }
}

private fun Comment.toProto(): CM.Comment = CM.Comment.newBuilder()
    .setId(id.toString())
    .setTaskId(taskId.toString())
    .setAuthorId(authorId)
    .setContent(content)
    .setCreatedAt(createdAt.toTimestamp())
    .setUpdatedAt(updatedAt.toTimestamp())
    .build()
