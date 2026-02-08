# Kotlin Implementation

Kotlin implementation of the Task Management API using Ktor, JDBI, and a custom Connect protocol adapter.

## Technology Stack

| Concern | Choice |
|---|---|
| Language | Kotlin 2.1 / JDK 21 |
| HTTP Server | Ktor 3.x (Netty) |
| Connect Protocol | Custom adapter on Ktor |
| PostgreSQL | JDBI 3.x + HikariCP |
| Protobuf | protobuf-java + protobuf-kotlin + protobuf-java-util |
| Background Jobs | Coroutine-based polling |
| Concurrency | kotlinx.coroutines + Semaphore |
| Logging | kotlin-logging + Logback + logstash-logback-encoder |
| Build | Gradle Kotlin DSL + Shadow plugin |

## Project Structure

```
src/main/kotlin/com/corestack/
  Main.kt                         # Entry point (server or worker)
  Server.kt                       # Ktor server setup
  Worker.kt                       # Background worker
  Config.kt                       # Environment config + JDBC
  connect/
    ConnectAdapter.kt              # Connect protocol router
    ConnectError.kt                # Error codes + HTTP mapping
    ConnectInterceptor.kt          # Interceptor type
  repository/                      # Data access layer
  service/                         # Business logic
  handler/                         # Proto translation + registration
  middleware/                      # Logging + recovery interceptors
  worker/
    NotificationProcessor.kt       # Polling notification processor
gen/                               # Generated protobuf (gitignored)
```

## Commands

```bash
# Build
./gradlew shadowJar

# Run server
./gradlew run

# Run worker
java -jar build/libs/app.jar worker

# Run tests
./gradlew test
```

## Port

This implementation runs on port **8083**.
