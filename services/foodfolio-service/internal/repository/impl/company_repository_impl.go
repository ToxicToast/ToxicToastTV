package impl

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/entity"
	"toxictoast/services/foodfolio-service/internal/repository/interfaces"
	"toxictoast/services/foodfolio-service/internal/repository/mapper"
)

type companyRepository struct {
	db *gorm.DB
}

// NewCompanyRepository creates a new company repository instance
func NewCompanyRepository(db *gorm.DB) interfaces.CompanyRepository {
	return &companyRepository{db: db}
}

func (r *companyRepository) Create(ctx context.Context, company *domain.Company) error {
	// Generate slug from name
	company.Slug = generateSlug(company.Name)
	e := mapper.CompanyToEntity(company)
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *companyRepository) GetByID(ctx context.Context, id string) (*domain.Company, error) {
	var e entity.CompanyEntity
	err := r.db.WithContext(ctx).First(&e, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.CompanyToDomain(&e), nil
}

func (r *companyRepository) GetBySlug(ctx context.Context, slug string) (*domain.Company, error) {
	var e entity.CompanyEntity
	err := r.db.WithContext(ctx).First(&e, "slug = ?", slug).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.CompanyToDomain(&e), nil
}

func (r *companyRepository) List(ctx context.Context, offset, limit int, search string, includeDeleted bool) ([]*domain.Company, int64, error) {
	var entities []*entity.CompanyEntity
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.CompanyEntity{})

	// Include deleted if requested
	if includeDeleted {
		query = query.Unscoped()
	}

	// Search filter
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch with pagination
	if err := query.Offset(offset).Limit(limit).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return mapper.CompaniesToDomain(entities), total, nil
}

func (r *companyRepository) Update(ctx context.Context, company *domain.Company) error {
	// Regenerate slug if name changed
	company.Slug = generateSlug(company.Name)
	e := mapper.CompanyToEntity(company)
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *companyRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.CompanyEntity{}, "id = ?", id).Error
}

func (r *companyRepository) HardDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.CompanyEntity{}, "id = ?", id).Error
}

// generateSlug creates a URL-friendly slug from a string
func generateSlug(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	// Remove special characters (basic implementation)
	// In production, consider using a proper slug library
	return s
}
