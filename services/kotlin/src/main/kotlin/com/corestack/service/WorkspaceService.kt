package com.corestack.service

import com.corestack.repository.*
import io.github.oshai.kotlinlogging.KotlinLogging
import java.util.UUID

private val logger = KotlinLogging.logger {}

class WorkspaceService(private val repo: WorkspaceRepo) {

    suspend fun create(params: CreateWorkspaceParams): Workspace {
        if (params.name.isBlank()) throw InvalidInputError("invalid input: name is required")
        if (params.slug.isBlank()) throw InvalidInputError("invalid input: slug is required")
        logger.debug { "creating workspace name=${params.name} slug=${params.slug}" }
        return repo.create(params)
    }

    suspend fun getById(id: UUID): Workspace = repo.getById(id)

    suspend fun list(params: ListWorkspacesParams): WorkspaceList = repo.list(params)

    suspend fun update(params: UpdateWorkspaceParams): Workspace {
        if (params.name.isBlank()) throw InvalidInputError("invalid input: name is required")
        return repo.update(params)
    }

    suspend fun delete(id: UUID) = repo.delete(id)
}
