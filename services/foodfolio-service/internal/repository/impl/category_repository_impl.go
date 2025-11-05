package impl

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/interfaces"
)

type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository creates a new category repository instance
func NewCategoryRepository(db *gorm.DB) interfaces.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category *domain.Category) error {
	category.Slug = generateSlug(category.Name)
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *categoryRepository) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	var category domain.Category
	err := r.db.WithContext(ctx).First(&category, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) GetBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	var category domain.Category
	err := r.db.WithContext(ctx).First(&category, "slug = ?", slug).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) List(ctx context.Context, offset, limit int, parentID *string, includeChildren bool, includeDeleted bool) ([]*domain.Category, int64, error) {
	var categories []*domain.Category
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Category{})

	if includeDeleted {
		query = query.Unscoped()
	}

	// Filter by parent ID
	if parentID != nil {
		if *parentID == "" {
			// Root categories (no parent)
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

	if err := queryWithPagination.Find(&categories).Error; err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

func (r *categoryRepository) GetTree(ctx context.Context, rootID *string, maxDepth int) ([]*domain.Category, error) {
	var categories []*domain.Category

	query := r.db.WithContext(ctx)

	if rootID != nil && *rootID != "" {
		// Get subtree starting from specific root
		query = query.Where("id = ?", *rootID)
	} else {
		// Get all root categories
		query = query.Where("parent_id IS NULL")
	}

	// Preload children recursively
	// Note: GORM doesn't support recursive preloading with depth limit out of the box
	// This is a simplified version - for production, consider using recursive CTE or separate queries
	if maxDepth == 0 || maxDepth >= 1 {
		query = query.Preload("Children")
		if maxDepth == 0 || maxDepth >= 2 {
			query = query.Preload("Children.Children")
			if maxDepth == 0 || maxDepth >= 3 {
				query = query.Preload("Children.Children.Children")
			}
		}
	}

	if err := query.Find(&categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *categoryRepository) GetChildren(ctx context.Context, parentID string) ([]*domain.Category, error) {
	var categories []*domain.Category

	err := r.db.WithContext(ctx).
		Where("parent_id = ?", parentID).
		Find(&categories).Error

	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *categoryRepository) GetRootCategories(ctx context.Context) ([]*domain.Category, error) {
	var categories []*domain.Category

	err := r.db.WithContext(ctx).
		Where("parent_id IS NULL").
		Find(&categories).Error

	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *categoryRepository) Update(ctx context.Context, category *domain.Category) error {
	category.Slug = generateSlug(category.Name)
	return r.db.WithContext(ctx).Save(category).Error
}

func (r *categoryRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Category{}, "id = ?", id).Error
}

func (r *categoryRepository) HardDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&domain.Category{}, "id = ?", id).Error
}
