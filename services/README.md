# services/ — Language Implementations

Each subdirectory contains a complete implementation of the Task Management API in a different language. All implementations share the same protobuf contract (`api/`), database (`db/`), and test suite (`tests/bruno/`).

## The Polyglot Concept

The goal is to implement the exact same API in multiple languages to compare:
- Idioms and patterns across ecosystems
- Performance characteristics
- Developer experience and boilerplate
- Dependency landscape

Every implementation must be interchangeable — the Bruno test suite should pass against any of them.

## Common Architecture

All implementations follow the same layered structure:

```
Handlers (Connect RPC)  ←  translate proto ↔ domain, map errors
    │
Services (business logic)  ←  validation, orchestration
    │
Repositories (data access)  ←  SQL queries, pgx/driver
    │
PostgreSQL
```

## What Each Implementation Provides

- **Server binary** — Connect RPC server (gRPC + REST JSON on a single port)
- **Worker binary** — Background job processor for notifications and bulk imports
- **Dockerfile** — Multi-stage build producing minimal runtime image
- **Same endpoints** — All RPCs defined in `api/` must be implemented
- **Same port convention** — Each language gets a dedicated port (see below)

## Implementation Status

| Language | Directory | Port | Status |
|---|---|---|---|
| Go | [`golang/`](golang/README.md) | 8080 | Complete |
| TypeScript | `typescript/` | 8081 | Planned |
| Rust | `rust/` | 8082 | Planned |
| Kotlin | `kotlin/` | 8083 | Planned |
| C# | `csharp/` | 8084 | Planned |
| Python | `python/` | 8085 | Planned |
