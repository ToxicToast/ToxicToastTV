package impl

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/interfaces"
)

type typeRepository struct {
	db *gorm.DB
}

// NewTypeRepository creates a new type repository instance
func NewTypeRepository(db *gorm.DB) interfaces.TypeRepository {
	return &typeRepository{db: db}
}

func (r *typeRepository) Create(ctx context.Context, typeEntity *domain.Type) error {
	typeEntity.Slug = generateSlug(typeEntity.Name)
	return r.db.WithContext(ctx).Create(typeEntity).Error
}

func (r *typeRepository) GetByID(ctx context.Context, id string) (*domain.Type, error) {
	var typeEntity domain.Type
	err := r.db.WithContext(ctx).First(&typeEntity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &typeEntity, nil
}

func (r *typeRepository) GetBySlug(ctx context.Context, slug string) (*domain.Type, error) {
	var typeEntity domain.Type
	err := r.db.WithContext(ctx).First(&typeEntity, "slug = ?", slug).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &typeEntity, nil
}

func (r *typeRepository) List(ctx context.Context, offset, limit int, search string, includeDeleted bool) ([]*domain.Type, int64, error) {
	var types []*domain.Type
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Type{})

	if includeDeleted {
		query = query.Unscoped()
	}

	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Find(&types).Error; err != nil {
		return nil, 0, err
	}

	return types, total, nil
}

func (r *typeRepository) Update(ctx context.Context, typeEntity *domain.Type) error {
	typeEntity.Slug = generateSlug(typeEntity.Name)
	return r.db.WithContext(ctx).Save(typeEntity).Error
}

func (r *typeRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Type{}, "id = ?", id).Error
}

func (r *typeRepository) HardDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&domain.Type{}, "id = ?", id).Error
}
