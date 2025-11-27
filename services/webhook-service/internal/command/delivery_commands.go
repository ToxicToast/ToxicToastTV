package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"github.com/toxictoast/toxictoastgo/shared/logger"

	"toxictoast/services/webhook-service/internal/delivery"
	"toxictoast/services/webhook-service/internal/domain"
	"toxictoast/services/webhook-service/internal/repository/interfaces"
)

type ProcessEventCommand struct {
	cqrs.BaseCommand
	Event *domain.Event `json:"event"`
}

func (c *ProcessEventCommand) CommandName() string { return "process_event" }
func (c *ProcessEventCommand) Validate() error {
	if c.Event == nil {
		return errors.New("event is required")
	}
	if c.Event.Type == "" {
		return errors.New("event type is required")
	}
	return nil
}

type RetryDeliveryCommand struct {
	cqrs.BaseCommand
}

func (c *RetryDeliveryCommand) CommandName() string { return "retry_delivery" }
func (c *RetryDeliveryCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("delivery ID is required")
	}
	return nil
}

type DeleteDeliveryCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteDeliveryCommand) CommandName() string { return "delete_delivery" }
func (c *DeleteDeliveryCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("delivery ID is required")
	}
	return nil
}

type CleanupOldDeliveriesCommand struct {
	cqrs.BaseCommand
	OlderThanDays int `json:"older_than_days"`
}

func (c *CleanupOldDeliveriesCommand) CommandName() string { return "cleanup_old_deliveries" }
func (c *CleanupOldDeliveriesCommand) Validate() error {
	if c.OlderThanDays <= 0 {
		return errors.New("older_than_days must be positive")
	}
	return nil
}

type TestWebhookCommand struct {
	cqrs.BaseCommand
	WebhookID string `json:"webhook_id"`
}

func (c *TestWebhookCommand) CommandName() string { return "test_webhook" }
func (c *TestWebhookCommand) Validate() error {
	if c.WebhookID == "" {
		return errors.New("webhook ID is required")
	}
	return nil
}

// Handlers

type ProcessEventHandler struct {
	deliveryRepo interfaces.DeliveryRepository
	webhookRepo  interfaces.WebhookRepository
	deliveryPool *delivery.Pool
}

func NewProcessEventHandler(
	deliveryRepo interfaces.DeliveryRepository,
	webhookRepo interfaces.WebhookRepository,
	deliveryPool *delivery.Pool,
) *ProcessEventHandler {
	return &ProcessEventHandler{
		deliveryRepo: deliveryRepo,
		webhookRepo:  webhookRepo,
		deliveryPool: deliveryPool,
	}
}

func (h *ProcessEventHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	processCmd := cmd.(*ProcessEventCommand)
	event := processCmd.Event

	logger.Info(fmt.Sprintf("Processing event %s (type: %s)", event.ID, event.Type))

	// Find all active webhooks that match this event type
	webhooks, err := h.webhookRepo.GetActiveWebhooksForEvent(ctx, event.Type)
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
		if err := h.deliveryRepo.Create(ctx, del); err != nil {
			logger.Error(fmt.Sprintf("Failed to create delivery for webhook %s: %v", webhook.ID, err))
			continue
		}

		// Queue for delivery
		if err := h.deliveryPool.QueueDelivery(del, webhook); err != nil {
			logger.Error(fmt.Sprintf("Failed to queue delivery %s: %v", del.ID, err))
			// Mark as failed in database
			del.Status = domain.DeliveryStatusFailed
			del.LastError = fmt.Sprintf("Failed to queue: %v", err)
			_ = h.deliveryRepo.Update(ctx, del)
			continue
		}

		logger.Info(fmt.Sprintf("Queued delivery %s for webhook %s", del.ID, webhook.ID))
	}

	return nil
}

type RetryDeliveryHandler struct {
	deliveryRepo interfaces.DeliveryRepository
}

func NewRetryDeliveryHandler(deliveryRepo interfaces.DeliveryRepository) *RetryDeliveryHandler {
	return &RetryDeliveryHandler{deliveryRepo: deliveryRepo}
}

func (h *RetryDeliveryHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	retryCmd := cmd.(*RetryDeliveryCommand)

	delivery, err := h.deliveryRepo.GetByID(ctx, retryCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("failed to get delivery: %w", err)
	}

	// Only retry failed deliveries
	if delivery.Status != domain.DeliveryStatusFailed {
		return fmt.Errorf("delivery %s is not in failed state (current: %s)", retryCmd.AggregateID, delivery.Status)
	}

	// Reset delivery for retry
	delivery.Status = domain.DeliveryStatusRetrying
	now := time.Now()
	delivery.NextRetryAt = &now
	delivery.LastError = ""

	if err := h.deliveryRepo.Update(ctx, delivery); err != nil {
		return fmt.Errorf("failed to update delivery: %w", err)
	}

	logger.Info(fmt.Sprintf("Manually retrying delivery %s", retryCmd.AggregateID))
	return nil
}

type DeleteDeliveryHandler struct {
	deliveryRepo interfaces.DeliveryRepository
}

func NewDeleteDeliveryHandler(deliveryRepo interfaces.DeliveryRepository) *DeleteDeliveryHandler {
	return &DeleteDeliveryHandler{deliveryRepo: deliveryRepo}
}

func (h *DeleteDeliveryHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteDeliveryCommand)

	if err := h.deliveryRepo.Delete(ctx, deleteCmd.AggregateID); err != nil {
		return fmt.Errorf("failed to delete delivery: %w", err)
	}

	logger.Info(fmt.Sprintf("Deleted delivery %s", deleteCmd.AggregateID))
	return nil
}

type CleanupOldDeliveriesHandler struct {
	deliveryRepo interfaces.DeliveryRepository
}

func NewCleanupOldDeliveriesHandler(deliveryRepo interfaces.DeliveryRepository) *CleanupOldDeliveriesHandler {
	return &CleanupOldDeliveriesHandler{deliveryRepo: deliveryRepo}
}

func (h *CleanupOldDeliveriesHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	cleanupCmd := cmd.(*CleanupOldDeliveriesCommand)

	duration := time.Duration(cleanupCmd.OlderThanDays) * 24 * time.Hour

	// Get count before cleanup
	beforeCount, _, err := h.deliveryRepo.List(ctx, "", "", 0, 0)
	if err != nil {
		return fmt.Errorf("failed to count deliveries: %w", err)
	}

	if err := h.deliveryRepo.CleanupOldDeliveries(ctx, duration); err != nil {
		return fmt.Errorf("failed to cleanup old deliveries: %w", err)
	}

	afterCount, _, err := h.deliveryRepo.List(ctx, "", "", 0, 0)
	if err != nil {
		return fmt.Errorf("failed to count deliveries after cleanup: %w", err)
	}

	deleted := len(beforeCount) - len(afterCount)
	logger.Info(fmt.Sprintf("Cleaned up %d old deliveries", deleted))

	return nil
}

type TestWebhookHandler struct {
	deliveryRepo interfaces.DeliveryRepository
	webhookRepo  interfaces.WebhookRepository
	deliveryPool *delivery.Pool
}

func NewTestWebhookHandler(
	deliveryRepo interfaces.DeliveryRepository,
	webhookRepo interfaces.WebhookRepository,
	deliveryPool *delivery.Pool,
) *TestWebhookHandler {
	return &TestWebhookHandler{
		deliveryRepo: deliveryRepo,
		webhookRepo:  webhookRepo,
		deliveryPool: deliveryPool,
	}
}

func (h *TestWebhookHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	testCmd := cmd.(*TestWebhookCommand)

	webhook, err := h.webhookRepo.GetByID(ctx, testCmd.WebhookID)
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
			"webhook_id": testCmd.WebhookID,
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
	if err := h.deliveryRepo.Create(ctx, del); err != nil {
		return fmt.Errorf("failed to create test delivery: %w", err)
	}

	// Queue for delivery
	if err := h.deliveryPool.QueueDelivery(del, webhook); err != nil {
		return fmt.Errorf("failed to queue test delivery: %w", err)
	}

	logger.Info(fmt.Sprintf("Sent test webhook to %s", webhook.URL))
	return nil
}
