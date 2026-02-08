package com.corestack.middleware

import com.corestack.connect.ConnectCode
import com.corestack.connect.ConnectError
import com.corestack.connect.ConnectInterceptor
import io.github.oshai.kotlinlogging.KotlinLogging

private val logger = KotlinLogging.logger {}

fun recoveryInterceptor(): ConnectInterceptor = { procedure, next ->
    try {
        next()
    } catch (e: ConnectError) {
        throw e
    } catch (e: Exception) {
        logger.error(e) { "panic recovered procedure=$procedure" }
        throw ConnectError(ConnectCode.INTERNAL, "internal error")
    }
}
