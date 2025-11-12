package mapper

import (
	"gorm.io/gorm"

	"toxictoast/services/webhook-service/internal/domain"
	"toxictoast/services/webhook-service/internal/repository/entity"
)

func WebhookToEntity(webhook *domain.Webhook) *entity.WebhookEntity {
	if webhook == nil {
		return nil
	}

	e := &entity.WebhookEntity{
		ID:                webhook.ID,
		URL:               webhook.URL,
		Secret:            webhook.Secret,
		EventTypes:        webhook.EventTypes,
		Description:       webhook.Description,
		Active:            webhook.Active,
		CreatedAt:         webhook.CreatedAt,
		UpdatedAt:         webhook.UpdatedAt,
		TotalDeliveries:   webhook.TotalDeliveries,
		SuccessDeliveries: webhook.SuccessDeliveries,
		FailedDeliveries:  webhook.FailedDeliveries,
		LastDeliveryAt:    webhook.LastDeliveryAt,
		LastSuccessAt:     webhook.LastSuccessAt,
		LastFailureAt:     webhook.LastFailureAt,
	}

	if webhook.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{Time: *webhook.DeletedAt, Valid: true}
	}

	return e
}

func WebhookToDomain(e *entity.WebhookEntity) *domain.Webhook {
	if e == nil {
		return nil
	}

	webhook := &domain.Webhook{
		ID:                e.ID,
		URL:               e.URL,
		Secret:            e.Secret,
		EventTypes:        e.EventTypes,
		Description:       e.Description,
		Active:            e.Active,
		CreatedAt:         e.CreatedAt,
		UpdatedAt:         e.UpdatedAt,
		TotalDeliveries:   e.TotalDeliveries,
		SuccessDeliveries: e.SuccessDeliveries,
		FailedDeliveries:  e.FailedDeliveries,
		LastDeliveryAt:    e.LastDeliveryAt,
		LastSuccessAt:     e.LastSuccessAt,
		LastFailureAt:     e.LastFailureAt,
	}

	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		webhook.DeletedAt = &deletedAt
	}

	return webhook
}

func WebhooksToDomain(entities []*entity.WebhookEntity) []*domain.Webhook {
	webhooks := make([]*domain.Webhook, len(entities))
	for i, e := range entities {
		webhooks[i] = WebhookToDomain(e)
	}
	return webhooks
}
