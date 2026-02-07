package handler

import (
	"context"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "github.com/igorrmotta/api-corestack/services/golang/gen/common/v1"
	commentv1 "github.com/igorrmotta/api-corestack/services/golang/gen/comment/v1"
	"github.com/igorrmotta/api-corestack/services/golang/gen/comment/v1/commentv1connect"
	"github.com/igorrmotta/api-corestack/services/golang/internal/domain"
	"github.com/igorrmotta/api-corestack/services/golang/internal/service"
)

type CommentHandler struct {
	commentv1connect.UnimplementedCommentServiceHandler
	svc *service.CommentService
}

func NewCommentHandler(svc *service.CommentService) *CommentHandler {
	return &CommentHandler{svc: svc}
}

func (h *CommentHandler) CreateComment(ctx context.Context, req *connect.Request[commentv1.CreateCommentRequest]) (*connect.Response[commentv1.CreateCommentResponse], error) {
	taskID, err := uuid.Parse(req.Msg.TaskId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	c, err := h.svc.Create(ctx, domain.CreateCommentParams{
		TaskID:   taskID,
		AuthorID: req.Msg.AuthorId,
		Content:  req.Msg.Content,
	})
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&commentv1.CreateCommentResponse{
		Comment: commentToProto(c),
	}), nil
}

func (h *CommentHandler) ListComments(ctx context.Context, req *connect.Request[commentv1.ListCommentsRequest]) (*connect.Response[commentv1.ListCommentsResponse], error) {
	taskID, err := uuid.Parse(req.Msg.TaskId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	var params domain.ListCommentsParams
	params.TaskID = taskID
	if req.Msg.Pagination != nil {
		params.PageSize = req.Msg.Pagination.PageSize
		params.PageToken = req.Msg.Pagination.PageToken
	}
	list, err := h.svc.List(ctx, params)
	if err != nil {
		return nil, toConnectError(err)
	}
	comments := make([]*commentv1.Comment, len(list.Comments))
	for i, c := range list.Comments {
		comments[i] = commentToProto(&c)
	}
	return connect.NewResponse(&commentv1.ListCommentsResponse{
		Comments: comments,
		Pagination: &commonv1.PaginationResponse{
			NextPageToken: list.NextPageToken,
			TotalCount:    list.TotalCount,
		},
	}), nil
}

func (h *CommentHandler) DeleteComment(ctx context.Context, req *connect.Request[commentv1.DeleteCommentRequest]) (*connect.Response[commentv1.DeleteCommentResponse], error) {
	id, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if err := h.svc.Delete(ctx, id); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&commentv1.DeleteCommentResponse{}), nil
}

func commentToProto(c *domain.Comment) *commentv1.Comment {
	return &commentv1.Comment{
		Id:        c.ID.String(),
		TaskId:    c.TaskID.String(),
		AuthorId:  c.AuthorID,
		Content:   c.Content,
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
	}
}
