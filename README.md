# api-corestack

A polyglot Task Management API built as a learning project to explore how the same API spec can be implemented across multiple programming languages — all sharing one PostgreSQL database, protobuf contract, and test suite.

Go is the reference implementation (complete). Other languages are planned.

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
   │  (complete) │     │  (planned)  │      │  (planned)  │
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
│   └── golang/             # Go reference implementation
├── tests/bruno/            # Bruno API test collections
├── docker-compose.yml
├── Makefile
└── .env.example
```

## Quick Start

### Prerequisites

- [Docker](https://www.docker.com/) and Docker Compose
- [Go 1.24+](https://go.dev/) (for local development)
- [Buf CLI](https://buf.build/docs/installation) (for protobuf generation)
- [Bruno](https://www.usebruno.com/) (optional, for API testing)

### Run with Docker (easiest)

```bash
# Start everything: PostgreSQL, migrations, Go server + worker
make docker-go
```

The API is available at `http://localhost:8080`.

### Run Locally

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

### Run API Tests

```bash
make test-bruno
```

## Technology Decisions

| Concern | Choice | Why |
|---|---|---|
| API Protocol | [Connect RPC](https://connectrpc.com/) | gRPC + REST JSON on a single port |
| Schema | [Protobuf](https://protobuf.dev/) + [Buf](https://buf.build/) | Language-neutral contract, lint, breaking change detection |
| Database | PostgreSQL 16 | JSONB, pg_notify, partial indexes, CHECK constraints |
| Migrations | [dbmate](https://github.com/amacneil/dbmate) | Language-agnostic, plain SQL |
| Background Jobs | [River](https://riverqueue.com/) | Transactional enqueue in the same DB |
| API Tests | [Bruno](https://www.usebruno.com/) | Git-friendly, language-agnostic collections |

## Implementation Status

| Language | Status | Port |
|---|---|---|
| Go | Complete | 8080 |
| TypeScript | Planned | 8081 |
| Rust | Planned | 8082 |
| Kotlin | Planned | 8083 |
| C# | Planned | 8084 |
| Python | Planned | 8085 |

## Documentation

- [`api/`](api/README.md) — Protobuf definitions and Buf tooling
- [`db/`](db/README.md) — Database schema and migrations
- [`services/`](services/README.md) — Service implementations overview
- [`services/golang/`](services/golang/README.md) — Go reference implementation
