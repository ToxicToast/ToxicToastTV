package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"toxictoast/services/blog-service/internal/domain"
	"toxictoast/services/blog-service/internal/repository/entity"
	"toxictoast/services/blog-service/internal/repository/mapper"
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
	e := mapper.MediaToEntity(media)
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *mediaRepository) GetByID(ctx context.Context, id string) (*domain.Media, error) {
	var e entity.MediaEntity
	err := r.db.WithContext(ctx).First(&e, "id = ?", id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("media not found")
		}
		return nil, err
	}

	return mapper.MediaToDomain(&e), nil
}

func (r *mediaRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.MediaEntity{}, "id = ?", id).Error
}

func (r *mediaRepository) List(ctx context.Context, filters MediaFilters) ([]domain.Media, int64, error) {
	var entities []entity.MediaEntity
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.MediaEntity{})

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
	if err := query.Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	// Convert to domain
	media := make([]domain.Media, 0, len(entities))
	for _, e := range entities {
		media = append(media, *mapper.MediaToDomain(&e))
	}

	return media, total, nil
}
