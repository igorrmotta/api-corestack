package com.corestack.service

import com.corestack.repository.*
import io.github.oshai.kotlinlogging.KotlinLogging
import java.util.UUID

private val logger = KotlinLogging.logger {}

class ProjectService(private val repo: ProjectRepo) {

    suspend fun create(params: CreateProjectParams): Project {
        if (params.name.isBlank()) throw InvalidInputError("invalid input: name is required")
        if (params.workspaceId == UUID(0, 0)) throw InvalidInputError("invalid input: workspace_id is required")
        logger.debug { "creating project name=${params.name} workspace_id=${params.workspaceId}" }
        return repo.create(params)
    }

    suspend fun getById(id: UUID): Project = repo.getById(id)

    suspend fun list(params: ListProjectsParams): ProjectList = repo.list(params)

    suspend fun update(params: UpdateProjectParams): Project {
        if (params.name.isBlank()) throw InvalidInputError("invalid input: name is required")
        val validStatuses = setOf("active", "archived")
        if (params.status.isNotEmpty() && params.status !in validStatuses) {
            throw InvalidInputError("invalid input: invalid status: ${params.status}")
        }
        return repo.update(params)
    }

    suspend fun delete(id: UUID) = repo.delete(id)
}
