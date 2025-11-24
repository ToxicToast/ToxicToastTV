package interfaces

import (
	"context"
	"toxictoast/services/auth-service/internal/domain"
)

// PermissionRepository defines the interface for permission data access
type PermissionRepository interface {
	Create(ctx context.Context, permission *domain.Permission) error
	GetByID(ctx context.Context, id string) (*domain.Permission, error)
	GetByResourceAction(ctx context.Context, resource, action string) (*domain.Permission, error)
	List(ctx context.Context, offset, limit int, resource *string) ([]*domain.Permission, int64, error)
	Update(ctx context.Context, permission *domain.Permission) error
	Delete(ctx context.Context, id string) error
	HardDelete(ctx context.Context, id string) error
}
