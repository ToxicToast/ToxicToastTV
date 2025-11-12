package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"toxictoast/services/webhook-service/internal/delivery"
	"toxictoast/services/webhook-service/internal/domain"
	"toxictoast/services/webhook-service/internal/repository/interfaces"

	"github.com/google/uuid"
	"github.com/toxictoast/toxictoastgo/shared/logger"
)

type DeliveryUseCase struct {
	deliveryRepo interfaces.DeliveryRepository
	webhookRepo  interfaces.WebhookRepository
	deliveryPool *delivery.Pool
}

func NewDeliveryUseCase(
	deliveryRepo interfaces.DeliveryRepository,
	webhookRepo interfaces.WebhookRepository,
	deliveryPool *delivery.Pool,
) *DeliveryUseCase {
	return &DeliveryUseCase{
		deliveryRepo: deliveryRepo,
		webhookRepo:  webhookRepo,
		deliveryPool: deliveryPool,
	}
}

// ProcessEvent processes an event and queues deliveries to matching webhooks
func (uc *DeliveryUseCase) ProcessEvent(ctx context.Context, event *domain.Event) error {
	logger.Info(fmt.Sprintf("Processing event %s (type: %s)", event.ID, event.Type))

	// Find all active webhooks that match this event type
	webhooks, err := uc.webhookRepo.GetActiveWebhooksForEvent(ctx, event.Type)
	if err != nil {
		return fmt.Errorf("failed to get webhooks for event: %w", err)
	}

	if len(webhooks) == 0 {
		logger.Info(fmt.Sprintf("No webhooks found for event type %s", event.Type))
		return nil
	}

	logger.Info(fmt.Sprintf("Found %d webhooks for event %s", len(webhooks), event.Type))

	// Convert event to JSON payload
	payloadBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	payload := string(payloadBytes)

	// Create and queue deliveries for each webhook
	for _, webhook := range webhooks {
		del := &domain.Delivery{
			ID:           uuid.New().String(),
			WebhookID:    webhook.ID,
			EventID:      event.ID,
			EventType:    event.Type,
			EventPayload: payload,
			Status:       domain.DeliveryStatusPending,
			AttemptCount: 0,
			Webhook:      webhook, // Include webhook for pool
		}

		// Save delivery to database
		if err := uc.deliveryRepo.Create(ctx, del); err != nil {
			logger.Error(fmt.Sprintf("Failed to create delivery for webhook %s: %v", webhook.ID, err))
			continue
		}

		// Queue for delivery
		if err := uc.deliveryPool.QueueDelivery(del, webhook); err != nil {
			logger.Error(fmt.Sprintf("Failed to queue delivery %s: %v", del.ID, err))
			// Mark as failed in database
			del.Status = domain.DeliveryStatusFailed
			del.LastError = fmt.Sprintf("Failed to queue: %v", err)
			_ = uc.deliveryRepo.Update(ctx, del)
			continue
		}

		logger.Info(fmt.Sprintf("Queued delivery %s for webhook %s", del.ID, webhook.ID))
	}

	return nil
}

// GetDelivery gets a delivery by ID with all attempts
func (uc *DeliveryUseCase) GetDelivery(ctx context.Context, id string) (*domain.Delivery, []*domain.DeliveryAttempt, error) {
	delivery, err := uc.deliveryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get delivery: %w", err)
	}

	attempts, err := uc.deliveryRepo.GetAttempts(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get delivery attempts: %w", err)
	}

	return delivery, attempts, nil
}

// ListDeliveries lists deliveries with filters
func (uc *DeliveryUseCase) ListDeliveries(ctx context.Context, webhookID string, status domain.DeliveryStatus, limit, offset int) ([]*domain.Delivery, int64, error) {
	deliveries, total, err := uc.deliveryRepo.List(ctx, webhookID, status, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list deliveries: %w", err)
	}
	return deliveries, total, nil
}

// RetryDelivery manually retries a failed delivery
func (uc *DeliveryUseCase) RetryDelivery(ctx context.Context, id string) error {
	delivery, err := uc.deliveryRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get delivery: %w", err)
	}

	// Only retry failed deliveries
	if delivery.Status != domain.DeliveryStatusFailed {
		return fmt.Errorf("delivery %s is not in failed state (current: %s)", id, delivery.Status)
	}

	// Reset delivery for retry
	delivery.Status = domain.DeliveryStatusRetrying
	now := time.Now()
	delivery.NextRetryAt = &now
	delivery.LastError = ""

	if err := uc.deliveryRepo.Update(ctx, delivery); err != nil {
		return fmt.Errorf("failed to update delivery: %w", err)
	}

	logger.Info(fmt.Sprintf("Manually retrying delivery %s", id))
	return nil
}

// DeleteDelivery soft deletes a delivery
func (uc *DeliveryUseCase) DeleteDelivery(ctx context.Context, id string) error {
	if err := uc.deliveryRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete delivery: %w", err)
	}

	logger.Info(fmt.Sprintf("Deleted delivery %s", id))
	return nil
}

// CleanupOldDeliveries removes old completed/failed deliveries
func (uc *DeliveryUseCase) CleanupOldDeliveries(ctx context.Context, olderThanDays int) (int, error) {
	if olderThanDays <= 0 {
		return 0, fmt.Errorf("olderThanDays must be positive")
	}

	duration := time.Duration(olderThanDays) * 24 * time.Hour

	// Get count before cleanup (we'll need to query manually since GORM Delete doesn't return count easily)
	beforeCount, _, err := uc.deliveryRepo.List(ctx, "", "", 0, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to count deliveries: %w", err)
	}

	if err := uc.deliveryRepo.CleanupOldDeliveries(ctx, duration); err != nil {
		return 0, fmt.Errorf("failed to cleanup old deliveries: %w", err)
	}

	afterCount, _, err := uc.deliveryRepo.List(ctx, "", "", 0, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to count deliveries after cleanup: %w", err)
	}

	deleted := len(beforeCount) - len(afterCount)
	logger.Info(fmt.Sprintf("Cleaned up %d old deliveries", deleted))

	return deleted, nil
}

// GetQueueStatus returns the current delivery queue status
func (uc *DeliveryUseCase) GetQueueStatus() (deliveryQueueSize, retryQueueSize int) {
	return uc.deliveryPool.GetQueueStatus()
}

// TestWebhook sends a test event to a webhook
func (uc *DeliveryUseCase) TestWebhook(ctx context.Context, webhookID string) error {
	webhook, err := uc.webhookRepo.GetByID(ctx, webhookID)
	if err != nil {
		return fmt.Errorf("failed to get webhook: %w", err)
	}

	// Create test event
	testEvent := &domain.Event{
		ID:        uuid.New().String(),
		Type:      "test.webhook",
		Source:    "webhook-service",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"message":    "This is a test webhook delivery",
			"webhook_id": webhookID,
			"test":       true,
		},
	}

	payloadBytes, err := json.Marshal(testEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal test event: %w", err)
	}

	// Create delivery
	del := &domain.Delivery{
		ID:           uuid.New().String(),
		WebhookID:    webhook.ID,
		EventID:      testEvent.ID,
		EventType:    testEvent.Type,
		EventPayload: string(payloadBytes),
		Status:       domain.DeliveryStatusPending,
		AttemptCount: 0,
		Webhook:      webhook,
	}

	// Save to database
	if err := uc.deliveryRepo.Create(ctx, del); err != nil {
		return fmt.Errorf("failed to create test delivery: %w", err)
	}

	// Queue for delivery
	if err := uc.deliveryPool.QueueDelivery(del, webhook); err != nil {
		return fmt.Errorf("failed to queue test delivery: %w", err)
	}

	logger.Info(fmt.Sprintf("Sent test webhook to %s", webhook.URL))
	return nil
}
