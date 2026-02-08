package com.corestack.connect

import com.google.protobuf.Message
import com.google.protobuf.util.JsonFormat
import io.ktor.http.*
import io.ktor.server.request.*
import io.ktor.server.response.*
import io.ktor.server.routing.*

class ConnectAdapter {
    private data class Route(
        val path: String,
        val defaultInstance: Message,
        val handler: suspend (Message) -> Message
    )

    private val routes = mutableListOf<Route>()
    private val interceptors = mutableListOf<ConnectInterceptor>()

    private val jsonParser: JsonFormat.Parser = JsonFormat.parser().ignoringUnknownFields()
    private val jsonPrinter: JsonFormat.Printer = JsonFormat.printer().alwaysPrintFieldsWithNoPresence()

    fun addInterceptor(interceptor: ConnectInterceptor) {
        interceptors.add(interceptor)
    }

    fun <Req : Message, Res : Message> registerUnary(
        service: String,
        method: String,
        defaultInstance: Req,
        handler: suspend (Req) -> Res
    ) {
        val path = "/$service/$method"
        routes.add(Route(
            path = path,
            defaultInstance = defaultInstance,
            handler = { msg ->
                @Suppress("UNCHECKED_CAST")
                handler(msg as Req) as Message
            }
        ))
    }

    fun install(routing: Routing) {
        for (route in routes) {
            routing.post(route.path) {
                try {
                    // Parse request
                    val body = call.receiveText()
                    val builder = route.defaultInstance.newBuilderForType()
                    if (body.isNotBlank()) {
                        jsonParser.merge(body, builder)
                    }
                    val request = builder.build()

                    // Build interceptor chain
                    val procedure = route.path
                    val baseHandler: suspend () -> Message = { route.handler(request) }
                    val chain = interceptors.foldRight(baseHandler) { interceptor, next ->
                        { interceptor(procedure, next) }
                    }

                    // Execute
                    val response = chain()

                    // Serialize response
                    val json = jsonPrinter.print(response)
                    call.respondText(json, ContentType.Application.Json, HttpStatusCode.OK)
                } catch (e: ConnectError) {
                    val errorJson = """{"code":"${e.connectCode.code}","message":"${escapeJson(e.message)}"}"""
                    call.respondText(
                        errorJson,
                        ContentType.Application.Json,
                        HttpStatusCode.fromValue(e.connectCode.httpStatus)
                    )
                } catch (e: Exception) {
                    val errorJson = """{"code":"internal","message":"${escapeJson(e.message ?: "Internal error")}"}"""
                    call.respondText(
                        errorJson,
                        ContentType.Application.Json,
                        HttpStatusCode.InternalServerError
                    )
                }
            }
        }
    }

    private fun escapeJson(s: String): String {
        return s.replace("\\", "\\\\")
            .replace("\"", "\\\"")
            .replace("\n", "\\n")
            .replace("\r", "\\r")
            .replace("\t", "\\t")
    }
}
