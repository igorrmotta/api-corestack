package repository

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	ProjectID   uuid.UUID
	Title       string
	Description string
	Status      string // todo, in_progress, review, done
	Priority    string // low, medium, high, critical
	AssignedTo  string
	DueDate     *time.Time
	Metadata    json.RawMessage
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

type CreateTaskParams struct {
	WorkspaceID uuid.UUID
	ProjectID   uuid.UUID
	Title       string
	Description string
	Priority    string
	AssignedTo  string
	DueDate     *time.Time
	Metadata    json.RawMessage
}

type UpdateTaskParams struct {
	ID          uuid.UUID
	Title       string
	Description string
	Status      string
	Priority    string
	AssignedTo  string
	DueDate     *time.Time
	Metadata    json.RawMessage
}

type ListTasksParams struct {
	WorkspaceID uuid.UUID
	ProjectID   uuid.UUID // optional filter
	Status      string    // optional filter
	Priority    string    // optional filter
	AssignedTo  string    // optional filter
	PageSize    int32
	PageToken   string
}

type TaskList struct {
	Tasks         []Task
	NextPageToken string
	TotalCount    int32
}

type TaskInput struct {
	Title       string
	Description string
	Priority    string
	AssignedTo  string
	DueDate     *time.Time
	Metadata    json.RawMessage
}

type ImportResult struct {
	Total     int32
	Succeeded int32
	Failed    int32
	Errors    []ImportError
}

type ImportError struct {
	Index int32
	Error string
}
