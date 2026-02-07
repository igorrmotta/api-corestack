package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/igorrmotta/api-corestack/services/golang/internal/domain"
)

type TaskService struct {
	repo      domain.TaskRepository
	notifRepo domain.NotificationRepository
}

func NewTaskService(repo domain.TaskRepository, notifRepo domain.NotificationRepository) *TaskService {
	return &TaskService{repo: repo, notifRepo: notifRepo}
}

func (s *TaskService) Create(ctx context.Context, params domain.CreateTaskParams) (*domain.Task, error) {
	if params.Title == "" {
		return nil, fmt.Errorf("%w: title is required", domain.ErrInvalidInput)
	}
	if params.WorkspaceID == uuid.Nil {
		return nil, fmt.Errorf("%w: workspace_id is required", domain.ErrInvalidInput)
	}
	if params.ProjectID == uuid.Nil {
		return nil, fmt.Errorf("%w: project_id is required", domain.ErrInvalidInput)
	}

	validPriorities := map[string]bool{"low": true, "medium": true, "high": true, "critical": true, "": true}
	if !validPriorities[params.Priority] {
		return nil, fmt.Errorf("%w: invalid priority: %s", domain.ErrInvalidInput, params.Priority)
	}

	slog.DebugContext(ctx, "creating task", "title", params.Title, "project_id", params.ProjectID)
	task, err := s.repo.Create(ctx, params)
	if err != nil {
		return nil, err
	}

	// Enqueue notification
	if s.notifRepo != nil {
		_, _ = s.notifRepo.Create(ctx, domain.CreateNotificationParams{
			WorkspaceID: task.WorkspaceID,
			EventType:   "task.created",
			Payload:     []byte(fmt.Sprintf(`{"task_id":"%s","title":"%s"}`, task.ID, task.Title)),
		})
	}

	return task, nil
}

func (s *TaskService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TaskService) List(ctx context.Context, params domain.ListTasksParams) (*domain.TaskList, error) {
	validStatuses := map[string]bool{"todo": true, "in_progress": true, "review": true, "done": true, "": true}
	if !validStatuses[params.Status] {
		return nil, fmt.Errorf("%w: invalid status filter: %s", domain.ErrInvalidInput, params.Status)
	}
	return s.repo.List(ctx, params)
}

func (s *TaskService) Update(ctx context.Context, params domain.UpdateTaskParams) (*domain.Task, error) {
	if params.Title == "" {
		return nil, fmt.Errorf("%w: title is required", domain.ErrInvalidInput)
	}

	validStatuses := map[string]bool{"todo": true, "in_progress": true, "review": true, "done": true, "": true}
	if !validStatuses[params.Status] {
		return nil, fmt.Errorf("%w: invalid status: %s", domain.ErrInvalidInput, params.Status)
	}

	return s.repo.Update(ctx, params)
}

func (s *TaskService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
