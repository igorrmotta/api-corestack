package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CommentRepo struct {
	pool *pgxpool.Pool
}

func NewCommentRepo(pool *pgxpool.Pool) *CommentRepo {
	return &CommentRepo{pool: pool}
}

func (r *CommentRepo) Create(ctx context.Context, params CreateCommentParams) (*Comment, error) {
	var c Comment
	err := r.pool.QueryRow(ctx,
		`INSERT INTO task_comments (id, task_id, author_id, content, created_at, updated_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, NOW(), NOW())
		 RETURNING id, task_id, author_id, content, created_at, updated_at`,
		params.TaskID, params.AuthorID, params.Content,
	).Scan(&c.ID, &c.TaskID, &c.AuthorID, &c.Content, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create comment: %w", err)
	}
	return &c, nil
}

func (r *CommentRepo) List(ctx context.Context, params ListCommentsParams) (*CommentList, error) {
	pageSize := params.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	var totalCount int32
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*)::int FROM task_comments WHERE task_id = $1`,
		params.TaskID,
	).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count comments: %w", err)
	}

	var rows pgx.Rows
	if params.PageToken != "" {
		cursorID, parseErr := uuid.Parse(params.PageToken)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid page token: %w", parseErr)
		}
		rows, err = r.pool.Query(ctx,
			`SELECT id, task_id, author_id, content, created_at, updated_at
			 FROM task_comments WHERE task_id = $1 AND id < $2
			 ORDER BY created_at DESC, id DESC LIMIT $3`,
			params.TaskID, cursorID, pageSize+1,
		)
	} else {
		rows, err = r.pool.Query(ctx,
			`SELECT id, task_id, author_id, content, created_at, updated_at
			 FROM task_comments WHERE task_id = $1
			 ORDER BY created_at DESC, id DESC LIMIT $2`,
			params.TaskID, pageSize+1,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		if err := rows.Scan(&c.ID, &c.TaskID, &c.AuthorID, &c.Content, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}
		comments = append(comments, c)
	}

	var nextPageToken string
	if len(comments) > int(pageSize) {
		nextPageToken = comments[pageSize].ID.String()
		comments = comments[:pageSize]
	}

	return &CommentList{
		Comments:      comments,
		NextPageToken: nextPageToken,
		TotalCount:    totalCount,
	}, nil
}

func (r *CommentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM task_comments WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
