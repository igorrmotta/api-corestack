package handler

import (
	"context"
	"encoding/json"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "github.com/igorrmotta/api-corestack/services/golang/gen/common/v1"
	notificationv1 "github.com/igorrmotta/api-corestack/services/golang/gen/notification/v1"
	"github.com/igorrmotta/api-corestack/services/golang/gen/notification/v1/notificationv1connect"
	"github.com/igorrmotta/api-corestack/services/golang/internal/domain"
)

type NotificationHandler struct {
	notificationv1connect.UnimplementedNotificationServiceHandler
	repo domain.NotificationRepository
}

func NewNotificationHandler(repo domain.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{repo: repo}
}

func (h *NotificationHandler) ListNotifications(ctx context.Context, req *connect.Request[notificationv1.ListNotificationsRequest]) (*connect.Response[notificationv1.ListNotificationsResponse], error) {
	workspaceID, err := uuid.Parse(req.Msg.WorkspaceId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	var params domain.ListNotificationsParams
	params.WorkspaceID = workspaceID
	params.Status = req.Msg.Status
	if req.Msg.Pagination != nil {
		params.PageSize = req.Msg.Pagination.PageSize
		params.PageToken = req.Msg.Pagination.PageToken
	}
	list, err := h.repo.List(ctx, params)
	if err != nil {
		return nil, toConnectError(err)
	}
	notifications := make([]*notificationv1.Notification, len(list.Notifications))
	for i, n := range list.Notifications {
		proto, err := notificationToProto(&n)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		notifications[i] = proto
	}
	return connect.NewResponse(&notificationv1.ListNotificationsResponse{
		Notifications: notifications,
		Pagination: &commonv1.PaginationResponse{
			NextPageToken: list.NextPageToken,
			TotalCount:    list.TotalCount,
		},
	}), nil
}

func (h *NotificationHandler) MarkNotificationRead(ctx context.Context, req *connect.Request[notificationv1.MarkNotificationReadRequest]) (*connect.Response[notificationv1.MarkNotificationReadResponse], error) {
	n, err := h.repo.MarkProcessed(ctx, req.Msg.Id)
	if err != nil {
		return nil, toConnectError(err)
	}
	proto, err := notificationToProto(n)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&notificationv1.MarkNotificationReadResponse{
		Notification: proto,
	}), nil
}

func notificationToProto(n *domain.Notification) (*notificationv1.Notification, error) {
	proto := &notificationv1.Notification{
		Id:          n.ID,
		WorkspaceId: n.WorkspaceID.String(),
		EventType:   n.EventType,
		Status:      n.Status,
		RetryCount:  n.RetryCount,
		CreatedAt:   timestamppb.New(n.CreatedAt),
	}
	if n.ProcessedAt != nil {
		proto.ProcessedAt = timestamppb.New(*n.ProcessedAt)
	}
	if len(n.Payload) > 0 {
		var m map[string]any
		if err := json.Unmarshal(n.Payload, &m); err != nil {
			return nil, err
		}
		s, err := structpb.NewStruct(m)
		if err != nil {
			return nil, err
		}
		proto.Payload = s
	}
	return proto, nil
}
