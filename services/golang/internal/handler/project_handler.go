package handler

import (
	"context"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "github.com/igorrmotta/api-corestack/services/golang/gen/common/v1"
	projectv1 "github.com/igorrmotta/api-corestack/services/golang/gen/project/v1"
	"github.com/igorrmotta/api-corestack/services/golang/gen/project/v1/projectv1connect"
	"github.com/igorrmotta/api-corestack/services/golang/internal/repository"
	"github.com/igorrmotta/api-corestack/services/golang/internal/service"
)

type ProjectHandler struct {
	projectv1connect.UnimplementedProjectServiceHandler
	svc *service.ProjectService
}

func NewProjectHandler(svc *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

func (h *ProjectHandler) CreateProject(ctx context.Context, req *connect.Request[projectv1.CreateProjectRequest]) (*connect.Response[projectv1.CreateProjectResponse], error) {
	workspaceID, err := uuid.Parse(req.Msg.WorkspaceId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	p, err := h.svc.Create(ctx, repository.CreateProjectParams{
		WorkspaceID: workspaceID,
		Name:        req.Msg.Name,
		Description: req.Msg.Description,
	})
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&projectv1.CreateProjectResponse{
		Project: projectToProto(p),
	}), nil
}

func (h *ProjectHandler) GetProject(ctx context.Context, req *connect.Request[projectv1.GetProjectRequest]) (*connect.Response[projectv1.GetProjectResponse], error) {
	id, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	p, err := h.svc.GetByID(ctx, id)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&projectv1.GetProjectResponse{
		Project: projectToProto(p),
	}), nil
}

func (h *ProjectHandler) ListProjects(ctx context.Context, req *connect.Request[projectv1.ListProjectsRequest]) (*connect.Response[projectv1.ListProjectsResponse], error) {
	workspaceID, err := uuid.Parse(req.Msg.WorkspaceId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	var params repository.ListProjectsParams
	params.WorkspaceID = workspaceID
	if req.Msg.Pagination != nil {
		params.PageSize = req.Msg.Pagination.PageSize
		params.PageToken = req.Msg.Pagination.PageToken
	}
	list, err := h.svc.List(ctx, params)
	if err != nil {
		return nil, toConnectError(err)
	}
	projects := make([]*projectv1.Project, len(list.Projects))
	for i, p := range list.Projects {
		projects[i] = projectToProto(&p)
	}
	return connect.NewResponse(&projectv1.ListProjectsResponse{
		Projects: projects,
		Pagination: &commonv1.PaginationResponse{
			NextPageToken: list.NextPageToken,
			TotalCount:    list.TotalCount,
		},
	}), nil
}

func (h *ProjectHandler) UpdateProject(ctx context.Context, req *connect.Request[projectv1.UpdateProjectRequest]) (*connect.Response[projectv1.UpdateProjectResponse], error) {
	id, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	p, err := h.svc.Update(ctx, repository.UpdateProjectParams{
		ID:          id,
		Name:        req.Msg.Name,
		Description: req.Msg.Description,
		Status:      req.Msg.Status,
	})
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&projectv1.UpdateProjectResponse{
		Project: projectToProto(p),
	}), nil
}

func (h *ProjectHandler) DeleteProject(ctx context.Context, req *connect.Request[projectv1.DeleteProjectRequest]) (*connect.Response[projectv1.DeleteProjectResponse], error) {
	id, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if err := h.svc.Delete(ctx, id); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&projectv1.DeleteProjectResponse{}), nil
}

func projectToProto(p *repository.Project) *projectv1.Project {
	return &projectv1.Project{
		Id:          p.ID.String(),
		WorkspaceId: p.WorkspaceID.String(),
		Name:        p.Name,
		Description: p.Description,
		Status:      p.Status,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}
}
