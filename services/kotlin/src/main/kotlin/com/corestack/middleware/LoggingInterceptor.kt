package com.corestack.middleware

import com.corestack.connect.ConnectInterceptor
import com.google.protobuf.Message
import io.github.oshai.kotlinlogging.KotlinLogging

private val logger = KotlinLogging.logger {}

fun loggingInterceptor(): ConnectInterceptor = { procedure, next ->
    val start = System.nanoTime()
    logger.info { "rpc started procedure=$procedure" }
    try {
        val result = next()
        val durationMs = (System.nanoTime() - start) / 1_000_000
        logger.info { "rpc completed procedure=$procedure duration=${durationMs}ms" }
        result
    } catch (e: Exception) {
        val durationMs = (System.nanoTime() - start) / 1_000_000
        logger.error { "rpc failed procedure=$procedure duration=${durationMs}ms error=${e.message}" }
        throw e
    }
}
