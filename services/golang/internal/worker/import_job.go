package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/riverqueue/river"

	"github.com/igorrmotta/api-corestack/services/golang/internal/repository"
)

// ImportJobArgs represents the args for a bulk import job.
type ImportJobArgs struct {
	WorkspaceID string            `json:"workspace_id"`
	ProjectID   string            `json:"project_id"`
	Tasks       []ImportTaskInput `json:"tasks"`
}

func (ImportJobArgs) Kind() string { return "bulk_import" }

type ImportTaskInput struct {
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Priority    string          `json:"priority"`
	AssignedTo  string          `json:"assigned_to"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
}

// ImportWorker processes bulk import jobs asynchronously.
type ImportWorker struct {
	river.WorkerDefaults[ImportJobArgs]
	taskRepo  *repository.TaskRepo
	notifRepo *repository.NotificationRepo
}

func NewImportWorker(taskRepo *repository.TaskRepo, notifRepo *repository.NotificationRepo) *ImportWorker {
	return &ImportWorker{
		taskRepo:  taskRepo,
		notifRepo: notifRepo,
	}
}

func (w *ImportWorker) Work(ctx context.Context, job *river.Job[ImportJobArgs]) error {
	workspaceID, err := uuid.Parse(job.Args.WorkspaceID)
	if err != nil {
		return fmt.Errorf("invalid workspace_id: %w", err)
	}
	projectID, err := uuid.Parse(job.Args.ProjectID)
	if err != nil {
		return fmt.Errorf("invalid project_id: %w", err)
	}

	slog.InfoContext(ctx, "starting async bulk import",
		"workspace_id", workspaceID,
		"project_id", projectID,
		"total_tasks", len(job.Args.Tasks),
	)

	var succeeded, failed int
	for i, input := range job.Args.Tasks {
		task, createErr := w.taskRepo.Create(ctx, repository.CreateTaskParams{
			WorkspaceID: workspaceID,
			ProjectID:   projectID,
			Title:       input.Title,
			Description: input.Description,
			Priority:    input.Priority,
			AssignedTo:  input.AssignedTo,
			Metadata:    input.Metadata,
		})
		if createErr != nil {
			failed++
			slog.WarnContext(ctx, "import task failed",
				"index", i,
				"error", createErr,
			)
			continue
		}

		succeeded++

		// Enqueue notification for imported task
		if w.notifRepo != nil {
			_, _ = w.notifRepo.Create(ctx, repository.CreateNotificationParams{
				WorkspaceID: workspaceID,
				EventType:   "task.imported",
				Payload:     []byte(fmt.Sprintf(`{"task_id":"%s","title":"%s"}`, task.ID, task.Title)),
			})
		}
	}

	slog.InfoContext(ctx, "async bulk import completed",
		"total", len(job.Args.Tasks),
		"succeeded", succeeded,
		"failed", failed,
	)

	return nil
}
