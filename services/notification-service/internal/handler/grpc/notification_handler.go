package grpc

import (
	"context"

	"toxictoast/services/notification-service/internal/handler/mapper"
	"toxictoast/services/notification-service/internal/usecase"
	pb "toxictoast/services/notification-service/api/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NotificationHandler struct {
	pb.UnimplementedNotificationServiceServer
	notificationUC *usecase.NotificationUseCase
}

func NewNotificationHandler(notificationUC *usecase.NotificationUseCase) *NotificationHandler {
	return &NotificationHandler{
		notificationUC: notificationUC,
	}
}

func (h *NotificationHandler) GetNotification(ctx context.Context, req *pb.GetNotificationRequest) (*pb.NotificationResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	notification, attempts, err := h.notificationUC.GetNotification(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "notification not found: %v", err)
	}

	return &pb.NotificationResponse{
		Notification: mapper.ToProtoNotification(notification),
		Attempts:     mapper.ToProtoAttempts(attempts),
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

	notifications, total, err := h.notificationUC.ListNotifications(ctx, req.ChannelId, domainStatus, limit, int(req.Offset))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list notifications: %v", err)
	}

	return &pb.ListNotificationsResponse{
		Notifications: mapper.ToProtoNotifications(notifications),
		Total:         total,
		Limit:         int32(limit),
		Offset:        req.Offset,
	}, nil
}

func (h *NotificationHandler) DeleteNotification(ctx context.Context, req *pb.DeleteNotificationRequest) (*pb.DeleteNotificationResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if err := h.notificationUC.DeleteNotification(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete notification: %v", err)
	}

	return &pb.DeleteNotificationResponse{Success: true, Message: "Notification deleted successfully"}, nil
}

func (h *NotificationHandler) CleanupOldNotifications(ctx context.Context, req *pb.CleanupOldNotificationsRequest) (*pb.CleanupOldNotificationsResponse, error) {
	if req.OlderThanDays <= 0 {
		return nil, status.Error(codes.InvalidArgument, "older_than_days must be positive")
	}

	if err := h.notificationUC.CleanupOldNotifications(ctx, int(req.OlderThanDays)); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to cleanup notifications: %v", err)
	}

	return &pb.CleanupOldNotificationsResponse{Message: "Cleanup completed successfully"}, nil
}
