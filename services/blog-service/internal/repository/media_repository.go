package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"toxictoast/services/blog-service/internal/domain"
)

type MediaRepository interface {
	Create(ctx context.Context, media *domain.Media) error
	GetByID(ctx context.Context, id string) (*domain.Media, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filters MediaFilters) ([]domain.Media, int64, error)
}

type MediaFilters struct {
	Page       int
	PageSize   int
	MimeType   *string
	UploadedBy *string
}

type mediaRepository struct {
	db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) MediaRepository {
	return &mediaRepository{db: db}
}

func (r *mediaRepository) Create(ctx context.Context, media *domain.Media) error {
	return r.db.WithContext(ctx).Create(media).Error
}

func (r *mediaRepository) GetByID(ctx context.Context, id string) (*domain.Media, error) {
	var media domain.Media
	err := r.db.WithContext(ctx).First(&media, "id = ?", id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("media not found")
		}
		return nil, err
	}

	return &media, nil
}

func (r *mediaRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Media{}, "id = ?", id).Error
}

func (r *mediaRepository) List(ctx context.Context, filters MediaFilters) ([]domain.Media, int64, error) {
	var media []domain.Media
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Media{})

	// Filter by MIME type
	if filters.MimeType != nil {
		query = query.Where("mime_type = ?", *filters.MimeType)
	}

	// Filter by uploader
	if filters.UploadedBy != nil {
		query = query.Where("uploaded_by = ?", *filters.UploadedBy)
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

	// Order by creation date (newest first)
	query = query.Order("created_at DESC")

	// Execute query
	if err := query.Find(&media).Error; err != nil {
		return nil, 0, err
	}

	return media, total, nil
}
