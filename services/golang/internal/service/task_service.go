package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/igorrmotta/api-corestack/services/golang/internal/repository"
)

type TaskService struct {
	repo      *repository.TaskRepo
	notifRepo *repository.NotificationRepo
}

func NewTaskService(repo *repository.TaskRepo, notifRepo *repository.NotificationRepo) *TaskService {
	return &TaskService{repo: repo, notifRepo: notifRepo}
}

func (s *TaskService) Create(ctx context.Context, params repository.CreateTaskParams) (*repository.Task, error) {
	if params.Title == "" {
		return nil, fmt.Errorf("%w: title is required", repository.ErrInvalidInput)
	}
	if params.WorkspaceID == uuid.Nil {
		return nil, fmt.Errorf("%w: workspace_id is required", repository.ErrInvalidInput)
	}
	if params.ProjectID == uuid.Nil {
		return nil, fmt.Errorf("%w: project_id is required", repository.ErrInvalidInput)
	}

	validPriorities := map[string]bool{"low": true, "medium": true, "high": true, "critical": true, "": true}
	if !validPriorities[params.Priority] {
		return nil, fmt.Errorf("%w: invalid priority: %s", repository.ErrInvalidInput, params.Priority)
	}

	slog.DebugContext(ctx, "creating task", "title", params.Title, "project_id", params.ProjectID)
	task, err := s.repo.Create(ctx, params)
	if err != nil {
		return nil, err
	}

	// Enqueue notification
	if s.notifRepo != nil {
		_, _ = s.notifRepo.Create(ctx, repository.CreateNotificationParams{
			WorkspaceID: task.WorkspaceID,
			EventType:   "task.created",
			Payload:     []byte(fmt.Sprintf(`{"task_id":"%s","title":"%s"}`, task.ID, task.Title)),
		})
	}

	return task, nil
}

func (s *TaskService) GetByID(ctx context.Context, id uuid.UUID) (*repository.Task, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TaskService) List(ctx context.Context, params repository.ListTasksParams) (*repository.TaskList, error) {
	validStatuses := map[string]bool{"todo": true, "in_progress": true, "review": true, "done": true, "": true}
	if !validStatuses[params.Status] {
		return nil, fmt.Errorf("%w: invalid status filter: %s", repository.ErrInvalidInput, params.Status)
	}
	return s.repo.List(ctx, params)
}

func (s *TaskService) Update(ctx context.Context, params repository.UpdateTaskParams) (*repository.Task, error) {
	if params.Title == "" {
		return nil, fmt.Errorf("%w: title is required", repository.ErrInvalidInput)
	}

	validStatuses := map[string]bool{"todo": true, "in_progress": true, "review": true, "done": true, "": true}
	if !validStatuses[params.Status] {
		return nil, fmt.Errorf("%w: invalid status: %s", repository.ErrInvalidInput, params.Status)
	}

	return s.repo.Update(ctx, params)
}

func (s *TaskService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
