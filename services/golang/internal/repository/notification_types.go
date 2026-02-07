package repository

import (
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
