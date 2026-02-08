package com.corestack

fun main(args: Array<String>) {
    if (args.isNotEmpty() && args[0] == "worker") {
        startWorker()
    } else {
        startServer()
    }
}
