# api/ — Protobuf Definitions

Shared API contract for all language implementations. Uses [Buf v2](https://buf.build/) for linting, breaking change detection, and code generation.

## Buf Workspace

The workspace is defined in `buf.yaml` with 6 modules:

| Module | Path | Description |
|---|---|---|
| common | `common/v1/` | Shared types: `SortOrder`, `PaginationRequest`, `PaginationResponse` |
| workspace | `workspace/v1/` | Workspace CRUD |
| project | `project/v1/` | Project CRUD (scoped to workspace) |
| task | `task/v1/` | Task CRUD + bulk import |
| comment | `comment/v1/` | Task comments |
| notification | `notification/v1/` | Notification listing and acknowledgment |

## Services and RPCs

### WorkspaceService

| RPC | Description |
|---|---|
| `CreateWorkspace` | Create a new workspace |
| `GetWorkspace` | Get workspace by ID |
| `ListWorkspaces` | Paginated list |
| `UpdateWorkspace` | Update name/slug |
| `DeleteWorkspace` | Soft delete |

### ProjectService

| RPC | Description |
|---|---|
| `CreateProject` | Create project in a workspace |
| `GetProject` | Get project by ID |
| `ListProjects` | Paginated list filtered by workspace |
| `UpdateProject` | Update name/description/status |
| `DeleteProject` | Soft delete |

### TaskService

| RPC | Description |
|---|---|
| `CreateTask` | Create task in a project |
| `GetTask` | Get task by ID |
| `ListTasks` | Paginated list with filters (status, priority, assigned_to) |
| `UpdateTask` | Update any task field |
| `DeleteTask` | Soft delete |
| `BulkImportTasks` | Import multiple tasks with error reporting per item |

### CommentService

| RPC | Description |
|---|---|
| `CreateComment` | Add comment to a task |
| `ListComments` | Paginated comments for a task |
| `DeleteComment` | Delete a comment |

### NotificationService

| RPC | Description |
|---|---|
| `ListNotifications` | Paginated list filtered by workspace/status |
| `MarkNotificationRead` | Mark a notification as processed |

## Shared Types

**PaginationRequest** — cursor-based pagination:
- `page_size` (int32) — items per page
- `page_token` (string) — opaque cursor (UUID of last item)

**PaginationResponse**:
- `next_page_token` (string) — cursor for next page, empty if no more
- `total_count` (int32)

**SortOrder** — `SORT_ORDER_UNSPECIFIED`, `SORT_ORDER_ASC`, `SORT_ORDER_DESC`

## Commands

```bash
# Generate code (currently Go only)
make proto-gen

# Lint protobuf files
make proto-lint

# Check for breaking changes against main branch
make proto-breaking
```

## Prerequisites

- [Buf CLI](https://buf.build/docs/installation)

## Code Generation

`buf.gen.yaml` configures which plugins generate code and where output goes:

```yaml
version: v2
plugins:
  - remote: buf.build/protocolbuffers/go
    out: ../services/golang/gen
    opt: paths=source_relative
  - remote: buf.build/connectrpc/go
    out: ../services/golang/gen
    opt: paths=source_relative
```

Currently generates Go code only. Generated files go to `services/golang/gen/` (gitignored).

## Adding Code Generation for a New Language

1. Add a plugin entry to `buf.gen.yaml` with the appropriate remote plugin and output path
2. Set the output to `../services/<language>/gen` (or equivalent)
3. Run `make proto-gen`
4. Add the generated directory to `.gitignore` if desired
