-- name: CreateNotification :one
INSERT INTO notification_queue (workspace_id, event_type, payload, status, created_at)
VALUES (@workspace_id, @event_type, @payload, 'pending', NOW())
RETURNING id, workspace_id, event_type, payload, status, retry_count, max_retries, next_retry_at, last_error, created_at, processed_at;

-- name: ListNotifications :many
SELECT id, workspace_id, event_type, payload, status, retry_count, max_retries, next_retry_at, last_error, created_at, processed_at
FROM notification_queue
WHERE workspace_id = @workspace_id
  AND (sqlc.narg('status_filter')::varchar IS NULL OR status = sqlc.narg('status_filter')::varchar)
ORDER BY created_at DESC
LIMIT @page_limit OFFSET @page_offset;

-- name: CountNotifications :one
SELECT COUNT(*)::int FROM notification_queue
WHERE workspace_id = @workspace_id
  AND (sqlc.narg('status_filter')::varchar IS NULL OR status = sqlc.narg('status_filter')::varchar);

-- name: MarkNotificationProcessed :one
UPDATE notification_queue
SET status = 'processed', processed_at = NOW()
WHERE id = @id
RETURNING id, workspace_id, event_type, payload, status, retry_count, max_retries, next_retry_at, last_error, created_at, processed_at;

-- name: FetchPendingNotifications :many
SELECT id, workspace_id, event_type, payload, status, retry_count, max_retries, next_retry_at, last_error, created_at, processed_at
FROM notification_queue
WHERE status IN ('pending', 'failed')
  AND (next_retry_at IS NULL OR next_retry_at <= NOW())
ORDER BY created_at ASC
LIMIT @fetch_limit
FOR UPDATE SKIP LOCKED;

-- name: MarkNotificationFailed :exec
UPDATE notification_queue
SET status = 'failed',
    last_error = @last_error,
    retry_count = retry_count + 1,
    next_retry_at = NOW() + (INTERVAL '1 second' * POWER(2, retry_count + 1))
WHERE id = @id;
