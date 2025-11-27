package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	pb "toxictoast/services/notification-service/api/proto"
	"toxictoast/services/notification-service/internal/command"
	"toxictoast/services/notification-service/internal/handler/mapper"
	"toxictoast/services/notification-service/internal/query"
)

type NotificationHandler struct {
	pb.UnimplementedNotificationServiceServer
	commandBus *cqrs.CommandBus
	queryBus   *cqrs.QueryBus
}

func NewNotificationHandler(commandBus *cqrs.CommandBus, queryBus *cqrs.QueryBus) *NotificationHandler {
	return &NotificationHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
	}
}

func (h *NotificationHandler) GetNotification(ctx context.Context, req *pb.GetNotificationRequest) (*pb.NotificationResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	qry := &query.GetNotificationByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "notification not found: %v", err)
	}

	notifResult := result.(*query.GetNotificationResult)

	return &pb.NotificationResponse{
		Notification: mapper.ToProtoNotification(notifResult.Notification),
		Attempts:     mapper.ToProtoAttempts(notifResult.Attempts),
	}, nil
}

func (h *NotificationHandler) ListNotifications(ctx context.Context, req *pb.ListNotificationsRequest) (*pb.ListNotificationsResponse, error) {
	limit := int(req.Limit)
	if limit == 0 {
		limit = 50
	}
	if limit > 1000 {
		return nil, status.Error(codes.InvalidArgument, "limit cannot exceed 1000")
	}

	domainStatus := mapper.FromProtoStatus(req.Status)

	qry := &query.ListNotificationsQuery{
		BaseQuery: cqrs.BaseQuery{},
		ChannelID: req.ChannelId,
		Status:    domainStatus,
		Limit:     limit,
		Offset:    int(req.Offset),
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list notifications: %v", err)
	}

	listResult := result.(*query.ListNotificationsResult)

	return &pb.ListNotificationsResponse{
		Notifications: mapper.ToProtoNotifications(listResult.Notifications),
		Total:         listResult.Total,
		Limit:         int32(limit),
		Offset:        req.Offset,
	}, nil
}

func (h *NotificationHandler) DeleteNotification(ctx context.Context, req *pb.DeleteNotificationRequest) (*pb.DeleteNotificationResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	cmd := &command.DeleteNotificationCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete notification: %v", err)
	}

	return &pb.DeleteNotificationResponse{Success: true, Message: "Notification deleted successfully"}, nil
}

func (h *NotificationHandler) CleanupOldNotifications(ctx context.Context, req *pb.CleanupOldNotificationsRequest) (*pb.CleanupOldNotificationsResponse, error) {
	if req.OlderThanDays <= 0 {
		return nil, status.Error(codes.InvalidArgument, "older_than_days must be positive")
	}

	cmd := &command.CleanupOldNotificationsCommand{
		BaseCommand:   cqrs.BaseCommand{},
		OlderThanDays: int(req.OlderThanDays),
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to cleanup notifications: %v", err)
	}

	return &pb.CleanupOldNotificationsResponse{Message: "Cleanup completed successfully"}, nil
}
