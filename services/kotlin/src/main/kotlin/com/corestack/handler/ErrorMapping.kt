package com.corestack.handler

import com.corestack.connect.ConnectCode
import com.corestack.connect.ConnectError
import com.corestack.repository.*

fun toConnectError(e: Exception): ConnectError = when (e) {
    is NotFoundError -> ConnectError(ConnectCode.NOT_FOUND, e.message ?: "not found")
    is InvalidInputError -> ConnectError(ConnectCode.INVALID_ARGUMENT, e.message ?: "invalid input")
    is AlreadyExistsError -> ConnectError(ConnectCode.ALREADY_EXISTS, e.message ?: "already exists")
    is ConflictError -> ConnectError(ConnectCode.ALREADY_EXISTS, e.message ?: "conflict")
    is ConnectError -> e
    else -> ConnectError(ConnectCode.INTERNAL, e.message ?: "internal error")
}
