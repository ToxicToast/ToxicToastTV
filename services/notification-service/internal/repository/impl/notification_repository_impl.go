package impl

import (
	"context"
	"time"

	"gorm.io/gorm"

	"toxictoast/services/notification-service/internal/domain"
	"toxictoast/services/notification-service/internal/repository/entity"
	"toxictoast/services/notification-service/internal/repository/interfaces"
	"toxictoast/services/notification-service/internal/repository/mapper"
)

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) interfaces.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	e := mapper.NotificationToEntity(notification)
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *notificationRepository) GetByID(ctx context.Context, id string) (*domain.Notification, error) {
	var e entity.NotificationEntity
	err := r.db.WithContext(ctx).
		Preload("Channel").
		Preload("Attempts").
		Where("id = ?", id).
		First(&e).Error
	if err != nil {
		return nil, err
	}
	return mapper.NotificationToDomain(&e), nil
}

func (r *notificationRepository) List(ctx context.Context, channelID string, status domain.NotificationStatus, limit, offset int) ([]*domain.Notification, int64, error) {
	var entities []entity.NotificationEntity
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.NotificationEntity{})

	if channelID != "" {
		query = query.Where("channel_id = ?", channelID)
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

	if err := query.Order("created_at DESC").Preload("Channel").Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return mapper.NotificationsToDomain(entities), total, nil
}

func (r *notificationRepository) Update(ctx context.Context, notification *domain.Notification) error {
	e := mapper.NotificationToEntity(notification)
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *notificationRepository) CreateAttempt(ctx context.Context, attempt *domain.NotificationAttempt) error {
	e := mapper.NotificationAttemptToEntity(attempt)
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *notificationRepository) GetAttempts(ctx context.Context, notificationID string) ([]*domain.NotificationAttempt, error) {
	var entities []entity.NotificationAttemptEntity
	err := r.db.WithContext(ctx).
		Where("notification_id = ?", notificationID).
		Order("created_at ASC").
		Find(&entities).Error
	if err != nil {
		return nil, err
	}
	return mapper.NotificationAttemptsToDomain(entities), nil
}

func (r *notificationRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.NotificationEntity{}).Error
}

func (r *notificationRepository) CleanupOldNotifications(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)

	return r.db.WithContext(ctx).
		Where("created_at < ? AND (status = ? OR status = ?)", cutoff, domain.NotificationStatusSuccess, domain.NotificationStatusFailed).
		Delete(&entity.NotificationEntity{}).Error
}

func (r *notificationRepository) GetFailedNotifications(ctx context.Context, maxRetries int) ([]domain.Notification, error) {
	var entities []entity.NotificationEntity

	err := r.db.WithContext(ctx).
		Where("status = ? AND attempt_count < ?", domain.NotificationStatusFailed, maxRetries).
		Order("created_at ASC").
		Limit(100).
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	notifications := make([]domain.Notification, len(entities))
	for i, ent := range entities {
		notifications[i] = *mapper.NotificationToDomain(&ent)
	}

	return notifications, nil
}

func (r *notificationRepository) DeleteOldSuccessfulNotifications(ctx context.Context, cutoffDate time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("status = ? AND created_at < ?", domain.NotificationStatusSuccess, cutoffDate).
		Delete(&entity.NotificationEntity{})

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}
