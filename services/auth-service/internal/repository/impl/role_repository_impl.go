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

type roleRepository struct {
	db *gorm.DB
}

// NewRoleRepository creates a new role repository instance
func NewRoleRepository(db *gorm.DB) interfaces.RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(ctx context.Context, role *domain.Role) error {
	e := mapper.RoleToEntity(role)
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *roleRepository) GetByID(ctx context.Context, id string) (*domain.Role, error) {
	var e entity.RoleEntity
	err := r.db.WithContext(ctx).First(&e, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.RoleToDomain(&e), nil
}

func (r *roleRepository) GetByName(ctx context.Context, name string) (*domain.Role, error) {
	var e entity.RoleEntity
	err := r.db.WithContext(ctx).First(&e, "name = ?", name).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.RoleToDomain(&e), nil
}

func (r *roleRepository) List(ctx context.Context, offset, limit int) ([]*domain.Role, int64, error) {
	var entities []*entity.RoleEntity
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.RoleEntity{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Offset(offset).
		Limit(limit).
		Order("name ASC").
		Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return mapper.RolesToDomain(entities), total, nil
}

func (r *roleRepository) Update(ctx context.Context, role *domain.Role) error {
	e := mapper.RoleToEntity(role)
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *roleRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.RoleEntity{}, "id = ?", id).Error
}

func (r *roleRepository) HardDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.RoleEntity{}, "id = ?", id).Error
}
