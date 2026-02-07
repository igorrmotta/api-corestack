package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/igorrmotta/api-corestack/services/golang/internal/domain"
)

type WorkspaceService struct {
	repo domain.WorkspaceRepository
}

func NewWorkspaceService(repo domain.WorkspaceRepository) *WorkspaceService {
	return &WorkspaceService{repo: repo}
}

func (s *WorkspaceService) Create(ctx context.Context, params domain.CreateWorkspaceParams) (*domain.Workspace, error) {
	if params.Name == "" {
		return nil, fmt.Errorf("%w: name is required", domain.ErrInvalidInput)
	}
	if params.Slug == "" {
		return nil, fmt.Errorf("%w: slug is required", domain.ErrInvalidInput)
	}
	slog.DebugContext(ctx, "creating workspace", "name", params.Name, "slug", params.Slug)
	return s.repo.Create(ctx, params)
}

func (s *WorkspaceService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Workspace, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *WorkspaceService) List(ctx context.Context, params domain.ListWorkspacesParams) (*domain.WorkspaceList, error) {
	return s.repo.List(ctx, params)
}

func (s *WorkspaceService) Update(ctx context.Context, params domain.UpdateWorkspaceParams) (*domain.Workspace, error) {
	if params.Name == "" {
		return nil, fmt.Errorf("%w: name is required", domain.ErrInvalidInput)
	}
	return s.repo.Update(ctx, params)
}

func (s *WorkspaceService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
