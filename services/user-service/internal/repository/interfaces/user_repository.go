package interfaces

import (
	"context"
	"toxictoast/services/user-service/internal/domain"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	List(ctx context.Context, offset, limit int, status *domain.UserStatus, search *string, sortBy, sortOrder string) ([]*domain.User, int64, error)
	Update(ctx context.Context, user *domain.User) error
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
	UpdateLastLogin(ctx context.Context, userID string) error
	UpdateStatus(ctx context.Context, userID string, status domain.UserStatus) error
	Delete(ctx context.Context, id string) error
	HardDelete(ctx context.Context, id string) error
}
