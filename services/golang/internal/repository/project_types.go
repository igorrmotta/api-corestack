package repository

import (
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
