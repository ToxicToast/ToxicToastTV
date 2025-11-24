package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"toxictoast/services/auth-service/internal/domain"
	"toxictoast/services/auth-service/internal/repository/interfaces"
)

// RoleUseCase handles role business logic
type RoleUseCase struct {
	roleRepo interfaces.RoleRepository
}

// NewRoleUseCase creates a new role use case
func NewRoleUseCase(roleRepo interfaces.RoleRepository) *RoleUseCase {
	return &RoleUseCase{
		roleRepo: roleRepo,
	}
}

// CreateRole creates a new role
func (uc *RoleUseCase) CreateRole(ctx context.Context, name, description string) (*domain.Role, error) {
	// Check if role with same name exists
	existing, err := uc.roleRepo.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing role: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("role with name '%s' already exists", name)
	}

	role := &domain.Role{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := uc.roleRepo.Create(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return role, nil
}

// GetRole retrieves a role by ID
func (uc *RoleUseCase) GetRole(ctx context.Context, id string) (*domain.Role, error) {
	role, err := uc.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	if role == nil {
		return nil, fmt.Errorf("role not found")
	}
	return role, nil
}

// ListRoles retrieves all roles with pagination
func (uc *RoleUseCase) ListRoles(ctx context.Context, page, pageSize int) ([]*domain.Role, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	roles, total, err := uc.roleRepo.List(ctx, offset, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list roles: %w", err)
	}

	return roles, total, nil
}

// UpdateRole updates an existing role
func (uc *RoleUseCase) UpdateRole(ctx context.Context, id string, name, description *string) (*domain.Role, error) {
	// Get existing role
	role, err := uc.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	if role == nil {
		return nil, fmt.Errorf("role not found")
	}

	// Update fields
	if name != nil && *name != "" {
		// Check if new name conflicts with another role
		existing, err := uc.roleRepo.GetByName(ctx, *name)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing role: %w", err)
		}
		if existing != nil && existing.ID != id {
			return nil, fmt.Errorf("role with name '%s' already exists", *name)
		}
		role.Name = *name
	}

	if description != nil {
		role.Description = *description
	}

	role.UpdatedAt = time.Now()

	if err := uc.roleRepo.Update(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	return role, nil
}

// DeleteRole soft deletes a role
func (uc *RoleUseCase) DeleteRole(ctx context.Context, id string) error {
	// Check if role exists
	role, err := uc.roleRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}
	if role == nil {
		return fmt.Errorf("role not found")
	}

	return uc.roleRepo.Delete(ctx, id)
}
