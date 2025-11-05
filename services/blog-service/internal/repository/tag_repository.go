package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"toxictoast/services/blog-service/internal/domain"
)

type TagRepository interface {
	Create(ctx context.Context, tag *domain.Tag) error
	GetByID(ctx context.Context, id string) (*domain.Tag, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Tag, error)
	Update(ctx context.Context, tag *domain.Tag) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filters TagFilters) ([]domain.Tag, int64, error)
	SlugExists(ctx context.Context, slug string) (bool, error)
	GetByIDs(ctx context.Context, ids []string) ([]domain.Tag, error)
}

type TagFilters struct {
	Page     int
	PageSize int
	Search   *string
}

type tagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepository{db: db}
}

func (r *tagRepository) Create(ctx context.Context, tag *domain.Tag) error {
	return r.db.WithContext(ctx).Create(tag).Error
}

func (r *tagRepository) GetByID(ctx context.Context, id string) (*domain.Tag, error) {
	var tag domain.Tag
	err := r.db.WithContext(ctx).First(&tag, "id = ?", id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tag not found")
		}
		return nil, err
	}

	return &tag, nil
}

func (r *tagRepository) GetBySlug(ctx context.Context, slug string) (*domain.Tag, error) {
	var tag domain.Tag
	err := r.db.WithContext(ctx).First(&tag, "slug = ?", slug).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tag not found")
		}
		return nil, err
	}

	return &tag, nil
}

func (r *tagRepository) Update(ctx context.Context, tag *domain.Tag) error {
	return r.db.WithContext(ctx).Save(tag).Error
}

func (r *tagRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Tag{}, "id = ?", id).Error
}

func (r *tagRepository) List(ctx context.Context, filters TagFilters) ([]domain.Tag, int64, error) {
	var tags []domain.Tag
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Tag{})

	// Search filter
	if filters.Search != nil && *filters.Search != "" {
		searchTerm := "%" + *filters.Search + "%"
		query = query.Where("name ILIKE ?", searchTerm)
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

	// Order by name
	query = query.Order("name ASC")

	// Execute query
	if err := query.Find(&tags).Error; err != nil {
		return nil, 0, err
	}

	return tags, total, nil
}

func (r *tagRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Tag{}).Where("slug = ?", slug).Count(&count).Error
	return count > 0, err
}

func (r *tagRepository) GetByIDs(ctx context.Context, ids []string) ([]domain.Tag, error) {
	var tags []domain.Tag
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&tags).Error
	return tags, err
}
