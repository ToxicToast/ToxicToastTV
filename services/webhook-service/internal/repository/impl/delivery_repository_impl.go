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

type deliveryRepository struct {
	db *gorm.DB
}

func NewDeliveryRepository(db *gorm.DB) interfaces.DeliveryRepository {
	return &deliveryRepository{db: db}
}

func (r *deliveryRepository) Create(ctx context.Context, delivery *domain.Delivery) error {
	e := mapper.DeliveryToEntity(delivery)
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *deliveryRepository) GetByID(ctx context.Context, id string) (*domain.Delivery, error) {
	var e entity.DeliveryEntity
	err := r.db.WithContext(ctx).
		Preload("Webhook").
		Preload("Attempts").
		Where("id = ?", id).
		First(&e).Error
	if err != nil {
		return nil, err
	}
	return mapper.DeliveryToDomain(&e), nil
}

func (r *deliveryRepository) List(ctx context.Context, webhookID string, status domain.DeliveryStatus, limit, offset int) ([]*domain.Delivery, int64, error) {
	var entities []*entity.DeliveryEntity
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.DeliveryEntity{})

	if webhookID != "" {
		query = query.Where("webhook_id = ?", webhookID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
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

	if err := query.Order("created_at DESC").Preload("Webhook").Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return mapper.DeliveriesToDomain(entities), total, nil
}

func (r *deliveryRepository) Update(ctx context.Context, delivery *domain.Delivery) error {
	e := mapper.DeliveryToEntity(delivery)
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *deliveryRepository) GetPendingRetries(ctx context.Context, limit int) ([]*domain.Delivery, error) {
	var entities []*entity.DeliveryEntity

	now := time.Now()

	err := r.db.WithContext(ctx).
		Preload("Webhook").
		Where("status = ? AND next_retry_at IS NOT NULL AND next_retry_at <= ?", domain.DeliveryStatusRetrying, now).
		Limit(limit).
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	return mapper.DeliveriesToDomain(entities), nil
}

func (r *deliveryRepository) CreateAttempt(ctx context.Context, attempt *domain.DeliveryAttempt) error {
	e := mapper.DeliveryAttemptToEntity(attempt)
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *deliveryRepository) GetAttempts(ctx context.Context, deliveryID string) ([]*domain.DeliveryAttempt, error) {
	var entities []entity.DeliveryAttemptEntity
	err := r.db.WithContext(ctx).
		Where("delivery_id = ?", deliveryID).
		Order("created_at ASC").
		Find(&entities).Error
	if err != nil {
		return nil, err
	}

	// Convert from []domain.DeliveryAttempt to []*domain.DeliveryAttempt
	domainAttempts := mapper.DeliveryAttemptsToDomain(entities)
	result := make([]*domain.DeliveryAttempt, len(domainAttempts))
	for i := range domainAttempts {
		result[i] = &domainAttempts[i]
	}
	return result, nil
}

func (r *deliveryRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.DeliveryEntity{}).Error
}

func (r *deliveryRepository) CleanupOldDeliveries(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)

	return r.db.WithContext(ctx).
		Where("created_at < ? AND (status = ? OR status = ?)", cutoff, domain.DeliveryStatusSuccess, domain.DeliveryStatusFailed).
		Delete(&entity.DeliveryEntity{}).Error
}
