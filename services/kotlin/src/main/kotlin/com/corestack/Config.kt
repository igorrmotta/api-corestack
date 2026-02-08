package com.corestack

import com.zaxxer.hikari.HikariConfig
import com.zaxxer.hikari.HikariDataSource
import org.jdbi.v3.core.Jdbi
import org.jdbi.v3.core.kotlin.KotlinPlugin
import org.jdbi.v3.postgres.PostgresPlugin

data class AppConfig(
    val databaseUrl: String,
    val grpcPort: Int,
    val logLevel: String,
    val workerConcurrency: Int
)

fun loadConfig(): AppConfig {
    return AppConfig(
        databaseUrl = System.getenv("DATABASE_URL") ?: "postgres://postgres:postgres@localhost:5432/api_corestack?sslmode=disable",
        grpcPort = System.getenv("GRPC_PORT")?.toIntOrNull() ?: 8083,
        logLevel = System.getenv("LOG_LEVEL") ?: "info",
        workerConcurrency = System.getenv("WORKER_CONCURRENCY")?.toIntOrNull() ?: 10
    )
}

fun toJdbcUrl(postgresUrl: String): String {
    // Convert postgres://user:pass@host:port/db?params to jdbc:postgresql://host:port/db?params&user=user&password=pass
    val url = if (postgresUrl.startsWith("postgres://")) {
        postgresUrl.replaceFirst("postgres://", "postgresql://")
    } else if (postgresUrl.startsWith("postgresql://")) {
        postgresUrl
    } else {
        return postgresUrl // Already JDBC format
    }

    val withoutScheme = url.removePrefix("postgresql://")
    val atIndex = withoutScheme.indexOf('@')
    if (atIndex == -1) {
        return "jdbc:postgresql://$withoutScheme"
    }

    val credentials = withoutScheme.substring(0, atIndex)
    val rest = withoutScheme.substring(atIndex + 1)

    val colonIndex = credentials.indexOf(':')
    val user = if (colonIndex >= 0) credentials.substring(0, colonIndex) else credentials
    val password = if (colonIndex >= 0) credentials.substring(colonIndex + 1) else ""

    val separator = if (rest.contains('?')) '&' else '?'
    return "jdbc:postgresql://$rest${separator}user=$user&password=$password"
}

fun createDataSource(config: AppConfig): HikariDataSource {
    val hikariConfig = HikariConfig().apply {
        jdbcUrl = toJdbcUrl(config.databaseUrl)
        maximumPoolSize = 20
        minimumIdle = 5
        idleTimeout = 20000
        connectionTimeout = 10000
        driverClassName = "org.postgresql.Driver"
    }
    return HikariDataSource(hikariConfig)
}

fun createJdbi(dataSource: HikariDataSource): Jdbi {
    return Jdbi.create(dataSource)
        .installPlugin(KotlinPlugin())
        .installPlugin(PostgresPlugin())
}
