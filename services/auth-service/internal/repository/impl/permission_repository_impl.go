package impl

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"toxictoast/services/auth-service/internal/domain"
	"toxictoast/services/auth-service/internal/repository/entity"
	"toxictoast/services/auth-service/internal/repository/interfaces"
	"toxictoast/services/auth-service/internal/repository/mapper"
)

type permissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository creates a new permission repository instance
func NewPermissionRepository(db *gorm.DB) interfaces.PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) Create(ctx context.Context, permission *domain.Permission) error {
	e := mapper.PermissionToEntity(permission)
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *permissionRepository) GetByID(ctx context.Context, id string) (*domain.Permission, error) {
	var e entity.PermissionEntity
	err := r.db.WithContext(ctx).First(&e, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.PermissionToDomain(&e), nil
}

func (r *permissionRepository) GetByResourceAction(ctx context.Context, resource, action string) (*domain.Permission, error) {
	var e entity.PermissionEntity
	err := r.db.WithContext(ctx).
		Where("resource = ? AND action = ?", resource, action).
		First(&e).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.PermissionToDomain(&e), nil
}

func (r *permissionRepository) List(ctx context.Context, offset, limit int, resource *string) ([]*domain.Permission, int64, error) {
	var entities []*entity.PermissionEntity
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.PermissionEntity{})

	if resource != nil && *resource != "" {
		query = query.Where("resource = ?", *resource)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Offset(offset).
		Limit(limit).
		Order("resource ASC, action ASC").
		Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return mapper.PermissionsToDomain(entities), total, nil
}

func (r *permissionRepository) Update(ctx context.Context, permission *domain.Permission) error {
	e := mapper.PermissionToEntity(permission)
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *permissionRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.PermissionEntity{}, "id = ?", id).Error
}

func (r *permissionRepository) HardDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.PermissionEntity{}, "id = ?", id).Error
}
