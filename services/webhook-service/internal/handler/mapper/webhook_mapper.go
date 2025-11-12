package mapper

import (
	"strings"

	"toxictoast/services/webhook-service/internal/domain"
	pb "toxictoast/services/webhook-service/api/proto"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ToProtoWebhook converts domain.Webhook to proto Webhook
func ToProtoWebhook(webhook *domain.Webhook) *pb.Webhook {
	if webhook == nil {
		return nil
	}

	// Parse event types
	eventTypes := []string{}
	if webhook.EventTypes != "" {
		eventTypes = strings.Split(webhook.EventTypes, ",")
	}

	proto := &pb.Webhook{
		Id:               webhook.ID,
		Url:              webhook.URL,
		Secret:           webhook.Secret,
		EventTypes:       eventTypes,
		Description:      webhook.Description,
		Active:           webhook.Active,
		TotalDeliveries:  int32(webhook.TotalDeliveries),
		SuccessDeliveries: int32(webhook.SuccessDeliveries),
		FailedDeliveries: int32(webhook.FailedDeliveries),
		CreatedAt:        timestamppb.New(webhook.CreatedAt),
		UpdatedAt:        timestamppb.New(webhook.UpdatedAt),
	}

	if !webhook.LastDeliveryAt.IsZero() {
		proto.LastDeliveryAt = timestamppb.New(webhook.LastDeliveryAt)
	}
	if !webhook.LastSuccessAt.IsZero() {
		proto.LastSuccessAt = timestamppb.New(webhook.LastSuccessAt)
	}
	if !webhook.LastFailureAt.IsZero() {
		proto.LastFailureAt = timestamppb.New(webhook.LastFailureAt)
	}

	return proto
}

// ToProtoWebhooks converts a slice of domain webhooks to proto webhooks
func ToProtoWebhooks(webhooks []*domain.Webhook) []*pb.Webhook {
	result := make([]*pb.Webhook, 0, len(webhooks))
	for _, webhook := range webhooks {
		result = append(result, ToProtoWebhook(webhook))
	}
	return result
}
