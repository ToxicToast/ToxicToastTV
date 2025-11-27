package command

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"github.com/toxictoast/toxictoastgo/shared/logger"

	"toxictoast/services/webhook-service/internal/domain"
	"toxictoast/services/webhook-service/internal/repository/interfaces"
)

type CreateWebhookCommand struct {
	cqrs.BaseCommand
	URL         string   `json:"url"`
	Secret      string   `json:"secret"`
	EventTypes  []string `json:"event_types"`
	Description string   `json:"description"`
}

func (c *CreateWebhookCommand) CommandName() string { return "create_webhook" }
func (c *CreateWebhookCommand) Validate() error {
	if c.URL == "" {
		return errors.New("webhook URL is required")
	}
	return nil
}

type UpdateWebhookCommand struct {
	cqrs.BaseCommand
	URL         *string   `json:"url"`
	Secret      *string   `json:"secret"`
	EventTypes  *[]string `json:"event_types"`
	Description *string   `json:"description"`
	Active      *bool     `json:"active"`
}

func (c *UpdateWebhookCommand) CommandName() string { return "update_webhook" }
func (c *UpdateWebhookCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("webhook ID is required")
	}
	return nil
}

type DeleteWebhookCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteWebhookCommand) CommandName() string { return "delete_webhook" }
func (c *DeleteWebhookCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("webhook ID is required")
	}
	return nil
}

type ToggleWebhookCommand struct {
	cqrs.BaseCommand
	Active bool `json:"active"`
}

func (c *ToggleWebhookCommand) CommandName() string { return "toggle_webhook" }
func (c *ToggleWebhookCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("webhook ID is required")
	}
	return nil
}

type RegenerateSecretCommand struct {
	cqrs.BaseCommand
}

func (c *RegenerateSecretCommand) CommandName() string { return "regenerate_secret" }
func (c *RegenerateSecretCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("webhook ID is required")
	}
	return nil
}

// Handlers

type CreateWebhookHandler struct {
	webhookRepo interfaces.WebhookRepository
}

func NewCreateWebhookHandler(webhookRepo interfaces.WebhookRepository) *CreateWebhookHandler {
	return &CreateWebhookHandler{webhookRepo: webhookRepo}
}

func (h *CreateWebhookHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	createCmd := cmd.(*CreateWebhookCommand)

	existing, err := h.webhookRepo.GetByURL(ctx, createCmd.URL)
	if err != nil {
		return fmt.Errorf("failed to check existing webhook: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("webhook with URL %s already exists", createCmd.URL)
	}

	secret := createCmd.Secret
	if secret == "" {
		secret = uuid.New().String()
	}

	webhook := &domain.Webhook{
		ID:          uuid.New().String(),
		URL:         createCmd.URL,
		Secret:      secret,
		EventTypes:  strings.Join(createCmd.EventTypes, ","),
		Description: createCmd.Description,
		Active:      true,
	}

	if err := h.webhookRepo.Create(ctx, webhook); err != nil {
		return fmt.Errorf("failed to create webhook: %w", err)
	}

	createCmd.AggregateID = webhook.ID
	logger.Info(fmt.Sprintf("Created webhook %s for URL %s", webhook.ID, createCmd.URL))
	return nil
}

type UpdateWebhookHandler struct {
	webhookRepo interfaces.WebhookRepository
}

func NewUpdateWebhookHandler(webhookRepo interfaces.WebhookRepository) *UpdateWebhookHandler {
	return &UpdateWebhookHandler{webhookRepo: webhookRepo}
}

func (h *UpdateWebhookHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateWebhookCommand)

	webhook, err := h.webhookRepo.GetByID(ctx, updateCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("failed to get webhook: %w", err)
	}

	if updateCmd.URL != nil {
		webhook.URL = *updateCmd.URL
	}
	if updateCmd.Secret != nil {
		webhook.Secret = *updateCmd.Secret
	}
	if updateCmd.EventTypes != nil && len(*updateCmd.EventTypes) > 0 {
		webhook.EventTypes = strings.Join(*updateCmd.EventTypes, ",")
	}
	if updateCmd.Description != nil {
		webhook.Description = *updateCmd.Description
	}
	if updateCmd.Active != nil {
		webhook.Active = *updateCmd.Active
	}

	if err := h.webhookRepo.Update(ctx, webhook); err != nil {
		return fmt.Errorf("failed to update webhook: %w", err)
	}

	logger.Info(fmt.Sprintf("Updated webhook %s", webhook.ID))
	return nil
}

type DeleteWebhookHandler struct {
	webhookRepo interfaces.WebhookRepository
}

func NewDeleteWebhookHandler(webhookRepo interfaces.WebhookRepository) *DeleteWebhookHandler {
	return &DeleteWebhookHandler{webhookRepo: webhookRepo}
}

func (h *DeleteWebhookHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteWebhookCommand)

	if err := h.webhookRepo.Delete(ctx, deleteCmd.AggregateID); err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}

	logger.Info(fmt.Sprintf("Deleted webhook %s", deleteCmd.AggregateID))
	return nil
}

type ToggleWebhookHandler struct {
	webhookRepo interfaces.WebhookRepository
}

func NewToggleWebhookHandler(webhookRepo interfaces.WebhookRepository) *ToggleWebhookHandler {
	return &ToggleWebhookHandler{webhookRepo: webhookRepo}
}

func (h *ToggleWebhookHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	toggleCmd := cmd.(*ToggleWebhookCommand)

	webhook, err := h.webhookRepo.GetByID(ctx, toggleCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("failed to get webhook: %w", err)
	}

	webhook.Active = toggleCmd.Active

	if err := h.webhookRepo.Update(ctx, webhook); err != nil {
		return fmt.Errorf("failed to toggle webhook: %w", err)
	}

	status := "activated"
	if !toggleCmd.Active {
		status = "deactivated"
	}
	logger.Info(fmt.Sprintf("Webhook %s %s", webhook.ID, status))
	return nil
}

type RegenerateSecretHandler struct {
	webhookRepo interfaces.WebhookRepository
}

func NewRegenerateSecretHandler(webhookRepo interfaces.WebhookRepository) *RegenerateSecretHandler {
	return &RegenerateSecretHandler{webhookRepo: webhookRepo}
}

func (h *RegenerateSecretHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	regenCmd := cmd.(*RegenerateSecretCommand)

	webhook, err := h.webhookRepo.GetByID(ctx, regenCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("failed to get webhook: %w", err)
	}

	webhook.Secret = uuid.New().String()

	if err := h.webhookRepo.Update(ctx, webhook); err != nil {
		return fmt.Errorf("failed to regenerate secret: %w", err)
	}

	logger.Info(fmt.Sprintf("Regenerated secret for webhook %s", webhook.ID))
	return nil
}
