package command

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"github.com/toxictoast/toxictoastgo/shared/logger"

	"toxictoast/services/notification-service/internal/domain"
	"toxictoast/services/notification-service/internal/repository/interfaces"
)

// ============================================================================
// Commands
// ============================================================================

type CreateChannelCommand struct {
	cqrs.BaseCommand
	Name        string   `json:"name"`
	WebhookURL  string   `json:"webhook_url"`
	EventTypes  []string `json:"event_types"`
	Color       int      `json:"color"`
	Description string   `json:"description"`
}

func (c *CreateChannelCommand) CommandName() string {
	return "create_channel"
}

func (c *CreateChannelCommand) Validate() error {
	if c.Name == "" {
		return errors.New("channel name is required")
	}
	if c.WebhookURL == "" {
		return errors.New("webhook URL is required")
	}
	return nil
}

type UpdateChannelCommand struct {
	cqrs.BaseCommand
	Name        *string   `json:"name"`
	WebhookURL  *string   `json:"webhook_url"`
	EventTypes  *[]string `json:"event_types"`
	Color       *int      `json:"color"`
	Description *string   `json:"description"`
	Active      *bool     `json:"active"`
}

func (c *UpdateChannelCommand) CommandName() string {
	return "update_channel"
}

func (c *UpdateChannelCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("channel ID is required")
	}
	return nil
}

type DeleteChannelCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteChannelCommand) CommandName() string {
	return "delete_channel"
}

func (c *DeleteChannelCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("channel ID is required")
	}
	return nil
}

type ToggleChannelCommand struct {
	cqrs.BaseCommand
	Active bool `json:"active"`
}

func (c *ToggleChannelCommand) CommandName() string {
	return "toggle_channel"
}

func (c *ToggleChannelCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("channel ID is required")
	}
	return nil
}

// ============================================================================
// Command Handlers
// ============================================================================

type CreateChannelHandler struct {
	channelRepo interfaces.DiscordChannelRepository
}

func NewCreateChannelHandler(channelRepo interfaces.DiscordChannelRepository) *CreateChannelHandler {
	return &CreateChannelHandler{
		channelRepo: channelRepo,
	}
}

func (h *CreateChannelHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	createCmd := cmd.(*CreateChannelCommand)

	// Check if channel with this webhook URL already exists
	existing, err := h.channelRepo.GetByWebhookURL(ctx, createCmd.WebhookURL)
	if err != nil {
		return fmt.Errorf("failed to check existing channel: %w", err)
	}
	if existing != nil {
		return errors.New("channel with webhook URL already exists")
	}

	// Join event types
	eventTypesStr := strings.Join(createCmd.EventTypes, ",")

	// Default color if not provided
	color := createCmd.Color
	if color == 0 {
		color = 5814783 // Gray
	}

	channel := &domain.DiscordChannel{
		ID:          uuid.New().String(),
		Name:        createCmd.Name,
		WebhookURL:  createCmd.WebhookURL,
		EventTypes:  eventTypesStr,
		Color:       color,
		Description: createCmd.Description,
		Active:      true,
	}

	if err := h.channelRepo.Create(ctx, channel); err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}

	createCmd.AggregateID = channel.ID
	logger.Info(fmt.Sprintf("Created Discord channel %s (%s)", channel.Name, channel.ID))
	return nil
}

type UpdateChannelHandler struct {
	channelRepo interfaces.DiscordChannelRepository
}

func NewUpdateChannelHandler(channelRepo interfaces.DiscordChannelRepository) *UpdateChannelHandler {
	return &UpdateChannelHandler{
		channelRepo: channelRepo,
	}
}

func (h *UpdateChannelHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateChannelCommand)

	channel, err := h.channelRepo.GetByID(ctx, updateCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	// Update fields
	if updateCmd.Name != nil {
		channel.Name = *updateCmd.Name
	}
	if updateCmd.WebhookURL != nil {
		channel.WebhookURL = *updateCmd.WebhookURL
	}
	if updateCmd.EventTypes != nil && len(*updateCmd.EventTypes) > 0 {
		channel.EventTypes = strings.Join(*updateCmd.EventTypes, ",")
	}
	if updateCmd.Color != nil && *updateCmd.Color > 0 {
		channel.Color = *updateCmd.Color
	}
	if updateCmd.Description != nil {
		channel.Description = *updateCmd.Description
	}
	if updateCmd.Active != nil {
		channel.Active = *updateCmd.Active
	}

	if err := h.channelRepo.Update(ctx, channel); err != nil {
		return fmt.Errorf("failed to update channel: %w", err)
	}

	logger.Info(fmt.Sprintf("Updated Discord channel %s", channel.ID))
	return nil
}

type DeleteChannelHandler struct {
	channelRepo interfaces.DiscordChannelRepository
}

func NewDeleteChannelHandler(channelRepo interfaces.DiscordChannelRepository) *DeleteChannelHandler {
	return &DeleteChannelHandler{
		channelRepo: channelRepo,
	}
}

func (h *DeleteChannelHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteChannelCommand)

	if err := h.channelRepo.Delete(ctx, deleteCmd.AggregateID); err != nil {
		return fmt.Errorf("failed to delete channel: %w", err)
	}

	logger.Info(fmt.Sprintf("Deleted Discord channel %s", deleteCmd.AggregateID))
	return nil
}

type ToggleChannelHandler struct {
	channelRepo interfaces.DiscordChannelRepository
}

func NewToggleChannelHandler(channelRepo interfaces.DiscordChannelRepository) *ToggleChannelHandler {
	return &ToggleChannelHandler{
		channelRepo: channelRepo,
	}
}

func (h *ToggleChannelHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	toggleCmd := cmd.(*ToggleChannelCommand)

	channel, err := h.channelRepo.GetByID(ctx, toggleCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	channel.Active = toggleCmd.Active

	if err := h.channelRepo.Update(ctx, channel); err != nil {
		return fmt.Errorf("failed to toggle channel: %w", err)
	}

	status := "activated"
	if !toggleCmd.Active {
		status = "deactivated"
	}
	logger.Info(fmt.Sprintf("Discord channel %s %s", channel.ID, status))

	return nil
}
