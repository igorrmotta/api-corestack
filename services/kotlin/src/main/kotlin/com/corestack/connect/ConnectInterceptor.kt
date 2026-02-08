package com.corestack.connect

import com.google.protobuf.Message

typealias ConnectInterceptor = suspend (procedure: String, next: suspend () -> Message) -> Message
