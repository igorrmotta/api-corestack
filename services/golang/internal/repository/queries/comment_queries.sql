-- name: CreateComment :one
INSERT INTO task_comments (id, task_id, author_id, content, created_at, updated_at)
VALUES (gen_random_uuid(), @task_id, @author_id, @content, NOW(), NOW())
RETURNING id, task_id, author_id, content, created_at, updated_at;

-- name: ListComments :many
SELECT id, task_id, author_id, content, created_at, updated_at
FROM task_comments
WHERE task_id = @task_id
  AND (sqlc.narg('cursor_id')::uuid IS NULL OR id < sqlc.narg('cursor_id')::uuid)
ORDER BY created_at DESC, id DESC
LIMIT @page_limit;

-- name: CountComments :one
SELECT COUNT(*)::int FROM task_comments WHERE task_id = @task_id;

-- name: DeleteComment :exec
DELETE FROM task_comments WHERE id = @id;
