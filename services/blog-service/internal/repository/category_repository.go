package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"toxictoast/services/blog-service/internal/domain"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) error
	GetByID(ctx context.Context, id string) (*domain.Category, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Category, error)
	Update(ctx context.Context, category *domain.Category) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filters CategoryFilters) ([]domain.Category, int64, error)
	SlugExists(ctx context.Context, slug string) (bool, error)
	GetChildren(ctx context.Context, parentID string) ([]domain.Category, error)
}

type CategoryFilters struct {
	Page     int
	PageSize int
	ParentID *string
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category *domain.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *categoryRepository) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	var category domain.Category
	err := r.db.WithContext(ctx).
		Preload("Parent").
		Preload("Children").
		First(&category, "id = ?", id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("category not found")
		}
		return nil, err
	}

	return &category, nil
}

func (r *categoryRepository) GetBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	var category domain.Category
	err := r.db.WithContext(ctx).
		Preload("Parent").
		Preload("Children").
		First(&category, "slug = ?", slug).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("category not found")
		}
		return nil, err
	}

	return &category, nil
}

func (r *categoryRepository) Update(ctx context.Context, category *domain.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

func (r *categoryRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Category{}, "id = ?", id).Error
}

func (r *categoryRepository) List(ctx context.Context, filters CategoryFilters) ([]domain.Category, int64, error) {
	var categories []domain.Category
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Category{})

	// Filter by parent
	if filters.ParentID != nil {
		query = query.Where("parent_id = ?", *filters.ParentID)
	} else {
		// Default: only root categories (no parent)
		query = query.Where("parent_id IS NULL")
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	if filters.PageSize > 0 {
		offset := (filters.Page - 1) * filters.PageSize
		query = query.Offset(offset).Limit(filters.PageSize)
	}

	// Preload associations
	query = query.Preload("Parent").Preload("Children")

	// Order by name
	query = query.Order("name ASC")

	// Execute query
	if err := query.Find(&categories).Error; err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

func (r *categoryRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Category{}).Where("slug = ?", slug).Count(&count).Error
	return count > 0, err
}

func (r *categoryRepository) GetChildren(ctx context.Context, parentID string) ([]domain.Category, error) {
	var categories []domain.Category
	err := r.db.WithContext(ctx).
		Where("parent_id = ?", parentID).
		Order("name ASC").
		Find(&categories).Error

	return categories, err
}
