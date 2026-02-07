package repository

import (
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
