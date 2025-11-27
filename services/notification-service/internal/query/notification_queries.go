package query

import (
	"context"
	"errors"
	"fmt"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"

	"toxictoast/services/notification-service/internal/domain"
	"toxictoast/services/notification-service/internal/repository/interfaces"
)

// ============================================================================
// Queries
// ============================================================================

type GetNotificationByIDQuery struct {
	cqrs.BaseQuery
	ID string `json:"id"`
}

func (q *GetNotificationByIDQuery) QueryName() string {
	return "get_notification_by_id"
}

func (q *GetNotificationByIDQuery) Validate() error {
	if q.ID == "" {
		return errors.New("notification ID is required")
	}
	return nil
}

type GetNotificationResult struct {
	Notification *domain.Notification
	Attempts     []*domain.NotificationAttempt
}

type ListNotificationsQuery struct {
	cqrs.BaseQuery
	ChannelID string                     `json:"channel_id"`
	Status    domain.NotificationStatus  `json:"status"`
	Limit     int                        `json:"limit"`
	Offset    int                        `json:"offset"`
}

func (q *ListNotificationsQuery) QueryName() string {
	return "list_notifications"
}

func (q *ListNotificationsQuery) Validate() error {
	if q.Limit < 0 {
		return errors.New("limit must be non-negative")
	}
	if q.Offset < 0 {
		return errors.New("offset must be non-negative")
	}
	return nil
}

type ListNotificationsResult struct {
	Notifications []*domain.Notification
	Total         int64
}

// ============================================================================
// Query Handlers
// ============================================================================

type GetNotificationByIDHandler struct {
	notificationRepo interfaces.NotificationRepository
}

func NewGetNotificationByIDHandler(notificationRepo interfaces.NotificationRepository) *GetNotificationByIDHandler {
	return &GetNotificationByIDHandler{
		notificationRepo: notificationRepo,
	}
}

func (h *GetNotificationByIDHandler) Handle(ctx context.Context, qry cqrs.Query) (interface{}, error) {
	query := qry.(*GetNotificationByIDQuery)

	notification, err := h.notificationRepo.GetByID(ctx, query.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	attempts, err := h.notificationRepo.GetAttempts(ctx, query.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification attempts: %w", err)
	}

	return &GetNotificationResult{
		Notification: notification,
		Attempts:     attempts,
	}, nil
}

type ListNotificationsHandler struct {
	notificationRepo interfaces.NotificationRepository
}

func NewListNotificationsHandler(notificationRepo interfaces.NotificationRepository) *ListNotificationsHandler {
	return &ListNotificationsHandler{
		notificationRepo: notificationRepo,
	}
}

func (h *ListNotificationsHandler) Handle(ctx context.Context, qry cqrs.Query) (interface{}, error) {
	query := qry.(*ListNotificationsQuery)

	notifications, total, err := h.notificationRepo.List(ctx, query.ChannelID, query.Status, query.Limit, query.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}

	return &ListNotificationsResult{
		Notifications: notifications,
		Total:         total,
	}, nil
}
