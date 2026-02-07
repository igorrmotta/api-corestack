package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/igorrmotta/api-corestack/services/golang/internal/domain"
)

type WorkspaceRepo struct {
	pool *pgxpool.Pool
}

func NewWorkspaceRepo(pool *pgxpool.Pool) *WorkspaceRepo {
	return &WorkspaceRepo{pool: pool}
}

func (r *WorkspaceRepo) Create(ctx context.Context, params domain.CreateWorkspaceParams) (*domain.Workspace, error) {
	var w domain.Workspace
	err := r.pool.QueryRow(ctx,
		`INSERT INTO workspaces (id, name, slug, created_at, updated_at)
		 VALUES (gen_random_uuid(), $1, $2, NOW(), NOW())
		 RETURNING id, name, slug, created_at, updated_at, deleted_at`,
		params.Name, params.Slug,
	).Scan(&w.ID, &w.Name, &w.Slug, &w.CreatedAt, &w.UpdatedAt, &w.DeletedAt)
	if err != nil {
		return nil, fmt.Errorf("create workspace: %w", err)
	}
	return &w, nil
}

func (r *WorkspaceRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Workspace, error) {
	var w domain.Workspace
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, slug, created_at, updated_at, deleted_at
		 FROM workspaces WHERE id = $1 AND deleted_at IS NULL`,
		id,
	).Scan(&w.ID, &w.Name, &w.Slug, &w.CreatedAt, &w.UpdatedAt, &w.DeletedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get workspace: %w", err)
	}
	return &w, nil
}

func (r *WorkspaceRepo) List(ctx context.Context, params domain.ListWorkspacesParams) (*domain.WorkspaceList, error) {
	pageSize := params.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	var totalCount int32
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM workspaces WHERE deleted_at IS NULL`).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count workspaces: %w", err)
	}

	var rows pgx.Rows
	if params.PageToken != "" {
		cursorID, parseErr := uuid.Parse(params.PageToken)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid page token: %w", parseErr)
		}
		rows, err = r.pool.Query(ctx,
			`SELECT id, name, slug, created_at, updated_at, deleted_at
			 FROM workspaces WHERE deleted_at IS NULL AND id < $1
			 ORDER BY created_at DESC, id DESC LIMIT $2`,
			cursorID, pageSize+1,
		)
	} else {
		rows, err = r.pool.Query(ctx,
			`SELECT id, name, slug, created_at, updated_at, deleted_at
			 FROM workspaces WHERE deleted_at IS NULL
			 ORDER BY created_at DESC, id DESC LIMIT $1`,
			pageSize+1,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("list workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []domain.Workspace
	for rows.Next() {
		var w domain.Workspace
		if err := rows.Scan(&w.ID, &w.Name, &w.Slug, &w.CreatedAt, &w.UpdatedAt, &w.DeletedAt); err != nil {
			return nil, fmt.Errorf("scan workspace: %w", err)
		}
		workspaces = append(workspaces, w)
	}

	var nextPageToken string
	if len(workspaces) > int(pageSize) {
		nextPageToken = workspaces[pageSize].ID.String()
		workspaces = workspaces[:pageSize]
	}

	return &domain.WorkspaceList{
		Workspaces:    workspaces,
		NextPageToken: nextPageToken,
		TotalCount:    totalCount,
	}, nil
}

func (r *WorkspaceRepo) Update(ctx context.Context, params domain.UpdateWorkspaceParams) (*domain.Workspace, error) {
	var w domain.Workspace
	err := r.pool.QueryRow(ctx,
		`UPDATE workspaces SET name = $1, slug = $2, updated_at = NOW()
		 WHERE id = $3 AND deleted_at IS NULL
		 RETURNING id, name, slug, created_at, updated_at, deleted_at`,
		params.Name, params.Slug, params.ID,
	).Scan(&w.ID, &w.Name, &w.Slug, &w.CreatedAt, &w.UpdatedAt, &w.DeletedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("update workspace: %w", err)
	}
	return &w, nil
}

func (r *WorkspaceRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE workspaces SET deleted_at = NOW(), updated_at = NOW()
		 WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("delete workspace: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
