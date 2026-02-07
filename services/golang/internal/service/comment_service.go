package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/igorrmotta/api-corestack/services/golang/internal/domain"
)

type CommentService struct {
	repo domain.CommentRepository
}

func NewCommentService(repo domain.CommentRepository) *CommentService {
	return &CommentService{repo: repo}
}

func (s *CommentService) Create(ctx context.Context, params domain.CreateCommentParams) (*domain.Comment, error) {
	if params.TaskID == uuid.Nil {
		return nil, fmt.Errorf("%w: task_id is required", domain.ErrInvalidInput)
	}
	if params.AuthorID == "" {
		return nil, fmt.Errorf("%w: author_id is required", domain.ErrInvalidInput)
	}
	if params.Content == "" {
		return nil, fmt.Errorf("%w: content is required", domain.ErrInvalidInput)
	}
	slog.DebugContext(ctx, "creating comment", "task_id", params.TaskID, "author_id", params.AuthorID)
	return s.repo.Create(ctx, params)
}

func (s *CommentService) List(ctx context.Context, params domain.ListCommentsParams) (*domain.CommentList, error) {
	return s.repo.List(ctx, params)
}

func (s *CommentService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
