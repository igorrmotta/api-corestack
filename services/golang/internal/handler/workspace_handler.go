package handler

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "github.com/igorrmotta/api-corestack/services/golang/gen/common/v1"
	workspacev1 "github.com/igorrmotta/api-corestack/services/golang/gen/workspace/v1"
	"github.com/igorrmotta/api-corestack/services/golang/gen/workspace/v1/workspacev1connect"
	"github.com/igorrmotta/api-corestack/services/golang/internal/domain"
	"github.com/igorrmotta/api-corestack/services/golang/internal/service"
)

type WorkspaceHandler struct {
	workspacev1connect.UnimplementedWorkspaceServiceHandler
	svc *service.WorkspaceService
}

func NewWorkspaceHandler(svc *service.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{svc: svc}
}

func (h *WorkspaceHandler) CreateWorkspace(ctx context.Context, req *connect.Request[workspacev1.CreateWorkspaceRequest]) (*connect.Response[workspacev1.CreateWorkspaceResponse], error) {
	w, err := h.svc.Create(ctx, domain.CreateWorkspaceParams{
		Name: req.Msg.Name,
		Slug: req.Msg.Slug,
	})
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&workspacev1.CreateWorkspaceResponse{
		Workspace: workspaceToProto(w),
	}), nil
}

func (h *WorkspaceHandler) GetWorkspace(ctx context.Context, req *connect.Request[workspacev1.GetWorkspaceRequest]) (*connect.Response[workspacev1.GetWorkspaceResponse], error) {
	id, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	w, err := h.svc.GetByID(ctx, id)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&workspacev1.GetWorkspaceResponse{
		Workspace: workspaceToProto(w),
	}), nil
}

func (h *WorkspaceHandler) ListWorkspaces(ctx context.Context, req *connect.Request[workspacev1.ListWorkspacesRequest]) (*connect.Response[workspacev1.ListWorkspacesResponse], error) {
	var params domain.ListWorkspacesParams
	if req.Msg.Pagination != nil {
		params.PageSize = req.Msg.Pagination.PageSize
		params.PageToken = req.Msg.Pagination.PageToken
	}
	list, err := h.svc.List(ctx, params)
	if err != nil {
		return nil, toConnectError(err)
	}
	workspaces := make([]*workspacev1.Workspace, len(list.Workspaces))
	for i, w := range list.Workspaces {
		workspaces[i] = workspaceToProto(&w)
	}
	return connect.NewResponse(&workspacev1.ListWorkspacesResponse{
		Workspaces: workspaces,
		Pagination: &commonv1.PaginationResponse{
			NextPageToken: list.NextPageToken,
			TotalCount:    list.TotalCount,
		},
	}), nil
}

func (h *WorkspaceHandler) UpdateWorkspace(ctx context.Context, req *connect.Request[workspacev1.UpdateWorkspaceRequest]) (*connect.Response[workspacev1.UpdateWorkspaceResponse], error) {
	id, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	w, err := h.svc.Update(ctx, domain.UpdateWorkspaceParams{
		ID:   id,
		Name: req.Msg.Name,
		Slug: req.Msg.Slug,
	})
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&workspacev1.UpdateWorkspaceResponse{
		Workspace: workspaceToProto(w),
	}), nil
}

func (h *WorkspaceHandler) DeleteWorkspace(ctx context.Context, req *connect.Request[workspacev1.DeleteWorkspaceRequest]) (*connect.Response[workspacev1.DeleteWorkspaceResponse], error) {
	id, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if err := h.svc.Delete(ctx, id); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&workspacev1.DeleteWorkspaceResponse{}), nil
}

func workspaceToProto(w *domain.Workspace) *workspacev1.Workspace {
	return &workspacev1.Workspace{
		Id:        w.ID.String(),
		Name:      w.Name,
		Slug:      w.Slug,
		CreatedAt: timestamppb.New(w.CreatedAt),
		UpdatedAt: timestamppb.New(w.UpdatedAt),
	}
}

// toConnectError maps domain errors to Connect RPC error codes.
func toConnectError(err error) error {
	if errors.Is(err, domain.ErrNotFound) {
		return connect.NewError(connect.CodeNotFound, err)
	}
	if errors.Is(err, domain.ErrInvalidInput) {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}
	if errors.Is(err, domain.ErrConflict) {
		return connect.NewError(connect.CodeAlreadyExists, err)
	}
	return connect.NewError(connect.CodeInternal, err)
}
