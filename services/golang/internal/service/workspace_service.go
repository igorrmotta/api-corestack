package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/igorrmotta/api-corestack/services/golang/internal/repository"
)

type WorkspaceService struct {
	repo *repository.WorkspaceRepo
}

func NewWorkspaceService(repo *repository.WorkspaceRepo) *WorkspaceService {
	return &WorkspaceService{repo: repo}
}

func (s *WorkspaceService) Create(ctx context.Context, params repository.CreateWorkspaceParams) (*repository.Workspace, error) {
	if params.Name == "" {
		return nil, fmt.Errorf("%w: name is required", repository.ErrInvalidInput)
	}
	if params.Slug == "" {
		return nil, fmt.Errorf("%w: slug is required", repository.ErrInvalidInput)
	}
	slog.DebugContext(ctx, "creating workspace", "name", params.Name, "slug", params.Slug)
	return s.repo.Create(ctx, params)
}

func (s *WorkspaceService) GetByID(ctx context.Context, id uuid.UUID) (*repository.Workspace, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *WorkspaceService) List(ctx context.Context, params repository.ListWorkspacesParams) (*repository.WorkspaceList, error) {
	return s.repo.List(ctx, params)
}

func (s *WorkspaceService) Update(ctx context.Context, params repository.UpdateWorkspaceParams) (*repository.Workspace, error) {
	if params.Name == "" {
		return nil, fmt.Errorf("%w: name is required", repository.ErrInvalidInput)
	}
	return s.repo.Update(ctx, params)
}

func (s *WorkspaceService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
