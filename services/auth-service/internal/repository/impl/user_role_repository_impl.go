package impl

import (
	"context"

	"gorm.io/gorm"

	"toxictoast/services/auth-service/internal/domain"
	"toxictoast/services/auth-service/internal/repository/entity"
	"toxictoast/services/auth-service/internal/repository/interfaces"
	"toxictoast/services/auth-service/internal/repository/mapper"
)

type userRoleRepository struct {
	db *gorm.DB
}

// NewUserRoleRepository creates a new user-role repository instance
func NewUserRoleRepository(db *gorm.DB) interfaces.UserRoleRepository {
	return &userRoleRepository{db: db}
}

func (r *userRoleRepository) AssignRole(ctx context.Context, userRole *domain.UserRole) error {
	e := &entity.UserRoleEntity{
		UserID:    userRole.UserID,
		RoleID:    userRole.RoleID,
		CreatedAt: userRole.CreatedAt,
	}
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *userRoleRepository) RevokeRole(ctx context.Context, userID, roleID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&entity.UserRoleEntity{}).Error
}

func (r *userRoleRepository) GetUserRoles(ctx context.Context, userID string) ([]*domain.Role, error) {
	var roles []*entity.RoleEntity

	err := r.db.WithContext(ctx).
		Table("azkaban_roles").
		Joins("INNER JOIN azkaban_user_roles ON azkaban_roles.id = azkaban_user_roles.role_id").
		Where("azkaban_user_roles.user_id = ?", userID).
		Find(&roles).Error

	if err != nil {
		return nil, err
	}

	return mapper.RolesToDomain(roles), nil
}

func (r *userRoleRepository) GetRoleUsers(ctx context.Context, roleID string) ([]string, error) {
	var userIDs []string

	err := r.db.WithContext(ctx).
		Table("azkaban_user_roles").
		Where("role_id = ?", roleID).
		Pluck("user_id", &userIDs).Error

	if err != nil {
		return nil, err
	}

	return userIDs, nil
}

func (r *userRoleRepository) HasRole(ctx context.Context, userID, roleID string) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&entity.UserRoleEntity{}).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
