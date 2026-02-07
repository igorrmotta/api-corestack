-- name: CreateProject :one
INSERT INTO projects (id, workspace_id, name, description, status, created_at, updated_at)
VALUES (gen_random_uuid(), @workspace_id, @name, @description, 'active', NOW(), NOW())
RETURNING id, workspace_id, name, description, status, created_at, updated_at, deleted_at;

-- name: GetProjectByID :one
SELECT id, workspace_id, name, description, status, created_at, updated_at, deleted_at
FROM projects
WHERE id = @id AND deleted_at IS NULL;

-- name: ListProjects :many
SELECT id, workspace_id, name, description, status, created_at, updated_at, deleted_at
FROM projects
WHERE workspace_id = @workspace_id AND deleted_at IS NULL
  AND (sqlc.narg('cursor_id')::uuid IS NULL OR id < sqlc.narg('cursor_id')::uuid)
ORDER BY created_at DESC, id DESC
LIMIT @page_limit;

-- name: CountProjects :one
SELECT COUNT(*)::int FROM projects WHERE workspace_id = @workspace_id AND deleted_at IS NULL;

-- name: UpdateProject :one
UPDATE projects
SET name = @name, description = @description, status = @status, updated_at = NOW()
WHERE id = @id AND deleted_at IS NULL
RETURNING id, workspace_id, name, description, status, created_at, updated_at, deleted_at;

-- name: SoftDeleteProject :exec
UPDATE projects SET deleted_at = NOW(), updated_at = NOW()
WHERE id = @id AND deleted_at IS NULL;
