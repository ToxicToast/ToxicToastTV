package impl

import (
	"context"
	"time"

	"gorm.io/gorm"

	"toxictoast/services/webhook-service/internal/domain"
	"toxictoast/services/webhook-service/internal/repository/entity"
	"toxictoast/services/webhook-service/internal/repository/interfaces"
	"toxictoast/services/webhook-service/internal/repository/mapper"
)

type webhookRepository struct {
	db *gorm.DB
}

func NewWebhookRepository(db *gorm.DB) interfaces.WebhookRepository {
	return &webhookRepository{db: db}
}

func (r *webhookRepository) Create(ctx context.Context, webhook *domain.Webhook) error {
	e := mapper.WebhookToEntity(webhook)
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *webhookRepository) GetByID(ctx context.Context, id string) (*domain.Webhook, error) {
	var e entity.WebhookEntity
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&e).Error
	if err != nil {
		return nil, err
	}
	return mapper.WebhookToDomain(&e), nil
}

func (r *webhookRepository) GetByURL(ctx context.Context, url string) (*domain.Webhook, error) {
	var e entity.WebhookEntity
	err := r.db.WithContext(ctx).Where("url = ?", url).First(&e).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return mapper.WebhookToDomain(&e), nil
}

func (r *webhookRepository) List(ctx context.Context, limit, offset int, activeOnly bool) ([]*domain.Webhook, int64, error) {
	var entities []*entity.WebhookEntity
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.WebhookEntity{})
	if activeOnly {
		query = query.Where("active = ?", true)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Order("created_at DESC").Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return mapper.WebhooksToDomain(entities), total, nil
}

func (r *webhookRepository) Update(ctx context.Context, webhook *domain.Webhook) error {
	e := mapper.WebhookToEntity(webhook)
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *webhookRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.WebhookEntity{}).Error
}

func (r *webhookRepository) GetActiveWebhooksForEvent(ctx context.Context, eventType string) ([]*domain.Webhook, error) {
	var entities []*entity.WebhookEntity

	// Get all active webhooks
	err := r.db.WithContext(ctx).
		Where("active = ?", true).
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	// Convert to domain and filter by event type match
	webhooks := mapper.WebhooksToDomain(entities)
	result := make([]*domain.Webhook, 0)
	for _, webhook := range webhooks {
		if webhook.MatchesEvent(eventType) {
			result = append(result, webhook)
		}
	}

	return result, nil
}

func (r *webhookRepository) UpdateStatistics(ctx context.Context, id string, success bool) error {
	updates := map[string]interface{}{
		"total_deliveries": gorm.Expr("total_deliveries + 1"),
		"last_delivery_at": time.Now(),
	}

	if success {
		updates["success_deliveries"] = gorm.Expr("success_deliveries + 1")
		updates["last_success_at"] = time.Now()
	} else {
		updates["failed_deliveries"] = gorm.Expr("failed_deliveries + 1")
		updates["last_failure_at"] = time.Now()
	}

	return r.db.WithContext(ctx).
		Model(&entity.WebhookEntity{}).
		Where("id = ?", id).
		Updates(updates).Error
}
