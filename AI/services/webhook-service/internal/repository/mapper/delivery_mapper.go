package mapper

import (
	"toxictoast/services/webhook-service/internal/domain"
	"toxictoast/services/webhook-service/internal/repository/entity"
)

func DeliveryToEntity(delivery *domain.Delivery) *entity.DeliveryEntity {
	if delivery == nil {
		return nil
	}

	return &entity.DeliveryEntity{
		ID:            delivery.ID,
		WebhookID:     delivery.WebhookID,
		EventID:       delivery.EventID,
		EventType:     delivery.EventType,
		EventPayload:  delivery.EventPayload,
		Status:        string(delivery.Status),
		AttemptCount:  delivery.AttemptCount,
		NextRetryAt:   delivery.NextRetryAt,
		LastAttemptAt: delivery.LastAttemptAt,
		CompletedAt:   delivery.CompletedAt,
		LastError:     delivery.LastError,
		CreatedAt:     delivery.CreatedAt,
		UpdatedAt:     delivery.UpdatedAt,
	}
}

func DeliveryToDomain(e *entity.DeliveryEntity) *domain.Delivery {
	if e == nil {
		return nil
	}

	delivery := &domain.Delivery{
		ID:            e.ID,
		WebhookID:     e.WebhookID,
		EventID:       e.EventID,
		EventType:     e.EventType,
		EventPayload:  e.EventPayload,
		Status:        domain.DeliveryStatus(e.Status),
		AttemptCount:  e.AttemptCount,
		NextRetryAt:   e.NextRetryAt,
		LastAttemptAt: e.LastAttemptAt,
		CompletedAt:   e.CompletedAt,
		LastError:     e.LastError,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}

	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		delivery.DeletedAt = &deletedAt
	}

	// Convert webhook if preloaded
	if e.Webhook != nil {
		delivery.Webhook = WebhookToDomain(e.Webhook)
	}

	// Convert attempts if preloaded
	if len(e.Attempts) > 0 {
		delivery.Attempts = DeliveryAttemptsToDomain(e.Attempts)
	}

	return delivery
}

func DeliveriesToDomain(entities []*entity.DeliveryEntity) []*domain.Delivery {
	deliveries := make([]*domain.Delivery, len(entities))
	for i, e := range entities {
		deliveries[i] = DeliveryToDomain(e)
	}
	return deliveries
}
