package com.corestack

import com.corestack.worker.NotificationProcessor
import com.corestack.repository.NotificationRepo
import io.github.oshai.kotlinlogging.KotlinLogging
import kotlinx.coroutines.runBlocking

private val logger = KotlinLogging.logger {}

fun startWorker() {
    val config = loadConfig()
    val dataSource = createDataSource(config)
    val jdbi = createJdbi(dataSource)

    logger.info { "worker connected to database" }

    val notifRepo = NotificationRepo(jdbi)
    val processor = NotificationProcessor(notifRepo, config.workerConcurrency)

    Runtime.getRuntime().addShutdownHook(Thread {
        logger.info { "worker shutting down" }
        processor.stop()
        dataSource.close()
        logger.info { "worker stopped" }
    })

    logger.info { "worker starting concurrency=${config.workerConcurrency}" }
    runBlocking {
        processor.run()
    }
}
