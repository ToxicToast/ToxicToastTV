package mapper

import (
	"toxictoast/services/notification-service/internal/domain"
	pb "toxictoast/services/notification-service/api/proto"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToProtoNotification(notification *domain.Notification) *pb.Notification {
	if notification == nil {
		return nil
	}

	proto := &pb.Notification{
		Id:               notification.ID,
		ChannelId:        notification.ChannelID,
		EventId:          notification.EventID,
		EventType:        notification.EventType,
		EventPayload:     notification.EventPayload,
		DiscordMessageId: notification.DiscordMessageID,
		Status:           ToProtoStatus(notification.Status),
		AttemptCount:     int32(notification.AttemptCount),
		LastError:        notification.LastError,
		CreatedAt:        timestamppb.New(notification.CreatedAt),
		UpdatedAt:        timestamppb.New(notification.UpdatedAt),
	}

	if notification.SentAt != nil {
		proto.SentAt = timestamppb.New(*notification.SentAt)
	}

	return proto
}

func ToProtoNotifications(notifications []*domain.Notification) []*pb.Notification {
	result := make([]*pb.Notification, 0, len(notifications))
	for _, notification := range notifications {
		result = append(result, ToProtoNotification(notification))
	}
	return result
}

func ToProtoStatus(status domain.NotificationStatus) pb.NotificationStatus {
	switch status {
	case domain.NotificationStatusPending:
		return pb.NotificationStatus_NOTIFICATION_STATUS_PENDING
	case domain.NotificationStatusSuccess:
		return pb.NotificationStatus_NOTIFICATION_STATUS_SUCCESS
	case domain.NotificationStatusFailed:
		return pb.NotificationStatus_NOTIFICATION_STATUS_FAILED
	default:
		return pb.NotificationStatus_NOTIFICATION_STATUS_UNSPECIFIED
	}
}

func FromProtoStatus(status pb.NotificationStatus) domain.NotificationStatus {
	switch status {
	case pb.NotificationStatus_NOTIFICATION_STATUS_PENDING:
		return domain.NotificationStatusPending
	case pb.NotificationStatus_NOTIFICATION_STATUS_SUCCESS:
		return domain.NotificationStatusSuccess
	case pb.NotificationStatus_NOTIFICATION_STATUS_FAILED:
		return domain.NotificationStatusFailed
	default:
		return ""
	}
}

func ToProtoAttempt(attempt *domain.NotificationAttempt) *pb.NotificationAttempt {
	if attempt == nil {
		return nil
	}

	return &pb.NotificationAttempt{
		Id:               attempt.ID,
		NotificationId:   attempt.NotificationID,
		AttemptNumber:    int32(attempt.AttemptNumber),
		ResponseStatus:   int32(attempt.ResponseStatus),
		ResponseBody:     attempt.ResponseBody,
		DiscordMessageId: attempt.DiscordMessageID,
		Success:          attempt.Success,
		Error:            attempt.Error,
		DurationMs:       int32(attempt.DurationMs),
		CreatedAt:        timestamppb.New(attempt.CreatedAt),
	}
}

func ToProtoAttempts(attempts []*domain.NotificationAttempt) []*pb.NotificationAttempt {
	result := make([]*pb.NotificationAttempt, 0, len(attempts))
	for _, attempt := range attempts {
		result = append(result, ToProtoAttempt(attempt))
	}
	return result
}
