# services/golang/ — Go Reference Implementation

The first and reference implementation of the Task Management API. Built with Connect RPC, pgx, and River.

## Prerequisites

- Go 1.24+
- Docker (for PostgreSQL)
- [Buf CLI](https://buf.build/docs/installation)

## Getting Started

```bash
# 1. Generate protobuf code (required before compiling)
make proto-gen        # from repo root

# 2. Start PostgreSQL
make db-up            # from repo root

# 3. Run database migrations
make db-migrate       # from repo root

# 4. Start the API server
make go-server        # from repo root (or: go run ./cmd/server)

# 5. Start the background worker (separate terminal)
make go-worker        # from repo root (or: go run ./cmd/worker)
```

The server listens on `http://localhost:8080` and serves both gRPC and REST JSON via h2c.

## Package Structure

```
services/golang/
├── cmd/
│   ├── server/main.go        # HTTP server with Connect RPC handlers
│   └── worker/main.go        # River background job worker
├── gen/                       # Generated protobuf code (gitignored)
├── internal/
│   ├── config/                # Environment-based configuration
│   ├── domain/                # Entities, repository interfaces, domain errors
│   ├── handler/               # Connect RPC handlers (proto ↔ domain translation)
│   ├── service/               # Business logic
│   ├── repository/            # pgx implementations + SQL queries
│   │   └── queries/           # Raw SQL for sqlc
│   ├── middleware/            # Logging + recovery interceptors
│   └── worker/                # River job definitions
├── go.mod
├── sqlc.yaml
├── Makefile
└── Dockerfile
```

## Key Dependencies

| Package | Purpose |
|---|---|
| `connectrpc.com/connect` | Connect RPC framework |
| `github.com/jackc/pgx/v5` | PostgreSQL driver + connection pool |
| `github.com/riverqueue/river` | Background job processing |
| `github.com/google/uuid` | UUID generation |
| `golang.org/x/sync` | errgroup for concurrent operations |
| `golang.org/x/time` | Rate limiter for bulk import |
| `github.com/stretchr/testify` | Test assertions |
| `google.golang.org/protobuf` | Protobuf runtime |

## Configuration

Environment variables (with defaults from `.env.example`):

| Variable | Default | Description |
|---|---|---|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/api_corestack?sslmode=disable` | PostgreSQL connection string |
| `GRPC_PORT` | `8080` | Server listen port |
| `LOG_LEVEL` | `debug` | Log level (debug, info, warn, error) |
| `RIVER_CONCURRENCY` | `10` | Max concurrent background workers |

## Make Targets (local)

Run from `services/golang/`:

| Target | Command |
|---|---|
| `make build` | Build server + worker binaries to `bin/` |
| `make run-server` | `go run ./cmd/server` |
| `make run-worker` | `go run ./cmd/worker` |
| `make test` | Run all tests |
| `make test-unit` | Run service + handler tests only |
| `make test-integration` | Run repository tests only |
| `make lint` | Run golangci-lint |
| `make sqlc-generate` | Regenerate sqlc code from SQL queries |

## Docker

```bash
# Build and run everything (from repo root)
make docker-go

# Or build the image directly
docker build -t api-corestack-go services/golang/
```

The Dockerfile uses a multi-stage build:
1. `golang:1.23-alpine` — compiles server and worker binaries
2. `alpine:3.19` — minimal runtime image

The default command runs the server. To run the worker:
```bash
docker run api-corestack-go /app/worker
```

## Testing

```bash
# All Go tests
make go-test          # from repo root

# Unit tests (services + handlers)
cd services/golang && make test-unit

# Integration tests (repositories, requires running PostgreSQL)
cd services/golang && make test-integration

# API tests via Bruno (requires running server)
make test-bruno       # from repo root
```
