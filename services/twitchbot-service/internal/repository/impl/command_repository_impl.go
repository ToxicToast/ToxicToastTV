package impl

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/entity"
	"toxictoast/services/twitchbot-service/internal/repository/mapper"
	"toxictoast/services/twitchbot-service/internal/repository/interfaces"
)

type commandRepository struct {
	db *gorm.DB
}

// NewCommandRepository creates a new command repository instance
func NewCommandRepository(db *gorm.DB) interfaces.CommandRepository {
	return &commandRepository{db: db}
}

func (r *commandRepository) Create(ctx context.Context, command *domain.Command) error {
	return r.db.WithContext(ctx).Create(mapper.CommandToEntity(command)).Error
}

func (r *commandRepository) GetByID(ctx context.Context, id string) (*domain.Command, error) {
	var e entity.CommandEntity
	err := r.db.WithContext(ctx).First(&e, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.CommandToDomain(&e), nil
}

func (r *commandRepository) GetByName(ctx context.Context, name string) (*domain.Command, error) {
	var e entity.CommandEntity
	err := r.db.WithContext(ctx).First(&e, "name = ?", name).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.CommandToDomain(&e), nil
}

func (r *commandRepository) List(ctx context.Context, offset, limit int, onlyActive bool, includeDeleted bool) ([]*domain.Command, int64, error) {
	var entities []entity.CommandEntity
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.CommandEntity{})

	if includeDeleted {
		query = query.Unscoped()
	}

	if onlyActive {
		query = query.Where("is_active = ?", true)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("name ASC").Offset(offset).Limit(limit).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return mapper.CommandsToDomain(entities), total, nil
}

func (r *commandRepository) Update(ctx context.Context, command *domain.Command) error {
	e := mapper.CommandToEntity(command)
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *commandRepository) IncrementUsage(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&entity.CommandEntity{}).Where("id = ?", id).Updates(map[string]interface{}{
		"usage_count": gorm.Expr("usage_count + ?", 1),
		"last_used":   now,
	}).Error
}

func (r *commandRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.CommandEntity{}, "id = ?", id).Error
}

func (r *commandRepository) HardDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.CommandEntity{}, "id = ?", id).Error
}
