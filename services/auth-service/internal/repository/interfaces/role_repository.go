package interfaces

import (
	"context"
	"toxictoast/services/auth-service/internal/domain"
)

// RoleRepository defines the interface for role data access
type RoleRepository interface {
	Create(ctx context.Context, role *domain.Role) error
	GetByID(ctx context.Context, id string) (*domain.Role, error)
	GetByName(ctx context.Context, name string) (*domain.Role, error)
	List(ctx context.Context, offset, limit int) ([]*domain.Role, int64, error)
	Update(ctx context.Context, role *domain.Role) error
	Delete(ctx context.Context, id string) error
	HardDelete(ctx context.Context, id string) error
}
