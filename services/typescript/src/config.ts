import postgres from "postgres";
import pino from "pino";

export interface Config {
  databaseUrl: string;
  grpcPort: number;
  logLevel: string;
  workerConcurrency: number;
}

export function loadConfig(): Config {
  return {
    databaseUrl:
      process.env.DATABASE_URL ??
      "postgres://postgres:postgres@localhost:5432/api_corestack?sslmode=disable",
    grpcPort: parseInt(process.env.GRPC_PORT ?? "8080", 10),
    logLevel: process.env.LOG_LEVEL ?? "debug",
    workerConcurrency: parseInt(process.env.WORKER_CONCURRENCY ?? "10", 10),
  };
}

export function createSql(databaseUrl: string) {
  return postgres(databaseUrl, {
    max: 20,
    idle_timeout: 20,
    connect_timeout: 10,
  });
}

export function createLogger(level: string) {
  return pino({ level });
}
