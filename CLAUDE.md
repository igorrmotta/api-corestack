# CLAUDE.md

Polyglot Task Management API. Go is the reference implementation. All language implementations share the same PostgreSQL database, protobuf definitions (api/), and Bruno test suite (tests/bruno/).

## Repository Structure

```
api/                          # Protobuf definitions (Buf v2, 6 modules)
db/migrations/                # dbmate SQL migrations (6 tables)
services/golang/              # Go reference implementation
  cmd/server/main.go          #   HTTP server (Connect RPC, h2c, single port)
  cmd/worker/main.go          #   River background worker
  internal/domain/            #   Entities + repository interfaces + errors
  internal/handler/           #   Connect RPC handlers
  internal/service/           #   Business logic
  internal/repository/        #   pgx implementations
  internal/middleware/        #   Logging + recovery interceptors
  internal/worker/            #   River job definitions
  gen/                        #   Generated protobuf code (gitignored)
tests/bruno/                  # Language-agnostic API test collections
```

## Common Commands

```bash
# Database
make db-up              # Start PostgreSQL via Docker
make db-down            # Stop PostgreSQL
make db-migrate         # Run pending migrations
make db-rollback        # Rollback last migration
make db-new NAME=x      # Create new migration file

# Protobuf
make proto-gen          # Generate code from protos (must run before Go compiles)
make proto-lint         # Lint .proto files
make proto-breaking     # Check breaking changes vs main

# Go
make go-server          # Run Go server locally
make go-worker          # Run Go worker locally
make go-test            # Run all Go tests

# Docker
make docker-go          # Full stack: PostgreSQL + migrations + Go server + worker

# Tests
make test-bruno         # Run Bruno API tests against running server
```

## Architecture Rules

- **Layers: handler → service → repository.** Never skip layers. Handlers translate between proto and domain types. Services contain business logic. Repositories do SQL.
- **Domain is the center.** Entities and repository interfaces live in `internal/domain/`. No imports from handler/service/repository in domain.
- **Error mapping.** Domain errors (`ErrNotFound`, `ErrAlreadyExists`, `ErrInvalidInput`, `ErrConflict`) are mapped to Connect RPC codes (`NotFound`, `AlreadyExists`, `InvalidArgument`, `Aborted`) in handlers.
- **Single port.** Connect RPC serves gRPC and REST JSON on port 8080 via h2c. No separate HTTP gateway.
- **Soft deletes.** Workspaces, projects, and tasks use `deleted_at` column. Queries must filter `WHERE deleted_at IS NULL`.
- **Cursor-based pagination.** `page_token` is the UUID of the last item. Repos query `WHERE id > page_token ORDER BY id LIMIT page_size+1`.

## Code Conventions (Go)

- Wrap errors with `fmt.Errorf("context: %w", err)` for the chain.
- Use `log/slog` for structured logging.
- Table-driven tests with `testify/assert` and `testify/require`.
- Config via environment variables — see `internal/config/config.go`.

## Adding a New RPC

1. Define the RPC in the appropriate `.proto` file under `api/`
2. Run `make proto-gen` to regenerate Go code
3. Add domain types to `internal/domain/` if needed
4. Add repository method + SQL query in `internal/repository/`
5. Add service method in `internal/service/`
6. Implement the handler in `internal/handler/`
7. Add Bruno test in `tests/bruno/`

## Adding a New Language Implementation

1. Create `services/<language>/` directory
2. Add protobuf code generation plugin to `api/buf.gen.yaml`
3. Implement the handler → service → repository layers
4. Provide `cmd/server` and `cmd/worker` entry points
5. Add a `Dockerfile` (multi-stage build)
6. Add a Docker Compose profile in `docker-compose.yml`
7. Add a Bruno environment in `tests/bruno/environments/`
8. Verify all Bruno tests pass against the new implementation
