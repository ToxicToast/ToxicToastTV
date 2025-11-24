package impl

import (
	"context"

	"gorm.io/gorm"

	"toxictoast/services/auth-service/internal/domain"
	"toxictoast/services/auth-service/internal/repository/entity"
	"toxictoast/services/auth-service/internal/repository/interfaces"
	"toxictoast/services/auth-service/internal/repository/mapper"
)

type rolePermissionRepository struct {
	db *gorm.DB
}

// NewRolePermissionRepository creates a new role-permission repository instance
func NewRolePermissionRepository(db *gorm.DB) interfaces.RolePermissionRepository {
	return &rolePermissionRepository{db: db}
}

func (r *rolePermissionRepository) AssignPermission(ctx context.Context, rolePermission *domain.RolePermission) error {
	e := &entity.RolePermissionEntity{
		RoleID:       rolePermission.RoleID,
		PermissionID: rolePermission.PermissionID,
		CreatedAt:    rolePermission.CreatedAt,
	}
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *rolePermissionRepository) RevokePermission(ctx context.Context, roleID, permissionID string) error {
	return r.db.WithContext(ctx).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&entity.RolePermissionEntity{}).Error
}

func (r *rolePermissionRepository) GetRolePermissions(ctx context.Context, roleID string) ([]*domain.Permission, error) {
	var permissions []*entity.PermissionEntity

	err := r.db.WithContext(ctx).
		Table("azkaban_permissions").
		Joins("INNER JOIN azkaban_role_permissions ON azkaban_permissions.id = azkaban_role_permissions.permission_id").
		Where("azkaban_role_permissions.role_id = ?", roleID).
		Find(&permissions).Error

	if err != nil {
		return nil, err
	}

	return mapper.PermissionsToDomain(permissions), nil
}

func (r *rolePermissionRepository) GetPermissionRoles(ctx context.Context, permissionID string) ([]*domain.Role, error) {
	var roles []*entity.RoleEntity

	err := r.db.WithContext(ctx).
		Table("azkaban_roles").
		Joins("INNER JOIN azkaban_role_permissions ON azkaban_roles.id = azkaban_role_permissions.role_id").
		Where("azkaban_role_permissions.permission_id = ?", permissionID).
		Find(&roles).Error

	if err != nil {
		return nil, err
	}

	return mapper.RolesToDomain(roles), nil
}

func (r *rolePermissionRepository) HasPermission(ctx context.Context, roleID, permissionID string) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&entity.RolePermissionEntity{}).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *rolePermissionRepository) GetUserPermissions(ctx context.Context, userID string) ([]*domain.Permission, error) {
	var permissions []*entity.PermissionEntity

	err := r.db.WithContext(ctx).
		Table("azkaban_permissions").
		Joins("INNER JOIN azkaban_role_permissions ON azkaban_permissions.id = azkaban_role_permissions.permission_id").
		Joins("INNER JOIN azkaban_user_roles ON azkaban_role_permissions.role_id = azkaban_user_roles.role_id").
		Where("azkaban_user_roles.user_id = ?", userID).
		Distinct().
		Find(&permissions).Error

	if err != nil {
		return nil, err
	}

	return mapper.PermissionsToDomain(permissions), nil
}
