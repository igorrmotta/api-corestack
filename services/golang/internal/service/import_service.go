package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"

	"github.com/igorrmotta/api-corestack/services/golang/internal/domain"
)

type ImportService struct {
	taskRepo    domain.TaskRepository
	notifRepo   domain.NotificationRepository
	concurrency int
	rateLimit   rate.Limit
}

func NewImportService(taskRepo domain.TaskRepository, notifRepo domain.NotificationRepository, concurrency int, rateLimit float64) *ImportService {
	if concurrency <= 0 {
		concurrency = 10
	}
	if rateLimit <= 0 {
		rateLimit = 100
	}
	return &ImportService{
		taskRepo:    taskRepo,
		notifRepo:   notifRepo,
		concurrency: concurrency,
		rateLimit:   rate.Limit(rateLimit),
	}
}

func (s *ImportService) BulkImport(ctx context.Context, workspaceID, projectID uuid.UUID, inputs []domain.TaskInput) *domain.ImportResult {
	result := &domain.ImportResult{
		Total: int32(len(inputs)),
	}

	if len(inputs) == 0 {
		return result
	}

	slog.InfoContext(ctx, "starting bulk import",
		"workspace_id", workspaceID,
		"project_id", projectID,
		"total", len(inputs),
		"concurrency", s.concurrency,
	)

	var mu sync.Mutex
	limiter := rate.NewLimiter(s.rateLimit, s.concurrency)

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(s.concurrency)

	for i, input := range inputs {
		i, input := i, input // capture loop vars
		g.Go(func() error {
			if err := limiter.Wait(ctx); err != nil {
				mu.Lock()
				result.Failed++
				result.Errors = append(result.Errors, domain.ImportError{
					Index: int32(i),
					Error: fmt.Sprintf("rate limit: %v", err),
				})
				mu.Unlock()
				return nil // don't cancel other goroutines
			}

			task, err := s.taskRepo.Create(ctx, domain.CreateTaskParams{
				WorkspaceID: workspaceID,
				ProjectID:   projectID,
				Title:       input.Title,
				Description: input.Description,
				Priority:    input.Priority,
				AssignedTo:  input.AssignedTo,
				DueDate:     input.DueDate,
				Metadata:    input.Metadata,
			})
			if err != nil {
				mu.Lock()
				result.Failed++
				result.Errors = append(result.Errors, domain.ImportError{
					Index: int32(i),
					Error: err.Error(),
				})
				mu.Unlock()
				return nil
			}

			// Enqueue notification for each imported task
			if s.notifRepo != nil {
				_, _ = s.notifRepo.Create(ctx, domain.CreateNotificationParams{
					WorkspaceID: workspaceID,
					EventType:   "task.imported",
					Payload:     []byte(fmt.Sprintf(`{"task_id":"%s","title":"%s"}`, task.ID, task.Title)),
				})
			}

			mu.Lock()
			result.Succeeded++
			mu.Unlock()

			return nil
		})
	}

	_ = g.Wait()

	slog.InfoContext(ctx, "bulk import completed",
		"total", result.Total,
		"succeeded", result.Succeeded,
		"failed", result.Failed,
	)

	return result
}
