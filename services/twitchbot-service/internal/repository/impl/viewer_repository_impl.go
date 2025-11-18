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

type viewerRepository struct {
	db *gorm.DB
}

// NewViewerRepository creates a new viewer repository instance
func NewViewerRepository(db *gorm.DB) interfaces.ViewerRepository {
	return &viewerRepository{db: db}
}

func (r *viewerRepository) Create(ctx context.Context, viewer *domain.Viewer) error {
	return r.db.WithContext(ctx).Create(mapper.ViewerToEntity(viewer)).Error
}

func (r *viewerRepository) GetByID(ctx context.Context, id string) (*domain.Viewer, error) {
	var viewer domain.Viewer
	err := r.db.WithContext(ctx).First(&viewer, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &viewer, nil
}

func (r *viewerRepository) GetByTwitchID(ctx context.Context, twitchID string) (*domain.Viewer, error) {
	var viewer domain.Viewer
	err := r.db.WithContext(ctx).First(&viewer, "twitch_id = ?", twitchID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &viewer, nil
}

func (r *viewerRepository) List(ctx context.Context, offset, limit int, orderBy string, includeDeleted bool) ([]*domain.Viewer, int64, error) {
	var viewers []*domain.Viewer
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.ViewerEntity{})

	if includeDeleted {
		query = query.Unscoped()
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply ordering
	switch orderBy {
	case "total_messages":
		query = query.Order("total_messages DESC")
	case "total_streams_watched":
		query = query.Order("total_streams_watched DESC")
	case "first_seen":
		query = query.Order("first_seen ASC")
	case "last_seen":
		query = query.Order("last_seen DESC")
	default:
		query = query.Order("last_seen DESC")
	}

	if err := query.Offset(offset).Limit(limit).Find(&viewers).Error; err != nil {
		return nil, 0, err
	}

	return viewers, total, nil
}

func (r *viewerRepository) Update(ctx context.Context, viewer *domain.Viewer) error {
	return r.db.WithContext(ctx).Save(mapper.ViewerToEntity(viewer)).Error
}

func (r *viewerRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.ViewerEntity{}, "id = ?", id).Error
}

func (r *viewerRepository) HardDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.ViewerEntity{}, "id = ?", id).Error
}
