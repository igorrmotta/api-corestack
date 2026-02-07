package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/igorrmotta/api-corestack/services/golang/internal/repository"
)

type CommentService struct {
	repo *repository.CommentRepo
}

func NewCommentService(repo *repository.CommentRepo) *CommentService {
	return &CommentService{repo: repo}
}

func (s *CommentService) Create(ctx context.Context, params repository.CreateCommentParams) (*repository.Comment, error) {
	if params.TaskID == uuid.Nil {
		return nil, fmt.Errorf("%w: task_id is required", repository.ErrInvalidInput)
	}
	if params.AuthorID == "" {
		return nil, fmt.Errorf("%w: author_id is required", repository.ErrInvalidInput)
	}
	if params.Content == "" {
		return nil, fmt.Errorf("%w: content is required", repository.ErrInvalidInput)
	}
	slog.DebugContext(ctx, "creating comment", "task_id", params.TaskID, "author_id", params.AuthorID)
	return s.repo.Create(ctx, params)
}

func (s *CommentService) List(ctx context.Context, params repository.ListCommentsParams) (*repository.CommentList, error) {
	return s.repo.List(ctx, params)
}

func (s *CommentService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
