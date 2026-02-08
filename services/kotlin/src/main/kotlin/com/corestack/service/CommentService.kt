package com.corestack.service

import com.corestack.repository.*
import io.github.oshai.kotlinlogging.KotlinLogging
import java.util.UUID

private val logger = KotlinLogging.logger {}

class CommentService(private val repo: CommentRepo) {

    suspend fun create(params: CreateCommentParams): Comment {
        if (params.taskId == UUID(0, 0)) throw InvalidInputError("invalid input: task_id is required")
        if (params.authorId.isBlank()) throw InvalidInputError("invalid input: author_id is required")
        if (params.content.isBlank()) throw InvalidInputError("invalid input: content is required")
        logger.debug { "creating comment task_id=${params.taskId} author_id=${params.authorId}" }
        return repo.create(params)
    }

    suspend fun list(params: ListCommentsParams): CommentList = repo.list(params)

    suspend fun delete(id: UUID) = repo.delete(id)
}
