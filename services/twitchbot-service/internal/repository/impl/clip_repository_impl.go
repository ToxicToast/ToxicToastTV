package impl

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/entity"
	"toxictoast/services/twitchbot-service/internal/repository/mapper"
	"toxictoast/services/twitchbot-service/internal/repository/interfaces"
)

type clipRepository struct {
	db *gorm.DB
}

// NewClipRepository creates a new clip repository instance
func NewClipRepository(db *gorm.DB) interfaces.ClipRepository {
	return &clipRepository{db: db}
}

func (r *clipRepository) Create(ctx context.Context, clip *domain.Clip) error {
	return r.db.WithContext(ctx).Create(mapper.ClipToEntity(clip)).Error
}

func (r *clipRepository) GetByID(ctx context.Context, id string) (*domain.Clip, error) {
	var clip domain.Clip
	err := r.db.WithContext(ctx).First(&clip, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &clip, nil
}

func (r *clipRepository) GetByTwitchClipID(ctx context.Context, twitchClipID string) (*domain.Clip, error) {
	var clip domain.Clip
	err := r.db.WithContext(ctx).First(&clip, "twitch_clip_id = ?", twitchClipID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &clip, nil
}

func (r *clipRepository) List(ctx context.Context, offset, limit int, streamID, orderBy string, includeDeleted bool) ([]*domain.Clip, int64, error) {
	var clips []*domain.Clip
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.ClipEntity{})

	if includeDeleted {
		query = query.Unscoped()
	}

	if streamID != "" {
		query = query.Where("stream_id = ?", streamID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply ordering
	switch orderBy {
	case "view_count":
		query = query.Order("view_count DESC")
	case "created_at_twitch":
		query = query.Order("created_at_twitch DESC")
	case "duration_seconds":
		query = query.Order("duration_seconds DESC")
	default:
		query = query.Order("created_at_twitch DESC")
	}

	if err := query.Offset(offset).Limit(limit).Find(&clips).Error; err != nil {
		return nil, 0, err
	}

	return clips, total, nil
}

func (r *clipRepository) Update(ctx context.Context, clip *domain.Clip) error {
	return r.db.WithContext(ctx).Save(mapper.ClipToEntity(clip)).Error
}

func (r *clipRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.ClipEntity{}, "id = ?", id).Error
}

func (r *clipRepository) HardDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.ClipEntity{}, "id = ?", id).Error
}
