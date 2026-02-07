package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	Name        string
	Description string
	Status      string // active, archived
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

type CreateProjectParams struct {
	WorkspaceID uuid.UUID
	Name        string
	Description string
}

type UpdateProjectParams struct {
	ID          uuid.UUID
	Name        string
	Description string
	Status      string
}

type ListProjectsParams struct {
	WorkspaceID uuid.UUID
	PageSize    int32
	PageToken   string
}

type ProjectList struct {
	Projects      []Project
	NextPageToken string
	TotalCount    int32
}

type ProjectRepository interface {
	Create(ctx context.Context, params CreateProjectParams) (*Project, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Project, error)
	List(ctx context.Context, params ListProjectsParams) (*ProjectList, error)
	Update(ctx context.Context, params UpdateProjectParams) (*Project, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
