# db/ — Database Schema & Migrations

PostgreSQL 16 schema managed by [dbmate](https://github.com/amacneil/dbmate). All migrations are plain SQL, shared across every language implementation.

## Commands

```bash
# Start PostgreSQL via Docker
make db-up

# Stop PostgreSQL
make db-down

# Run all pending migrations
make db-migrate

# Rollback the last migration
make db-rollback

# Create a new migration file
make db-new NAME=create_something
```

## Tables

| Table | Purpose | Key Columns |
|---|---|---|
| `workspaces` | Top-level tenant | `id`, `name`, `slug` |
| `projects` | Groups tasks within a workspace | `id`, `workspace_id`, `name`, `status` |
| `tasks` | Core work items | `id`, `project_id`, `title`, `status`, `priority`, `metadata` (JSONB) |
| `task_comments` | Discussion on tasks | `id`, `task_id`, `author_id`, `content` |
| `attachments` | File references on tasks | `id`, `task_id`, `file_url`, `file_size` |
| `notification_queue` | Async event processing | `id`, `workspace_id`, `event_type`, `payload` (JSONB), `status` |

## Entity Relationships

```
workspaces
  │
  ├── 1:N ── projects
  │            │
  │            └── 1:N ── tasks
  │                         │
  │                         ├── 1:N ── task_comments
  │                         └── 1:N ── attachments
  │
  └── 1:N ── notification_queue
```

## Design Decisions

**UUIDs as primary keys** — `gen_random_uuid()` for all entity tables. Avoids sequential ID leakage and works well in distributed setups.

**CHECK constraints over enums** — Status and priority fields use `CHECK (column IN (...))` instead of PostgreSQL `CREATE TYPE`. Easier to extend without migrations.

**Soft deletes** — `deleted_at TIMESTAMPTZ` on workspaces, projects, and tasks. Queries filter with `WHERE deleted_at IS NULL`.

**JSONB metadata** — Tasks have a `metadata JSONB DEFAULT '{}'` column for arbitrary key-value data without schema changes.

**pg_notify trigger** — The `tasks` table has an `AFTER INSERT OR UPDATE` trigger that sends real-time events via `pg_notify('task_events', ...)`. Useful for live dashboards or webhook dispatching.

**Notification queue** — `notification_queue` stores events with retry logic (`retry_count`, `max_retries`, `next_retry_at`). Processed by River workers using `FOR UPDATE SKIP LOCKED`.

## Index Strategy

| Index | Type | Purpose |
|---|---|---|
| `idx_*_not_deleted` | Partial (`WHERE deleted_at IS NULL`) | Fast lookups excluding soft-deleted rows |
| `idx_projects_workspace_id` | B-tree | FK lookup |
| `idx_tasks_workspace_id` | B-tree | FK lookup |
| `idx_tasks_project_id` | B-tree | FK lookup |
| `idx_tasks_status_created` | Composite | Filter by status, sort by created_at |
| `idx_tasks_assigned_to` | Partial (`WHERE assigned_to IS NOT NULL`) | Filter assigned tasks |
| `idx_task_comments_task_created` | Composite | Comments ordered by time per task |
| `idx_notification_queue_actionable` | Partial (`WHERE status IN (...)`) | Worker fetch of pending/failed items |

## River

[River](https://riverqueue.com/) is used for background job processing. It stores its own tables in the same database. River migrations must run before the worker starts — this happens automatically when the River client initializes with `riverpgxv5`.
