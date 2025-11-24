package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"toxictoast/services/auth-service/internal/domain"
	"toxictoast/services/auth-service/internal/repository/interfaces"
)

// PermissionUseCase handles permission business logic
type PermissionUseCase struct {
	permissionRepo interfaces.PermissionRepository
}

// NewPermissionUseCase creates a new permission use case
func NewPermissionUseCase(permissionRepo interfaces.PermissionRepository) *PermissionUseCase {
	return &PermissionUseCase{
		permissionRepo: permissionRepo,
	}
}

// CreatePermission creates a new permission
func (uc *PermissionUseCase) CreatePermission(ctx context.Context, resource, action, description string) (*domain.Permission, error) {
	// Check if permission with same resource:action exists
	existing, err := uc.permissionRepo.GetByResourceAction(ctx, resource, action)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing permission: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("permission '%s:%s' already exists", resource, action)
	}

	permission := &domain.Permission{
		ID:          uuid.New().String(),
		Resource:    resource,
		Action:      action,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := uc.permissionRepo.Create(ctx, permission); err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	return permission, nil
}

// GetPermission retrieves a permission by ID
func (uc *PermissionUseCase) GetPermission(ctx context.Context, id string) (*domain.Permission, error) {
	permission, err := uc.permissionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}
	if permission == nil {
		return nil, fmt.Errorf("permission not found")
	}
	return permission, nil
}

// ListPermissions retrieves all permissions with pagination
func (uc *PermissionUseCase) ListPermissions(ctx context.Context, page, pageSize int, resource *string) ([]*domain.Permission, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	permissions, total, err := uc.permissionRepo.List(ctx, offset, pageSize, resource)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list permissions: %w", err)
	}

	return permissions, total, nil
}

// UpdatePermission updates an existing permission
func (uc *PermissionUseCase) UpdatePermission(ctx context.Context, id string, resource, action, description *string) (*domain.Permission, error) {
	// Get existing permission
	permission, err := uc.permissionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}
	if permission == nil {
		return nil, fmt.Errorf("permission not found")
	}

	// Update fields
	if resource != nil && *resource != "" {
		permission.Resource = *resource
	}

	if action != nil && *action != "" {
		permission.Action = *action
	}

	// Check if updated resource:action conflicts with another permission
	if resource != nil || action != nil {
		existing, err := uc.permissionRepo.GetByResourceAction(ctx, permission.Resource, permission.Action)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing permission: %w", err)
		}
		if existing != nil && existing.ID != id {
			return nil, fmt.Errorf("permission '%s:%s' already exists", permission.Resource, permission.Action)
		}
	}

	if description != nil {
		permission.Description = *description
	}

	permission.UpdatedAt = time.Now()

	if err := uc.permissionRepo.Update(ctx, permission); err != nil {
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}

	return permission, nil
}

// DeletePermission soft deletes a permission
func (uc *PermissionUseCase) DeletePermission(ctx context.Context, id string) error {
	// Check if permission exists
	permission, err := uc.permissionRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get permission: %w", err)
	}
	if permission == nil {
		return fmt.Errorf("permission not found")
	}

	return uc.permissionRepo.Delete(ctx, id)
}
