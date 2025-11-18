package mapper

import (
	"toxictoast/services/webhook-service/internal/domain"
	pb "toxictoast/services/webhook-service/api/proto"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ToProtoDelivery converts domain.Delivery to proto Delivery
func ToProtoDelivery(delivery *domain.Delivery) *pb.Delivery {
	if delivery == nil {
		return nil
	}

	proto := &pb.Delivery{
		Id:           delivery.ID,
		WebhookId:    delivery.WebhookID,
		EventId:      delivery.EventID,
		EventType:    delivery.EventType,
		EventPayload: delivery.EventPayload,
		Status:       ToProtoDeliveryStatus(delivery.Status),
		AttemptCount: int32(delivery.AttemptCount),
		LastError:    delivery.LastError,
		CreatedAt:    timestamppb.New(delivery.CreatedAt),
		UpdatedAt:    timestamppb.New(delivery.UpdatedAt),
	}

	if delivery.NextRetryAt != nil {
		proto.NextRetryAt = timestamppb.New(*delivery.NextRetryAt)
	}
	if delivery.LastAttemptAt != nil {
		proto.LastAttemptAt = timestamppb.New(*delivery.LastAttemptAt)
	}

	return proto
}

// ToProtoDeliveries converts a slice of domain deliveries to proto deliveries
func ToProtoDeliveries(deliveries []*domain.Delivery) []*pb.Delivery {
	result := make([]*pb.Delivery, 0, len(deliveries))
	for _, delivery := range deliveries {
		result = append(result, ToProtoDelivery(delivery))
	}
	return result
}

// ToProtoDeliveryStatus converts domain.DeliveryStatus to proto DeliveryStatus
func ToProtoDeliveryStatus(status domain.DeliveryStatus) pb.DeliveryStatus {
	switch status {
	case domain.DeliveryStatusPending:
		return pb.DeliveryStatus_DELIVERY_STATUS_PENDING
	case domain.DeliveryStatusSuccess:
		return pb.DeliveryStatus_DELIVERY_STATUS_SUCCESS
	case domain.DeliveryStatusFailed:
		return pb.DeliveryStatus_DELIVERY_STATUS_FAILED
	case domain.DeliveryStatusRetrying:
		return pb.DeliveryStatus_DELIVERY_STATUS_RETRYING
	default:
		return pb.DeliveryStatus_DELIVERY_STATUS_UNSPECIFIED
	}
}

// FromProtoDeliveryStatus converts proto DeliveryStatus to domain.DeliveryStatus
func FromProtoDeliveryStatus(status pb.DeliveryStatus) domain.DeliveryStatus {
	switch status {
	case pb.DeliveryStatus_DELIVERY_STATUS_PENDING:
		return domain.DeliveryStatusPending
	case pb.DeliveryStatus_DELIVERY_STATUS_SUCCESS:
		return domain.DeliveryStatusSuccess
	case pb.DeliveryStatus_DELIVERY_STATUS_FAILED:
		return domain.DeliveryStatusFailed
	case pb.DeliveryStatus_DELIVERY_STATUS_RETRYING:
		return domain.DeliveryStatusRetrying
	default:
		return ""
	}
}

// ToProtoDeliveryAttempt converts domain.DeliveryAttempt to proto DeliveryAttempt
func ToProtoDeliveryAttempt(attempt *domain.DeliveryAttempt) *pb.DeliveryAttempt {
	if attempt == nil {
		return nil
	}

	return &pb.DeliveryAttempt{
		Id:             attempt.ID,
		DeliveryId:     attempt.DeliveryID,
		AttemptNumber:  int32(attempt.AttemptNumber),
		RequestUrl:     attempt.RequestURL,
		ResponseStatus: int32(attempt.ResponseStatus),
		ResponseBody:   attempt.ResponseBody,
		Success:        attempt.Success,
		ErrorMessage:   attempt.Error,
		DurationMs:     int32(attempt.DurationMs),
		CreatedAt:      timestamppb.New(attempt.CreatedAt),
	}
}

// ToProtoDeliveryAttempts converts a slice of domain attempts to proto attempts
func ToProtoDeliveryAttempts(attempts []*domain.DeliveryAttempt) []*pb.DeliveryAttempt {
	result := make([]*pb.DeliveryAttempt, 0, len(attempts))
	for _, attempt := range attempts {
		result = append(result, ToProtoDeliveryAttempt(attempt))
	}
	return result
}
