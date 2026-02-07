-- name: CreateTask :one
INSERT INTO tasks (id, workspace_id, project_id, title, description, status, priority, assigned_to, due_date, metadata, created_at, updated_at)
VALUES (gen_random_uuid(), @workspace_id, @project_id, @title, @description, COALESCE(NULLIF(@status, ''), 'todo'), COALESCE(NULLIF(@priority, ''), 'medium'), NULLIF(@assigned_to, ''), @due_date, COALESCE(@metadata, '{}'::jsonb), NOW(), NOW())
RETURNING id, workspace_id, project_id, title, description, status, priority, assigned_to, due_date, metadata, created_at, updated_at, deleted_at;

-- name: GetTaskByID :one
SELECT id, workspace_id, project_id, title, description, status, priority, assigned_to, due_date, metadata, created_at, updated_at, deleted_at
FROM tasks
WHERE id = @id AND deleted_at IS NULL;

-- name: UpdateTask :one
UPDATE tasks
SET title = @title, description = @description, status = @status, priority = @priority,
    assigned_to = NULLIF(@assigned_to, ''), due_date = @due_date, metadata = COALESCE(@metadata, '{}'::jsonb), updated_at = NOW()
WHERE id = @id AND deleted_at IS NULL
RETURNING id, workspace_id, project_id, title, description, status, priority, assigned_to, due_date, metadata, created_at, updated_at, deleted_at;

-- name: SoftDeleteTask :exec
UPDATE tasks SET deleted_at = NOW(), updated_at = NOW()
WHERE id = @id AND deleted_at IS NULL;

-- name: CountTasks :one
SELECT COUNT(*)::int FROM tasks WHERE workspace_id = @workspace_id AND deleted_at IS NULL;
