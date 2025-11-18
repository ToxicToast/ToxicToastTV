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

type discordChannelRepository struct {
	db *gorm.DB
}

func NewDiscordChannelRepository(db *gorm.DB) interfaces.DiscordChannelRepository {
	return &discordChannelRepository{db: db}
}

func (r *discordChannelRepository) Create(ctx context.Context, channel *domain.DiscordChannel) error {
	e := mapper.DiscordChannelToEntity(channel)
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *discordChannelRepository) GetByID(ctx context.Context, id string) (*domain.DiscordChannel, error) {
	var e entity.DiscordChannelEntity
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&e).Error
	if err != nil {
		return nil, err
	}
	return mapper.DiscordChannelToDomain(&e), nil
}

func (r *discordChannelRepository) GetByWebhookURL(ctx context.Context, webhookURL string) (*domain.DiscordChannel, error) {
	var e entity.DiscordChannelEntity
	err := r.db.WithContext(ctx).Where("webhook_url = ?", webhookURL).First(&e).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return mapper.DiscordChannelToDomain(&e), nil
}

func (r *discordChannelRepository) List(ctx context.Context, limit, offset int, activeOnly bool) ([]*domain.DiscordChannel, int64, error) {
	var entities []entity.DiscordChannelEntity
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.DiscordChannelEntity{})
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

	return mapper.DiscordChannelsToDomain(entities), total, nil
}

func (r *discordChannelRepository) Update(ctx context.Context, channel *domain.DiscordChannel) error {
	e := mapper.DiscordChannelToEntity(channel)
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *discordChannelRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.DiscordChannelEntity{}).Error
}

func (r *discordChannelRepository) GetActiveChannelsForEvent(ctx context.Context, eventType string) ([]*domain.DiscordChannel, error) {
	var entities []entity.DiscordChannelEntity

	// Get all active channels
	err := r.db.WithContext(ctx).
		Where("active = ?", true).
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	// Convert to domain models and filter by event type match
	channels := mapper.DiscordChannelsToDomain(entities)
	result := make([]*domain.DiscordChannel, 0)
	for _, channel := range channels {
		if channel.MatchesEvent(eventType) {
			result = append(result, channel)
		}
	}

	return result, nil
}

func (r *discordChannelRepository) UpdateStatistics(ctx context.Context, id string, success bool) error {
	updates := map[string]interface{}{
		"total_notifications":   gorm.Expr("total_notifications + 1"),
		"last_notification_at": time.Now(),
	}

	if success {
		updates["success_notifications"] = gorm.Expr("success_notifications + 1")
		updates["last_success_at"] = time.Now()
	} else {
		updates["failed_notifications"] = gorm.Expr("failed_notifications + 1")
		updates["last_failure_at"] = time.Now()
	}

	return r.db.WithContext(ctx).
		Model(&entity.DiscordChannelEntity{}).
		Where("id = ?", id).
		Updates(updates).Error
}
