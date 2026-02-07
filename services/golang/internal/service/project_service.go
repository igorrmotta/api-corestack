package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/igorrmotta/api-corestack/services/golang/internal/repository"
)

type ProjectService struct {
	repo *repository.ProjectRepo
}

func NewProjectService(repo *repository.ProjectRepo) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) Create(ctx context.Context, params repository.CreateProjectParams) (*repository.Project, error) {
	if params.Name == "" {
		return nil, fmt.Errorf("%w: name is required", repository.ErrInvalidInput)
	}
	if params.WorkspaceID == uuid.Nil {
		return nil, fmt.Errorf("%w: workspace_id is required", repository.ErrInvalidInput)
	}
	slog.DebugContext(ctx, "creating project", "name", params.Name, "workspace_id", params.WorkspaceID)
	return s.repo.Create(ctx, params)
}

func (s *ProjectService) GetByID(ctx context.Context, id uuid.UUID) (*repository.Project, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProjectService) List(ctx context.Context, params repository.ListProjectsParams) (*repository.ProjectList, error) {
	return s.repo.List(ctx, params)
}

func (s *ProjectService) Update(ctx context.Context, params repository.UpdateProjectParams) (*repository.Project, error) {
	if params.Name == "" {
		return nil, fmt.Errorf("%w: name is required", repository.ErrInvalidInput)
	}
	validStatuses := map[string]bool{"active": true, "archived": true}
	if params.Status != "" && !validStatuses[params.Status] {
		return nil, fmt.Errorf("%w: invalid status: %s", repository.ErrInvalidInput, params.Status)
	}
	return s.repo.Update(ctx, params)
}

func (s *ProjectService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
