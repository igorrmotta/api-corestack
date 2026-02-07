package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRepo struct {
	pool *pgxpool.Pool
}

func NewTaskRepo(pool *pgxpool.Pool) *TaskRepo {
	return &TaskRepo{pool: pool}
}

func (r *TaskRepo) Create(ctx context.Context, params CreateTaskParams) (*Task, error) {
	var t Task
	var assignedTo *string
	if params.AssignedTo != "" {
		assignedTo = &params.AssignedTo
	}
	metadata := params.Metadata
	if metadata == nil {
		metadata = []byte("{}")
	}

	err := r.pool.QueryRow(ctx,
		`INSERT INTO tasks (id, workspace_id, project_id, title, description, status, priority, assigned_to, due_date, metadata, created_at, updated_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4,
		         COALESCE(NULLIF($5, ''), 'todo'),
		         COALESCE(NULLIF($6, ''), 'medium'),
		         $7, $8, $9, NOW(), NOW())
		 RETURNING id, workspace_id, project_id, title, description, status, priority,
		           COALESCE(assigned_to, ''), due_date, metadata, created_at, updated_at, deleted_at`,
		params.WorkspaceID, params.ProjectID, params.Title, params.Description,
		"", params.Priority, assignedTo, params.DueDate, metadata,
	).Scan(&t.ID, &t.WorkspaceID, &t.ProjectID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&t.AssignedTo, &t.DueDate, &t.Metadata, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt)
	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	return &t, nil
}

func (r *TaskRepo) GetByID(ctx context.Context, id uuid.UUID) (*Task, error) {
	var t Task
	err := r.pool.QueryRow(ctx,
		`SELECT id, workspace_id, project_id, title, description, status, priority,
		        COALESCE(assigned_to, ''), due_date, metadata, created_at, updated_at, deleted_at
		 FROM tasks WHERE id = $1 AND deleted_at IS NULL`,
		id,
	).Scan(&t.ID, &t.WorkspaceID, &t.ProjectID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&t.AssignedTo, &t.DueDate, &t.Metadata, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get task: %w", err)
	}
	return &t, nil
}

func (r *TaskRepo) List(ctx context.Context, params ListTasksParams) (*TaskList, error) {
	pageSize := params.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	// Build dynamic WHERE clause
	conditions := []string{"workspace_id = $1", "deleted_at IS NULL"}
	args := []any{params.WorkspaceID}
	argIdx := 2

	if params.ProjectID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("project_id = $%d", argIdx))
		args = append(args, params.ProjectID)
		argIdx++
	}
	if params.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, params.Status)
		argIdx++
	}
	if params.Priority != "" {
		conditions = append(conditions, fmt.Sprintf("priority = $%d", argIdx))
		args = append(args, params.Priority)
		argIdx++
	}
	if params.AssignedTo != "" {
		conditions = append(conditions, fmt.Sprintf("assigned_to = $%d", argIdx))
		args = append(args, params.AssignedTo)
		argIdx++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count query
	var totalCount int32
	err := r.pool.QueryRow(ctx,
		fmt.Sprintf("SELECT COUNT(*)::int FROM tasks WHERE %s", whereClause),
		args...,
	).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count tasks: %w", err)
	}

	// Add cursor pagination
	if params.PageToken != "" {
		cursorID, parseErr := uuid.Parse(params.PageToken)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid page token: %w", parseErr)
		}
		conditions = append(conditions, fmt.Sprintf("id < $%d", argIdx))
		args = append(args, cursorID)
		argIdx++
		whereClause = strings.Join(conditions, " AND ")
	}

	// Add limit
	args = append(args, pageSize+1)

	query := fmt.Sprintf(
		`SELECT id, workspace_id, project_id, title, description, status, priority,
		        COALESCE(assigned_to, ''), due_date, metadata, created_at, updated_at, deleted_at
		 FROM tasks WHERE %s
		 ORDER BY created_at DESC, id DESC LIMIT $%d`,
		whereClause, argIdx,
	)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.WorkspaceID, &t.ProjectID, &t.Title, &t.Description,
			&t.Status, &t.Priority, &t.AssignedTo, &t.DueDate, &t.Metadata,
			&t.CreatedAt, &t.UpdatedAt, &t.DeletedAt); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}

	var nextPageToken string
	if len(tasks) > int(pageSize) {
		nextPageToken = tasks[pageSize].ID.String()
		tasks = tasks[:pageSize]
	}

	return &TaskList{
		Tasks:         tasks,
		NextPageToken: nextPageToken,
		TotalCount:    totalCount,
	}, nil
}

func (r *TaskRepo) Update(ctx context.Context, params UpdateTaskParams) (*Task, error) {
	var t Task
	var assignedTo *string
	if params.AssignedTo != "" {
		assignedTo = &params.AssignedTo
	}
	metadata := params.Metadata
	if metadata == nil {
		metadata = []byte("{}")
	}

	err := r.pool.QueryRow(ctx,
		`UPDATE tasks SET title = $1, description = $2, status = $3, priority = $4,
		        assigned_to = $5, due_date = $6, metadata = $7, updated_at = NOW()
		 WHERE id = $8 AND deleted_at IS NULL
		 RETURNING id, workspace_id, project_id, title, description, status, priority,
		           COALESCE(assigned_to, ''), due_date, metadata, created_at, updated_at, deleted_at`,
		params.Title, params.Description, params.Status, params.Priority,
		assignedTo, params.DueDate, metadata, params.ID,
	).Scan(&t.ID, &t.WorkspaceID, &t.ProjectID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&t.AssignedTo, &t.DueDate, &t.Metadata, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update task: %w", err)
	}
	return &t, nil
}

func (r *TaskRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE tasks SET deleted_at = NOW(), updated_at = NOW()
		 WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
