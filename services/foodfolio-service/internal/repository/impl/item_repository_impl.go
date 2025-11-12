package impl

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/entity"
	"toxictoast/services/foodfolio-service/internal/repository/interfaces"
	"toxictoast/services/foodfolio-service/internal/repository/mapper"
)

type itemRepository struct {
	db *gorm.DB
}

// NewItemRepository creates a new item repository instance
func NewItemRepository(db *gorm.DB) interfaces.ItemRepository {
	return &itemRepository{db: db}
}

func (r *itemRepository) Create(ctx context.Context, item *domain.Item) error {
	item.Slug = generateSlug(item.Name)
	e := mapper.ItemToEntity(item)
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *itemRepository) GetByID(ctx context.Context, id string) (*domain.Item, error) {
	var e entity.ItemEntity
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Company").
		Preload("Type").
		First(&e, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ItemToDomain(&e), nil
}

func (r *itemRepository) GetBySlug(ctx context.Context, slug string) (*domain.Item, error) {
	var e entity.ItemEntity
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Company").
		Preload("Type").
		First(&e, "slug = ?", slug).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ItemToDomain(&e), nil
}

func (r *itemRepository) GetWithVariants(ctx context.Context, id string, includeDetails bool) (*domain.Item, error) {
	var e entity.ItemEntity

	query := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Company").
		Preload("Type").
		Preload("ItemVariants").
		Preload("ItemVariants.Size")

	if includeDetails {
		query = query.Preload("ItemVariants.ItemDetails").
			Preload("ItemVariants.ItemDetails.Warehouse").
			Preload("ItemVariants.ItemDetails.Location")
	}

	err := query.First(&e, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return mapper.ItemToDomain(&e), nil
}

func (r *itemRepository) List(ctx context.Context, offset, limit int, categoryID, companyID, typeID, search *string, includeDeleted bool) ([]*domain.Item, int64, error) {
	var entities []*entity.ItemEntity
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.ItemEntity{})

	if includeDeleted {
		query = query.Unscoped()
	}

	// Filters
	if categoryID != nil && *categoryID != "" {
		query = query.Where("category_id = ?", *categoryID)
	}

	if companyID != nil && *companyID != "" {
		query = query.Where("company_id = ?", *companyID)
	}

	if typeID != nil && *typeID != "" {
		query = query.Where("type_id = ?", *typeID)
	}

	if search != nil && *search != "" {
		query = query.Where("name ILIKE ?", "%"+*search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Preload("Category").
		Preload("Company").
		Preload("Type").
		Offset(offset).
		Limit(limit).
		Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return mapper.ItemsToDomain(entities), total, nil
}

func (r *itemRepository) Search(ctx context.Context, query string, offset, limit int, categoryID, companyID *string) ([]*domain.Item, int64, error) {
	var entities []*entity.ItemEntity
	var total int64

	dbQuery := r.db.WithContext(ctx).Model(&entity.ItemEntity{})

	// Search in name
	dbQuery = dbQuery.Where("name ILIKE ?", "%"+query+"%")

	// Optional filters
	if categoryID != nil && *categoryID != "" {
		dbQuery = dbQuery.Where("category_id = ?", *categoryID)
	}

	if companyID != nil && *companyID != "" {
		dbQuery = dbQuery.Where("company_id = ?", *companyID)
	}

	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := dbQuery.
		Preload("Category").
		Preload("Company").
		Preload("Type").
		Offset(offset).
		Limit(limit).
		Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return mapper.ItemsToDomain(entities), total, nil
}

func (r *itemRepository) Update(ctx context.Context, item *domain.Item) error {
	item.Slug = generateSlug(item.Name)
	e := mapper.ItemToEntity(item)
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *itemRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.ItemEntity{}, "id = ?", id).Error
}

func (r *itemRepository) HardDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.ItemEntity{}, "id = ?", id).Error
}
