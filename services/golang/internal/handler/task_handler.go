package handler

import (
	"context"
	"encoding/json"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "github.com/igorrmotta/api-corestack/services/golang/gen/common/v1"
	taskv1 "github.com/igorrmotta/api-corestack/services/golang/gen/task/v1"
	"github.com/igorrmotta/api-corestack/services/golang/gen/task/v1/taskv1connect"
	"github.com/igorrmotta/api-corestack/services/golang/internal/domain"
	"github.com/igorrmotta/api-corestack/services/golang/internal/service"
)

type TaskHandler struct {
	taskv1connect.UnimplementedTaskServiceHandler
	svc       *service.TaskService
	importSvc *service.ImportService
}

func NewTaskHandler(svc *service.TaskService, importSvc *service.ImportService) *TaskHandler {
	return &TaskHandler{svc: svc, importSvc: importSvc}
}

func (h *TaskHandler) CreateTask(ctx context.Context, req *connect.Request[taskv1.CreateTaskRequest]) (*connect.Response[taskv1.CreateTaskResponse], error) {
	workspaceID, err := uuid.Parse(req.Msg.WorkspaceId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	projectID, err := uuid.Parse(req.Msg.ProjectId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	params := domain.CreateTaskParams{
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		Title:       req.Msg.Title,
		Description: req.Msg.Description,
		Priority:    req.Msg.Priority,
		AssignedTo:  req.Msg.AssignedTo,
	}
	if req.Msg.DueDate != nil {
		t := req.Msg.DueDate.AsTime()
		params.DueDate = &t
	}
	if req.Msg.Metadata != nil {
		b, err := json.Marshal(req.Msg.Metadata.AsMap())
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		params.Metadata = b
	}

	task, err := h.svc.Create(ctx, params)
	if err != nil {
		return nil, toConnectError(err)
	}
	proto, err := taskToProto(task)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&taskv1.CreateTaskResponse{
		Task: proto,
	}), nil
}

func (h *TaskHandler) GetTask(ctx context.Context, req *connect.Request[taskv1.GetTaskRequest]) (*connect.Response[taskv1.GetTaskResponse], error) {
	id, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	task, err := h.svc.GetByID(ctx, id)
	if err != nil {
		return nil, toConnectError(err)
	}
	proto, err := taskToProto(task)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&taskv1.GetTaskResponse{
		Task: proto,
	}), nil
}

func (h *TaskHandler) ListTasks(ctx context.Context, req *connect.Request[taskv1.ListTasksRequest]) (*connect.Response[taskv1.ListTasksResponse], error) {
	workspaceID, err := uuid.Parse(req.Msg.WorkspaceId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	params := domain.ListTasksParams{
		WorkspaceID: workspaceID,
		Status:      req.Msg.Status,
		Priority:    req.Msg.Priority,
		AssignedTo:  req.Msg.AssignedTo,
	}
	if req.Msg.ProjectId != "" {
		projectID, err := uuid.Parse(req.Msg.ProjectId)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		params.ProjectID = projectID
	}
	if req.Msg.Pagination != nil {
		params.PageSize = req.Msg.Pagination.PageSize
		params.PageToken = req.Msg.Pagination.PageToken
	}

	list, err := h.svc.List(ctx, params)
	if err != nil {
		return nil, toConnectError(err)
	}
	tasks := make([]*taskv1.Task, len(list.Tasks))
	for i, t := range list.Tasks {
		proto, err := taskToProto(&t)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		tasks[i] = proto
	}
	return connect.NewResponse(&taskv1.ListTasksResponse{
		Tasks: tasks,
		Pagination: &commonv1.PaginationResponse{
			NextPageToken: list.NextPageToken,
			TotalCount:    list.TotalCount,
		},
	}), nil
}

func (h *TaskHandler) UpdateTask(ctx context.Context, req *connect.Request[taskv1.UpdateTaskRequest]) (*connect.Response[taskv1.UpdateTaskResponse], error) {
	id, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	params := domain.UpdateTaskParams{
		ID:          id,
		Title:       req.Msg.Title,
		Description: req.Msg.Description,
		Status:      req.Msg.Status,
		Priority:    req.Msg.Priority,
		AssignedTo:  req.Msg.AssignedTo,
	}
	if req.Msg.DueDate != nil {
		t := req.Msg.DueDate.AsTime()
		params.DueDate = &t
	}
	if req.Msg.Metadata != nil {
		b, err := json.Marshal(req.Msg.Metadata.AsMap())
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		params.Metadata = b
	}

	task, err := h.svc.Update(ctx, params)
	if err != nil {
		return nil, toConnectError(err)
	}
	proto, err := taskToProto(task)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&taskv1.UpdateTaskResponse{
		Task: proto,
	}), nil
}

func (h *TaskHandler) DeleteTask(ctx context.Context, req *connect.Request[taskv1.DeleteTaskRequest]) (*connect.Response[taskv1.DeleteTaskResponse], error) {
	id, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if err := h.svc.Delete(ctx, id); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&taskv1.DeleteTaskResponse{}), nil
}

func (h *TaskHandler) BulkImportTasks(ctx context.Context, req *connect.Request[taskv1.BulkImportTasksRequest]) (*connect.Response[taskv1.BulkImportTasksResponse], error) {
	workspaceID, err := uuid.Parse(req.Msg.WorkspaceId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	projectID, err := uuid.Parse(req.Msg.ProjectId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	inputs := make([]domain.TaskInput, len(req.Msg.Tasks))
	for i, t := range req.Msg.Tasks {
		input := domain.TaskInput{
			Title:       t.Title,
			Description: t.Description,
			Priority:    t.Priority,
			AssignedTo:  t.AssignedTo,
		}
		if t.DueDate != nil {
			d := t.DueDate.AsTime()
			input.DueDate = &d
		}
		if t.Metadata != nil {
			b, err := json.Marshal(t.Metadata.AsMap())
			if err != nil {
				return nil, connect.NewError(connect.CodeInvalidArgument, err)
			}
			input.Metadata = b
		}
		inputs[i] = input
	}

	result := h.importSvc.BulkImport(ctx, workspaceID, projectID, inputs)

	taskErrors := make([]*taskv1.TaskError, len(result.Errors))
	for i, e := range result.Errors {
		taskErrors[i] = &taskv1.TaskError{
			Index: e.Index,
			Error: e.Error,
		}
	}

	return connect.NewResponse(&taskv1.BulkImportTasksResponse{
		Total:     result.Total,
		Succeeded: result.Succeeded,
		Failed:    result.Failed,
		Errors:    taskErrors,
	}), nil
}

func taskToProto(t *domain.Task) (*taskv1.Task, error) {
	proto := &taskv1.Task{
		Id:          t.ID.String(),
		WorkspaceId: t.WorkspaceID.String(),
		ProjectId:   t.ProjectID.String(),
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		Priority:    t.Priority,
		AssignedTo:  t.AssignedTo,
		CreatedAt:   timestamppb.New(t.CreatedAt),
		UpdatedAt:   timestamppb.New(t.UpdatedAt),
	}
	if t.DueDate != nil {
		proto.DueDate = timestamppb.New(*t.DueDate)
	}
	if len(t.Metadata) > 0 {
		var m map[string]any
		if err := json.Unmarshal(t.Metadata, &m); err != nil {
			return nil, err
		}
		s, err := structpb.NewStruct(m)
		if err != nil {
			return nil, err
		}
		proto.Metadata = s
	}
	return proto, nil
}
