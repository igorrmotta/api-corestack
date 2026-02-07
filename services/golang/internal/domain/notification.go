package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID          int64
	WorkspaceID uuid.UUID
	EventType   string
	Payload     json.RawMessage
	Status      string // pending, processing, processed, failed
	RetryCount  int32
	MaxRetries  int32
	NextRetryAt *time.Time
	LastError   string
	CreatedAt   time.Time
	ProcessedAt *time.Time
}

type CreateNotificationParams struct {
	WorkspaceID uuid.UUID
	EventType   string
	Payload     json.RawMessage
}

type ListNotificationsParams struct {
	WorkspaceID uuid.UUID
	Status      string // optional filter
	PageSize    int32
	PageToken   string
}

type NotificationList struct {
	Notifications []Notification
	NextPageToken string
	TotalCount    int32
}

type NotificationRepository interface {
	Create(ctx context.Context, params CreateNotificationParams) (*Notification, error)
	List(ctx context.Context, params ListNotificationsParams) (*NotificationList, error)
	MarkProcessed(ctx context.Context, id int64) (*Notification, error)
	// FetchPending retrieves pending/failed notifications for processing (FOR UPDATE SKIP LOCKED)
	FetchPending(ctx context.Context, limit int) ([]Notification, error)
	// MarkFailed marks a notification as failed with error message and updates retry info
	MarkFailed(ctx context.Context, id int64, errMsg string) error
}
