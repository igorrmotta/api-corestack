package repository

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationRepo struct {
	pool *pgxpool.Pool
}

func NewNotificationRepo(pool *pgxpool.Pool) *NotificationRepo {
	return &NotificationRepo{pool: pool}
}

func (r *NotificationRepo) Create(ctx context.Context, params CreateNotificationParams) (*Notification, error) {
	var n Notification
	err := r.pool.QueryRow(ctx,
		`INSERT INTO notification_queue (workspace_id, event_type, payload, status, created_at)
		 VALUES ($1, $2, $3, 'pending', NOW())
		 RETURNING id, workspace_id, event_type, payload, status, retry_count, max_retries,
		           next_retry_at, COALESCE(last_error, ''), created_at, processed_at`,
		params.WorkspaceID, params.EventType, params.Payload,
	).Scan(&n.ID, &n.WorkspaceID, &n.EventType, &n.Payload, &n.Status, &n.RetryCount,
		&n.MaxRetries, &n.NextRetryAt, &n.LastError, &n.CreatedAt, &n.ProcessedAt)
	if err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}
	return &n, nil
}

func (r *NotificationRepo) List(ctx context.Context, params ListNotificationsParams) (*NotificationList, error) {
	pageSize := params.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	// Parse offset from page token (offset-based pagination for notifications)
	var offset int
	if params.PageToken != "" {
		var parseErr error
		offset, parseErr = strconv.Atoi(params.PageToken)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid page token: %w", parseErr)
		}
	}

	// Count query
	var totalCount int32
	if params.Status != "" {
		err := r.pool.QueryRow(ctx,
			`SELECT COUNT(*)::int FROM notification_queue
			 WHERE workspace_id = $1 AND status = $2`,
			params.WorkspaceID, params.Status,
		).Scan(&totalCount)
		if err != nil {
			return nil, fmt.Errorf("count notifications: %w", err)
		}
	} else {
		err := r.pool.QueryRow(ctx,
			`SELECT COUNT(*)::int FROM notification_queue WHERE workspace_id = $1`,
			params.WorkspaceID,
		).Scan(&totalCount)
		if err != nil {
			return nil, fmt.Errorf("count notifications: %w", err)
		}
	}

	// List query
	var rows pgx.Rows
	var err error
	if params.Status != "" {
		rows, err = r.pool.Query(ctx,
			`SELECT id, workspace_id, event_type, payload, status, retry_count, max_retries,
			        next_retry_at, COALESCE(last_error, ''), created_at, processed_at
			 FROM notification_queue
			 WHERE workspace_id = $1 AND status = $2
			 ORDER BY created_at DESC
			 LIMIT $3 OFFSET $4`,
			params.WorkspaceID, params.Status, pageSize, offset,
		)
	} else {
		rows, err = r.pool.Query(ctx,
			`SELECT id, workspace_id, event_type, payload, status, retry_count, max_retries,
			        next_retry_at, COALESCE(last_error, ''), created_at, processed_at
			 FROM notification_queue
			 WHERE workspace_id = $1
			 ORDER BY created_at DESC
			 LIMIT $2 OFFSET $3`,
			params.WorkspaceID, pageSize, offset,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.WorkspaceID, &n.EventType, &n.Payload, &n.Status,
			&n.RetryCount, &n.MaxRetries, &n.NextRetryAt, &n.LastError, &n.CreatedAt, &n.ProcessedAt); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		notifications = append(notifications, n)
	}

	var nextPageToken string
	nextOffset := offset + int(pageSize)
	if nextOffset < int(totalCount) {
		nextPageToken = strconv.Itoa(nextOffset)
	}

	return &NotificationList{
		Notifications: notifications,
		NextPageToken: nextPageToken,
		TotalCount:    totalCount,
	}, nil
}

func (r *NotificationRepo) MarkProcessed(ctx context.Context, id int64) (*Notification, error) {
	var n Notification
	err := r.pool.QueryRow(ctx,
		`UPDATE notification_queue
		 SET status = 'processed', processed_at = NOW()
		 WHERE id = $1
		 RETURNING id, workspace_id, event_type, payload, status, retry_count, max_retries,
		           next_retry_at, COALESCE(last_error, ''), created_at, processed_at`,
		id,
	).Scan(&n.ID, &n.WorkspaceID, &n.EventType, &n.Payload, &n.Status, &n.RetryCount,
		&n.MaxRetries, &n.NextRetryAt, &n.LastError, &n.CreatedAt, &n.ProcessedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("mark notification processed: %w", err)
	}
	return &n, nil
}

func (r *NotificationRepo) FetchPending(ctx context.Context, limit int) ([]Notification, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, workspace_id, event_type, payload, status, retry_count, max_retries,
		        next_retry_at, COALESCE(last_error, ''), created_at, processed_at
		 FROM notification_queue
		 WHERE status IN ('pending', 'failed')
		   AND (next_retry_at IS NULL OR next_retry_at <= NOW())
		 ORDER BY created_at ASC
		 LIMIT $1
		 FOR UPDATE SKIP LOCKED`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("fetch pending notifications: %w", err)
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.WorkspaceID, &n.EventType, &n.Payload, &n.Status,
			&n.RetryCount, &n.MaxRetries, &n.NextRetryAt, &n.LastError, &n.CreatedAt, &n.ProcessedAt); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		notifications = append(notifications, n)
	}

	return notifications, nil
}

func (r *NotificationRepo) MarkFailed(ctx context.Context, id int64, errMsg string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE notification_queue
		 SET status = 'failed',
		     last_error = $1,
		     retry_count = retry_count + 1,
		     next_retry_at = NOW() + (INTERVAL '1 second' * POWER(2, retry_count + 1))
		 WHERE id = $2`,
		errMsg, id,
	)
	if err != nil {
		return fmt.Errorf("mark notification failed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
