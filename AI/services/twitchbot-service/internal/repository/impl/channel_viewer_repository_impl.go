package impl

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/entity"
	"toxictoast/services/twitchbot-service/internal/repository/mapper"
	"toxictoast/services/twitchbot-service/internal/repository/interfaces"
)

type channelViewerRepository struct {
	db *gorm.DB
}

// NewChannelViewerRepository creates a new channel viewer repository instance
func NewChannelViewerRepository(db *gorm.DB) interfaces.ChannelViewerRepository {
	return &channelViewerRepository{db: db}
}

func (r *channelViewerRepository) Upsert(ctx context.Context, channelViewer *domain.ChannelViewer) error {
	// Upsert: Insert or update on conflict
	e := mapper.ChannelViewerToEntity(channelViewer)
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "channel"}, {Name: "twitch_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"username", "display_name", "last_seen", "is_moderator", "is_vip", "updated_at"}),
	}).Create(e).Error
}

func (r *channelViewerRepository) GetByChannelAndTwitchID(ctx context.Context, channel, twitchID string) (*domain.ChannelViewer, error) {
	var e entity.ChannelViewerEntity
	err := r.db.WithContext(ctx).Where("channel = ? AND twitch_id = ?", channel, twitchID).First(&e).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ChannelViewerToDomain(&e), nil
}

func (r *channelViewerRepository) ListByChannel(ctx context.Context, channel string, limit, offset int) ([]*domain.ChannelViewer, int64, error) {
	var entities []entity.ChannelViewerEntity
	var total int64

	// Count total
	if err := r.db.WithContext(ctx).Model(&entity.ChannelViewerEntity{}).Where("channel = ?", channel).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := r.db.WithContext(ctx).Where("channel = ?", channel)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Order("last_seen DESC").Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return mapper.ChannelViewersToDomain(entities), total, nil
}

func (r *channelViewerRepository) UpdateLastSeen(ctx context.Context, channel, twitchID string) error {
	return r.db.WithContext(ctx).
		Model(&entity.ChannelViewerEntity{}).
		Where("channel = ? AND twitch_id = ?", channel, twitchID).
		Update("last_seen", time.Now()).
		Error
}

func (r *channelViewerRepository) Delete(ctx context.Context, channel, twitchID string) error {
	return r.db.WithContext(ctx).
		Where("channel = ? AND twitch_id = ?", channel, twitchID).
		Delete(&entity.ChannelViewerEntity{}).
		Error
}

func (r *channelViewerRepository) CountByChannel(ctx context.Context, channel string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.ChannelViewerEntity{}).
		Where("channel = ?", channel).
		Count(&count).
		Error
	return count, err
}
