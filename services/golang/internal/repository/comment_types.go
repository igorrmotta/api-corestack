package repository

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID        uuid.UUID
	TaskID    uuid.UUID
	AuthorID  string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateCommentParams struct {
	TaskID   uuid.UUID
	AuthorID string
	Content  string
}

type ListCommentsParams struct {
	TaskID    uuid.UUID
	PageSize  int32
	PageToken string
}

type CommentList struct {
	Comments      []Comment
	NextPageToken string
	TotalCount    int32
}
