-- name: CreateWorkspace :one
INSERT INTO workspaces (id, name, slug, created_at, updated_at)
VALUES (gen_random_uuid(), @name, @slug, NOW(), NOW())
RETURNING id, name, slug, created_at, updated_at, deleted_at;

-- name: GetWorkspaceByID :one
SELECT id, name, slug, created_at, updated_at, deleted_at
FROM workspaces
WHERE id = @id AND deleted_at IS NULL;

-- name: ListWorkspaces :many
SELECT id, name, slug, created_at, updated_at, deleted_at
FROM workspaces
WHERE deleted_at IS NULL
  AND (sqlc.narg('cursor_id')::uuid IS NULL OR id < sqlc.narg('cursor_id')::uuid)
ORDER BY created_at DESC, id DESC
LIMIT @page_limit;

-- name: CountWorkspaces :one
SELECT COUNT(*)::int FROM workspaces WHERE deleted_at IS NULL;

-- name: UpdateWorkspace :one
UPDATE workspaces
SET name = @name, slug = @slug, updated_at = NOW()
WHERE id = @id AND deleted_at IS NULL
RETURNING id, name, slug, created_at, updated_at, deleted_at;

-- name: SoftDeleteWorkspace :exec
UPDATE workspaces SET deleted_at = NOW(), updated_at = NOW()
WHERE id = @id AND deleted_at IS NULL;
