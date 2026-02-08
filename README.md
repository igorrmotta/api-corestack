# api-corestack

A polyglot Task Management API built as a learning project to explore how the same API spec can be implemented across multiple programming languages — all sharing one PostgreSQL database, protobuf contract, and test suite.

Go is the reference implementation. TypeScript is the second complete implementation. Other languages are planned.

## Architecture

```
                    ┌──────────────────────────────────┐
                    │         api/ (Protobuf)           │
                    │   Shared contract for all impls   │
                    └──────────┬───────────────────────┘
                               │ buf generate
          ┌────────────────────┼────────────────────┐
          ▼                    ▼                     ▼
   ┌─────────────┐     ┌─────────────┐      ┌─────────────┐
   │  Go :8080   │     │  TS :8081   │      │ Rust :8082  │
   │  (complete) │     │ (complete)  │      │  (planned)  │
   └──────┬──────┘     └──────┬──────┘      └──────┬──────┘
          │                   │                     │
          └───────────────────┼─────────────────────┘
                              ▼
                    ┌──────────────────┐
                    │  PostgreSQL 16   │
                    │  (shared DB)     │
                    └──────────────────┘
                              ▲
                    ┌─────────┴────────┐
                    │  tests/bruno/    │
                    │  Language-agnostic│
                    │  API tests       │
                    └──────────────────┘
```

## Repository Structure

```
api-corestack/
├── api/                    # Protobuf definitions (Buf v2 workspace)
├── db/migrations/          # dbmate SQL migrations
├── services/
│   ├── golang/             # Go reference implementation
│   └── typescript/         # TypeScript implementation
├── tests/bruno/            # Bruno API test collections
├── docker-compose.yml
├── Makefile
└── .env.example
```

## Quick Start

### Prerequisites

- [Docker](https://www.docker.com/) and Docker Compose
- [Go 1.24+](https://go.dev/) (for Go implementation)
- [Node.js 22+](https://nodejs.org/) and [pnpm](https://pnpm.io/) (for TypeScript implementation)
- [Buf CLI](https://buf.build/docs/installation) (for protobuf generation)
- [Bruno](https://www.usebruno.com/) (optional, for API testing)

### Run with Docker (easiest)

```bash
# Go implementation
make docker-go

# TypeScript implementation
make docker-ts
```

The API is available at `http://localhost:8080`.

### Run Locally (Go)

```bash
# 1. Start PostgreSQL
make db-up

# 2. Run migrations
make db-migrate

# 3. Generate protobuf code
make proto-gen

# 4. Start the Go server
make go-server

# 5. (In another terminal) Start the background worker
make go-worker
```

### Run Locally (TypeScript)

```bash
# 1. Start PostgreSQL
make db-up

# 2. Run migrations
make db-migrate

# 3. Generate protobuf code
make proto-gen

# 4. Install dependencies
make ts-install

# 5. Start the TypeScript server
make ts-server

# 6. (In another terminal) Start the background worker
make ts-worker
```

### Run API Tests

```bash
# Against Go server
make test-bruno

# Against TypeScript server
make test-bruno-ts
```

## Technology Decisions

| Concern | Choice | Why |
|---|---|---|
| API Protocol | [Connect RPC](https://connectrpc.com/) | gRPC + REST JSON on a single port |
| Schema | [Protobuf](https://protobuf.dev/) + [Buf](https://buf.build/) | Language-neutral contract, lint, breaking change detection |
| Database | PostgreSQL 16 | JSONB, pg_notify, partial indexes, CHECK constraints |
| Migrations | [dbmate](https://github.com/amacneil/dbmate) | Language-agnostic, plain SQL |
| Background Jobs (Go) | [River](https://riverqueue.com/) | Transactional enqueue in the same DB |
| Background Jobs (TS) | [graphile-worker](https://worker.graphile.org/) | PostgreSQL-based, no Redis dependency |
| API Tests | [Bruno](https://www.usebruno.com/) | Git-friendly, language-agnostic collections |

## Implementation Status

| Language | Status | Port |
|---|---|---|
| Go | Complete | 8080 |
| TypeScript | Complete | 8081 |
| Rust | Planned | 8082 |
| Kotlin | Planned | 8083 |
| C# | Planned | 8084 |
| Python | Planned | 8085 |

## Documentation

- [`api/`](api/README.md) — Protobuf definitions and Buf tooling
- [`db/`](db/README.md) — Database schema and migrations
- [`services/`](services/README.md) — Service implementations overview
- [`services/golang/`](services/golang/README.md) — Go reference implementation
- [`services/typescript/`](services/typescript/README.md) — TypeScript implementation
