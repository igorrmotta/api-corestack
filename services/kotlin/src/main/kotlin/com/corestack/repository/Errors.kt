package com.corestack.repository

import org.postgresql.util.PSQLException

sealed class RepositoryError(message: String) : Exception(message)
class NotFoundError(message: String = "not found") : RepositoryError(message)
class AlreadyExistsError(message: String = "already exists") : RepositoryError(message)
class InvalidInputError(message: String = "invalid input") : RepositoryError(message)
class ConflictError(message: String = "conflict") : RepositoryError(message)

fun isPgUniqueViolation(e: Exception): Boolean {
    val cause = if (e is PSQLException) e else e.cause as? PSQLException
    return cause?.sqlState == "23505"
}
