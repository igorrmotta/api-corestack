package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Workspace struct {
	ID        uuid.UUID
	Name      string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type CreateWorkspaceParams struct {
	Name string
	Slug string
}

type UpdateWorkspaceParams struct {
	ID   uuid.UUID
	Name string
	Slug string
}

type ListWorkspacesParams struct {
	PageSize  int32
	PageToken string // cursor: UUID of last item
}

type WorkspaceList struct {
	Workspaces    []Workspace
	NextPageToken string
	TotalCount    int32
}

type WorkspaceRepository interface {
	Create(ctx context.Context, params CreateWorkspaceParams) (*Workspace, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Workspace, error)
	List(ctx context.Context, params ListWorkspacesParams) (*WorkspaceList, error)
	Update(ctx context.Context, params UpdateWorkspaceParams) (*Workspace, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
