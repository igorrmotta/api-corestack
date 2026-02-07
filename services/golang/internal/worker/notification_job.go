package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/riverqueue/river"

	"github.com/igorrmotta/api-corestack/services/golang/internal/domain"
)

// NotificationJobArgs represents the args for a notification processing job.
type NotificationJobArgs struct {
	NotificationID int64 `json:"notification_id"`
}

func (NotificationJobArgs) Kind() string { return "notification_process" }

// NotificationWorker processes notification queue entries.
type NotificationWorker struct {
	river.WorkerDefaults[NotificationJobArgs]
	notifRepo domain.NotificationRepository
}

func NewNotificationWorker(notifRepo domain.NotificationRepository) *NotificationWorker {
	return &NotificationWorker{notifRepo: notifRepo}
}

func (w *NotificationWorker) Work(ctx context.Context, job *river.Job[NotificationJobArgs]) error {
	slog.InfoContext(ctx, "processing notification",
		"notification_id", job.Args.NotificationID,
		"attempt", job.Attempt,
	)

	// Simulate sending notification (log output for now)
	// In production, this would call an email service, push notification service, etc.
	time.Sleep(100 * time.Millisecond)

	slog.InfoContext(ctx, "notification processed successfully",
		"notification_id", job.Args.NotificationID,
	)

	return nil
}

// NotificationBatchJobArgs processes a batch of pending notifications.
type NotificationBatchJobArgs struct {
	BatchSize int `json:"batch_size"`
}

func (NotificationBatchJobArgs) Kind() string { return "notification_batch" }

// NotificationBatchWorker fetches and processes pending notifications.
type NotificationBatchWorker struct {
	river.WorkerDefaults[NotificationBatchJobArgs]
	notifRepo domain.NotificationRepository
}

func NewNotificationBatchWorker(notifRepo domain.NotificationRepository) *NotificationBatchWorker {
	return &NotificationBatchWorker{notifRepo: notifRepo}
}

func (w *NotificationBatchWorker) Work(ctx context.Context, job *river.Job[NotificationBatchJobArgs]) error {
	batchSize := job.Args.BatchSize
	if batchSize <= 0 {
		batchSize = 10
	}

	slog.InfoContext(ctx, "fetching pending notifications", "batch_size", batchSize)

	notifications, err := w.notifRepo.FetchPending(ctx, batchSize)
	if err != nil {
		return fmt.Errorf("fetch pending notifications: %w", err)
	}

	if len(notifications) == 0 {
		slog.DebugContext(ctx, "no pending notifications")
		return nil
	}

	slog.InfoContext(ctx, "processing notification batch", "count", len(notifications))

	for _, n := range notifications {
		slog.InfoContext(ctx, "sending notification",
			"id", n.ID,
			"event_type", n.EventType,
			"workspace_id", n.WorkspaceID,
			"retry_count", n.RetryCount,
		)

		// Simulate sending — in production, dispatch based on event_type
		time.Sleep(50 * time.Millisecond)

		_, markErr := w.notifRepo.MarkProcessed(ctx, n.ID)
		if markErr != nil {
			slog.ErrorContext(ctx, "failed to mark notification processed",
				"id", n.ID,
				"error", markErr,
			)
			_ = w.notifRepo.MarkFailed(ctx, n.ID, markErr.Error())
			continue
		}

		slog.InfoContext(ctx, "notification sent", "id", n.ID)
	}

	return nil
}
