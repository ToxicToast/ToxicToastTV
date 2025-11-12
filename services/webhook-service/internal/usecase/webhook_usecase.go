package usecase

import (
	"context"
	"fmt"
	"strings"

	"toxictoast/services/webhook-service/internal/domain"
	"toxictoast/services/webhook-service/internal/repository/interfaces"

	"github.com/google/uuid"
	"github.com/toxictoast/toxictoastgo/shared/logger"
)

type WebhookUseCase struct {
	webhookRepo interfaces.WebhookRepository
}

func NewWebhookUseCase(
	webhookRepo interfaces.WebhookRepository,
) *WebhookUseCase {
	return &WebhookUseCase{
		webhookRepo: webhookRepo,
	}
}

// CreateWebhook creates a new webhook
func (uc *WebhookUseCase) CreateWebhook(ctx context.Context, url, secret string, eventTypes []string, description string) (*domain.Webhook, error) {
	// Validate URL
	if url == "" {
		return nil, fmt.Errorf("webhook URL is required")
	}

	// Check if webhook with this URL already exists
	existing, err := uc.webhookRepo.GetByURL(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing webhook: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("webhook with URL %s already exists", url)
	}

	// Generate secret if not provided
	if secret == "" {
		secret = uuid.New().String()
	}

	// Join event types
	eventTypesStr := strings.Join(eventTypes, ",")

	webhook := &domain.Webhook{
		ID:          uuid.New().String(),
		URL:         url,
		Secret:      secret,
		EventTypes:  eventTypesStr,
		Description: description,
		Active:      true,
	}

	if err := uc.webhookRepo.Create(ctx, webhook); err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}

	logger.Info(fmt.Sprintf("Created webhook %s for URL %s", webhook.ID, url))
	return webhook, nil
}

// GetWebhook gets a webhook by ID
func (uc *WebhookUseCase) GetWebhook(ctx context.Context, id string) (*domain.Webhook, error) {
	webhook, err := uc.webhookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}
	return webhook, nil
}

// ListWebhooks lists webhooks with pagination
func (uc *WebhookUseCase) ListWebhooks(ctx context.Context, limit, offset int, activeOnly bool) ([]*domain.Webhook, int64, error) {
	webhooks, total, err := uc.webhookRepo.List(ctx, limit, offset, activeOnly)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list webhooks: %w", err)
	}
	return webhooks, total, nil
}

// UpdateWebhook updates a webhook
func (uc *WebhookUseCase) UpdateWebhook(ctx context.Context, id, url, secret string, eventTypes []string, description string, active bool) (*domain.Webhook, error) {
	// Get existing webhook
	webhook, err := uc.webhookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}

	// Update fields
	if url != "" {
		webhook.URL = url
	}
	if secret != "" {
		webhook.Secret = secret
	}
	if len(eventTypes) > 0 {
		webhook.EventTypes = strings.Join(eventTypes, ",")
	}
	if description != "" {
		webhook.Description = description
	}
	webhook.Active = active

	if err := uc.webhookRepo.Update(ctx, webhook); err != nil {
		return nil, fmt.Errorf("failed to update webhook: %w", err)
	}

	logger.Info(fmt.Sprintf("Updated webhook %s", webhook.ID))
	return webhook, nil
}

// DeleteWebhook soft deletes a webhook
func (uc *WebhookUseCase) DeleteWebhook(ctx context.Context, id string) error {
	if err := uc.webhookRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}

	logger.Info(fmt.Sprintf("Deleted webhook %s", id))
	return nil
}

// ToggleWebhook toggles webhook active status
func (uc *WebhookUseCase) ToggleWebhook(ctx context.Context, id string, active bool) (*domain.Webhook, error) {
	webhook, err := uc.webhookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}

	webhook.Active = active

	if err := uc.webhookRepo.Update(ctx, webhook); err != nil {
		return nil, fmt.Errorf("failed to toggle webhook: %w", err)
	}

	status := "activated"
	if !active {
		status = "deactivated"
	}
	logger.Info(fmt.Sprintf("Webhook %s %s", webhook.ID, status))

	return webhook, nil
}

// RegenerateSecret generates a new secret for a webhook
func (uc *WebhookUseCase) RegenerateSecret(ctx context.Context, id string) (*domain.Webhook, error) {
	webhook, err := uc.webhookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}

	webhook.Secret = uuid.New().String()

	if err := uc.webhookRepo.Update(ctx, webhook); err != nil {
		return nil, fmt.Errorf("failed to regenerate secret: %w", err)
	}

	logger.Info(fmt.Sprintf("Regenerated secret for webhook %s", webhook.ID))
	return webhook, nil
}

// GetActiveWebhooksForEvent gets all active webhooks that match an event type
func (uc *WebhookUseCase) GetActiveWebhooksForEvent(ctx context.Context, eventType string) ([]*domain.Webhook, error) {
	webhooks, err := uc.webhookRepo.GetActiveWebhooksForEvent(ctx, eventType)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhooks for event: %w", err)
	}
	return webhooks, nil
}
