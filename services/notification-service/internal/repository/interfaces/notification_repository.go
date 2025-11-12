package interfaces

import (
	"context"
	"time"

	"toxictoast/services/notification-service/internal/domain"
)

type NotificationRepository interface {
	// Create creates a new notification
	Create(ctx context.Context, notification *domain.Notification) error

	// GetByID gets a notification by ID
	GetByID(ctx context.Context, id string) (*domain.Notification, error)

	// List lists notifications with filters
	List(ctx context.Context, channelID string, status domain.NotificationStatus, limit, offset int) ([]*domain.Notification, int64, error)

	// Update updates a notification
	Update(ctx context.Context, notification *domain.Notification) error

	// CreateAttempt creates a notification attempt
	CreateAttempt(ctx context.Context, attempt *domain.NotificationAttempt) error

	// GetAttempts gets all attempts for a notification
	GetAttempts(ctx context.Context, notificationID string) ([]*domain.NotificationAttempt, error)

	// Delete soft deletes a notification
	Delete(ctx context.Context, id string) error

	// CleanupOldNotifications removes notifications older than the specified duration
	CleanupOldNotifications(ctx context.Context, olderThan time.Duration) error
}
