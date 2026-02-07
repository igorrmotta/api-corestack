package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProjectRepo struct {
	pool *pgxpool.Pool
}

func NewProjectRepo(pool *pgxpool.Pool) *ProjectRepo {
	return &ProjectRepo{pool: pool}
}

func (r *ProjectRepo) Create(ctx context.Context, params CreateProjectParams) (*Project, error) {
	var p Project
	err := r.pool.QueryRow(ctx,
		`INSERT INTO projects (id, workspace_id, name, description, status, created_at, updated_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, 'active', NOW(), NOW())
		 RETURNING id, workspace_id, name, description, status, created_at, updated_at, deleted_at`,
		params.WorkspaceID, params.Name, params.Description,
	).Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Description, &p.Status, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	return &p, nil
}

func (r *ProjectRepo) GetByID(ctx context.Context, id uuid.UUID) (*Project, error) {
	var p Project
	err := r.pool.QueryRow(ctx,
		`SELECT id, workspace_id, name, description, status, created_at, updated_at, deleted_at
		 FROM projects WHERE id = $1 AND deleted_at IS NULL`,
		id,
	).Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Description, &p.Status, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get project: %w", err)
	}
	return &p, nil
}

func (r *ProjectRepo) List(ctx context.Context, params ListProjectsParams) (*ProjectList, error) {
	pageSize := params.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	var totalCount int32
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*)::int FROM projects WHERE workspace_id = $1 AND deleted_at IS NULL`,
		params.WorkspaceID,
	).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count projects: %w", err)
	}

	var rows pgx.Rows
	if params.PageToken != "" {
		cursorID, parseErr := uuid.Parse(params.PageToken)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid page token: %w", parseErr)
		}
		rows, err = r.pool.Query(ctx,
			`SELECT id, workspace_id, name, description, status, created_at, updated_at, deleted_at
			 FROM projects WHERE workspace_id = $1 AND deleted_at IS NULL AND id < $2
			 ORDER BY created_at DESC, id DESC LIMIT $3`,
			params.WorkspaceID, cursorID, pageSize+1,
		)
	} else {
		rows, err = r.pool.Query(ctx,
			`SELECT id, workspace_id, name, description, status, created_at, updated_at, deleted_at
			 FROM projects WHERE workspace_id = $1 AND deleted_at IS NULL
			 ORDER BY created_at DESC, id DESC LIMIT $2`,
			params.WorkspaceID, pageSize+1,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Description, &p.Status, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, p)
	}

	var nextPageToken string
	if len(projects) > int(pageSize) {
		nextPageToken = projects[pageSize].ID.String()
		projects = projects[:pageSize]
	}

	return &ProjectList{
		Projects:      projects,
		NextPageToken: nextPageToken,
		TotalCount:    totalCount,
	}, nil
}

func (r *ProjectRepo) Update(ctx context.Context, params UpdateProjectParams) (*Project, error) {
	var p Project
	err := r.pool.QueryRow(ctx,
		`UPDATE projects SET name = $1, description = $2, status = $3, updated_at = NOW()
		 WHERE id = $4 AND deleted_at IS NULL
		 RETURNING id, workspace_id, name, description, status, created_at, updated_at, deleted_at`,
		params.Name, params.Description, params.Status, params.ID,
	).Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Description, &p.Status, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update project: %w", err)
	}
	return &p, nil
}

func (r *ProjectRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE projects SET deleted_at = NOW(), updated_at = NOW()
		 WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
