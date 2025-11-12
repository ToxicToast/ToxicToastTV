package impl

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/entity"
	"toxictoast/services/twitchbot-service/internal/repository/interfaces"
	"toxictoast/services/twitchbot-service/internal/repository/mapper"
)

type streamRepository struct {
	db *gorm.DB
}

// NewStreamRepository creates a new stream repository instance
func NewStreamRepository(db *gorm.DB) interfaces.StreamRepository {
	return &streamRepository{db: db}
}

func (r *streamRepository) Create(ctx context.Context, stream *domain.Stream) error {
	e := mapper.StreamToEntity(stream)
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *streamRepository) GetByID(ctx context.Context, id string) (*domain.Stream, error) {
	var e entity.StreamEntity
	err := r.db.WithContext(ctx).First(&e, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.StreamToDomain(&e), nil
}

func (r *streamRepository) List(ctx context.Context, offset, limit int, onlyActive bool, gameName string, includeDeleted bool) ([]*domain.Stream, int64, error) {
	var entities []entity.StreamEntity
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.StreamEntity{})

	if includeDeleted {
		query = query.Unscoped()
	}

	if onlyActive {
		query = query.Where("is_active = ?", true)
	}

	if gameName != "" {
		query = query.Where("game_name = ?", gameName)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("started_at DESC").Offset(offset).Limit(limit).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return mapper.StreamsToDomain(entities), total, nil
}

func (r *streamRepository) GetActive(ctx context.Context) (*domain.Stream, error) {
	var e entity.StreamEntity
	// Use Session with silent logger to avoid "record not found" logs (this is expected when no stream is active)
	err := r.db.WithContext(ctx).Session(&gorm.Session{Logger: r.db.Logger.LogMode(logger.Silent)}).
		Where("is_active = ?", true).
		First(&e).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.StreamToDomain(&e), nil
}

func (r *streamRepository) Update(ctx context.Context, stream *domain.Stream) error {
	e := mapper.StreamToEntity(stream)
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *streamRepository) EndStream(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&entity.StreamEntity{}).Where("id = ?", id).Updates(map[string]interface{}{
		"ended_at":  now,
		"is_active": false,
	}).Error
}

func (r *streamRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.StreamEntity{}, "id = ?", id).Error
}

func (r *streamRepository) HardDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.StreamEntity{}, "id = ?", id).Error
}
