package com.corestack.connect

enum class ConnectCode(val code: String, val httpStatus: Int) {
    OK("ok", 200),
    CANCELLED("canceled", 408),
    UNKNOWN("unknown", 500),
    INVALID_ARGUMENT("invalid_argument", 400),
    DEADLINE_EXCEEDED("deadline_exceeded", 408),
    NOT_FOUND("not_found", 404),
    ALREADY_EXISTS("already_exists", 409),
    PERMISSION_DENIED("permission_denied", 403),
    RESOURCE_EXHAUSTED("resource_exhausted", 429),
    FAILED_PRECONDITION("failed_precondition", 412),
    ABORTED("aborted", 409),
    OUT_OF_RANGE("out_of_range", 400),
    UNIMPLEMENTED("unimplemented", 404),
    INTERNAL("internal", 500),
    UNAVAILABLE("unavailable", 503),
    DATA_LOSS("data_loss", 500),
    UNAUTHENTICATED("unauthenticated", 401);
}

class ConnectError(
    val connectCode: ConnectCode,
    override val message: String,
    override val cause: Throwable? = null
) : Exception(message, cause)
