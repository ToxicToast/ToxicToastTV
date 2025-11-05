package impl

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/interfaces"
)

type locationRepository struct {
	db *gorm.DB
}

// NewLocationRepository creates a new location repository instance
func NewLocationRepository(db *gorm.DB) interfaces.LocationRepository {
	return &locationRepository{db: db}
}

func (r *locationRepository) Create(ctx context.Context, location *domain.Location) error {
	location.Slug = generateSlug(location.Name)
	return r.db.WithContext(ctx).Create(location).Error
}

func (r *locationRepository) GetByID(ctx context.Context, id string) (*domain.Location, error) {
	var location domain.Location
	err := r.db.WithContext(ctx).First(&location, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &location, nil
}

func (r *locationRepository) GetBySlug(ctx context.Context, slug string) (*domain.Location, error) {
	var location domain.Location
	err := r.db.WithContext(ctx).First(&location, "slug = ?", slug).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &location, nil
}

func (r *locationRepository) List(ctx context.Context, offset, limit int, parentID *string, includeChildren bool, includeDeleted bool) ([]*domain.Location, int64, error) {
	var locations []*domain.Location
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Location{})

	if includeDeleted {
		query = query.Unscoped()
	}

	// Filter by parent ID
	if parentID != nil {
		if *parentID == "" {
			// Root locations (no parent)
			query = query.Where("parent_id IS NULL")
		} else {
			query = query.Where("parent_id = ?", *parentID)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	queryWithPagination := query.Offset(offset).Limit(limit)

	if includeChildren {
		queryWithPagination = queryWithPagination.Preload("Children")
	}

	if err := queryWithPagination.Find(&locations).Error; err != nil {
		return nil, 0, err
	}

	return locations, total, nil
}

func (r *locationRepository) GetTree(ctx context.Context, rootID *string, maxDepth int) ([]*domain.Location, error) {
	var locations []*domain.Location

	query := r.db.WithContext(ctx)

	if rootID != nil && *rootID != "" {
		// Get subtree starting from specific root
		query = query.Where("id = ?", *rootID)
	} else {
		// Get all root locations
		query = query.Where("parent_id IS NULL")
	}

	// Preload children recursively
	if maxDepth == 0 || maxDepth >= 1 {
		query = query.Preload("Children")
		if maxDepth == 0 || maxDepth >= 2 {
			query = query.Preload("Children.Children")
			if maxDepth == 0 || maxDepth >= 3 {
				query = query.Preload("Children.Children.Children")
			}
		}
	}

	if err := query.Find(&locations).Error; err != nil {
		return nil, err
	}

	return locations, nil
}

func (r *locationRepository) GetChildren(ctx context.Context, parentID string) ([]*domain.Location, error) {
	var locations []*domain.Location

	err := r.db.WithContext(ctx).
		Where("parent_id = ?", parentID).
		Find(&locations).Error

	if err != nil {
		return nil, err
	}

	return locations, nil
}

func (r *locationRepository) GetRootLocations(ctx context.Context) ([]*domain.Location, error) {
	var locations []*domain.Location

	err := r.db.WithContext(ctx).
		Where("parent_id IS NULL").
		Find(&locations).Error

	if err != nil {
		return nil, err
	}

	return locations, nil
}

func (r *locationRepository) Update(ctx context.Context, location *domain.Location) error {
	location.Slug = generateSlug(location.Name)
	return r.db.WithContext(ctx).Save(location).Error
}

func (r *locationRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Location{}, "id = ?", id).Error
}

func (r *locationRepository) HardDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&domain.Location{}, "id = ?", id).Error
}
