# services/typescript/ — TypeScript Implementation

The TypeScript implementation of the Task Management API. Built with Connect RPC, postgres.js, and graphile-worker.

## Prerequisites

- Node.js 22+
- pnpm
- Docker (for PostgreSQL)
- [Buf CLI](https://buf.build/docs/installation)

## Getting Started

```bash
# 1. Generate protobuf code (required before compiling)
make proto-gen        # from repo root

# 2. Install dependencies
make ts-install       # from repo root (or: cd services/typescript && pnpm install)

# 3. Start PostgreSQL
make db-up            # from repo root

# 4. Run database migrations
make db-migrate       # from repo root

# 5. Start the API server
make ts-server        # from repo root (or: cd services/typescript && pnpm dev:server)

# 6. Start the background worker (separate terminal)
make ts-worker        # from repo root (or: cd services/typescript && pnpm dev:worker)
```

The server listens on `http://localhost:8080` and serves both gRPC and REST JSON via Connect protocol.

## Directory Structure

```
services/typescript/
├── src/
│   ├── server.ts                # HTTP server with Connect RPC handlers
│   ├── worker.ts                # graphile-worker background job processor
│   ├── config.ts                # Environment-based configuration
│   ├── handler/                 # Connect RPC handlers (proto ↔ repository type translation)
│   │   └── error-mapping.ts     # Repository error → Connect RPC code mapping
│   ├── service/                 # Business logic
│   ├── repository/              # postgres.js implementations, entities, errors
│   └── middleware/              # Logging + recovery interceptors
├── gen/                         # Generated protobuf code (gitignored)
├── dist/                        # Build output (gitignored)
├── package.json
├── tsconfig.json
├── tsup.config.ts
└── Dockerfile
```

## Key Dependencies

| Package | Purpose |
|---|---|
| `@connectrpc/connect` | Connect RPC framework |
| `@connectrpc/connect-node` | Node.js adapter for Connect |
| `@bufbuild/protobuf` | Protobuf-ES v2 runtime |
| `postgres` (postgres.js) | PostgreSQL client with tagged template literals |
| `graphile-worker` | PostgreSQL-based background job processing |
| `p-queue` | Concurrency control for bulk import |
| `pino` | JSON structured logging |
| `tsup` (dev) | Fast TypeScript bundler (esbuild) |
| `tsx` (dev) | TypeScript execution without build step |
| `vitest` (dev) | Test runner |

## Configuration

Environment variables (with defaults):

| Variable | Default | Description |
|---|---|---|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/api_corestack?sslmode=disable` | PostgreSQL connection string |
| `GRPC_PORT` | `8080` | Server listen port |
| `LOG_LEVEL` | `debug` | Log level (debug, info, warn, error) |
| `WORKER_CONCURRENCY` | `10` | Max concurrent background workers |

## Scripts

Run from `services/typescript/`:

| Script | Command |
|---|---|
| `pnpm build` | Build server + worker to `dist/` via tsup |
| `pnpm dev:server` | Run server with tsx (no build step) |
| `pnpm dev:worker` | Run worker with tsx (no build step) |
| `pnpm start:server` | Run built server (`node dist/server.js`) |
| `pnpm start:worker` | Run built worker (`node dist/worker.js`) |
| `pnpm test` | Run tests with vitest |

## Docker

```bash
# Build and run everything (from repo root)
make docker-ts

# Or build the image directly
docker build -t api-corestack-ts services/typescript/
```

The Dockerfile uses a multi-stage build:
1. `node:22-alpine` — installs dependencies, generates protobuf code, builds with tsup
2. `node:22-alpine` — production dependencies + `dist/` and `gen/`

The default command runs the server. To run the worker:
```bash
docker run api-corestack-ts node dist/worker.js
```

## Testing

```bash
# Unit tests
cd services/typescript && pnpm test

# API tests via Bruno (requires running server)
make test-bruno-ts    # from repo root
```
