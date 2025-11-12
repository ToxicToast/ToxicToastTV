package usecase

import (
	"context"
	"fmt"
	"strings"

	"toxictoast/services/notification-service/internal/domain"
	"toxictoast/services/notification-service/internal/repository/interfaces"

	"github.com/google/uuid"
	"github.com/toxictoast/toxictoastgo/shared/logger"
)

type ChannelUseCase struct {
	channelRepo interfaces.DiscordChannelRepository
}

func NewChannelUseCase(channelRepo interfaces.DiscordChannelRepository) *ChannelUseCase {
	return &ChannelUseCase{
		channelRepo: channelRepo,
	}
}

// CreateChannel creates a new Discord channel
func (uc *ChannelUseCase) CreateChannel(ctx context.Context, name, webhookURL string, eventTypes []string, color int, description string) (*domain.DiscordChannel, error) {
	// Validate
	if name == "" {
		return nil, fmt.Errorf("channel name is required")
	}
	if webhookURL == "" {
		return nil, fmt.Errorf("webhook URL is required")
	}

	// Check if channel with this webhook URL already exists
	existing, err := uc.channelRepo.GetByWebhookURL(ctx, webhookURL)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing channel: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("channel with webhook URL already exists")
	}

	// Join event types
	eventTypesStr := strings.Join(eventTypes, ",")

	// Default color if not provided
	if color == 0 {
		color = 5814783 // Gray
	}

	channel := &domain.DiscordChannel{
		ID:          uuid.New().String(),
		Name:        name,
		WebhookURL:  webhookURL,
		EventTypes:  eventTypesStr,
		Color:       color,
		Description: description,
		Active:      true,
	}

	if err := uc.channelRepo.Create(ctx, channel); err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	logger.Info(fmt.Sprintf("Created Discord channel %s (%s)", channel.Name, channel.ID))
	return channel, nil
}

// GetChannel gets a channel by ID
func (uc *ChannelUseCase) GetChannel(ctx context.Context, id string) (*domain.DiscordChannel, error) {
	channel, err := uc.channelRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}
	return channel, nil
}

// ListChannels lists channels with pagination
func (uc *ChannelUseCase) ListChannels(ctx context.Context, limit, offset int, activeOnly bool) ([]*domain.DiscordChannel, int64, error) {
	channels, total, err := uc.channelRepo.List(ctx, limit, offset, activeOnly)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list channels: %w", err)
	}
	return channels, total, nil
}

// UpdateChannel updates a channel
func (uc *ChannelUseCase) UpdateChannel(ctx context.Context, id, name, webhookURL string, eventTypes []string, color int, description string, active bool) (*domain.DiscordChannel, error) {
	channel, err := uc.channelRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	// Update fields
	if name != "" {
		channel.Name = name
	}
	if webhookURL != "" {
		channel.WebhookURL = webhookURL
	}
	if len(eventTypes) > 0 {
		channel.EventTypes = strings.Join(eventTypes, ",")
	}
	if color > 0 {
		channel.Color = color
	}
	if description != "" {
		channel.Description = description
	}
	channel.Active = active

	if err := uc.channelRepo.Update(ctx, channel); err != nil {
		return nil, fmt.Errorf("failed to update channel: %w", err)
	}

	logger.Info(fmt.Sprintf("Updated Discord channel %s", channel.ID))
	return channel, nil
}

// DeleteChannel soft deletes a channel
func (uc *ChannelUseCase) DeleteChannel(ctx context.Context, id string) error {
	if err := uc.channelRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete channel: %w", err)
	}

	logger.Info(fmt.Sprintf("Deleted Discord channel %s", id))
	return nil
}

// ToggleChannel toggles channel active status
func (uc *ChannelUseCase) ToggleChannel(ctx context.Context, id string, active bool) (*domain.DiscordChannel, error) {
	channel, err := uc.channelRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	channel.Active = active

	if err := uc.channelRepo.Update(ctx, channel); err != nil {
		return nil, fmt.Errorf("failed to toggle channel: %w", err)
	}

	status := "activated"
	if !active {
		status = "deactivated"
	}
	logger.Info(fmt.Sprintf("Discord channel %s %s", channel.ID, status))

	return channel, nil
}

// GetActiveChannelsForEvent gets all active channels that match an event type
func (uc *ChannelUseCase) GetActiveChannelsForEvent(ctx context.Context, eventType string) ([]*domain.DiscordChannel, error) {
	channels, err := uc.channelRepo.GetActiveChannelsForEvent(ctx, eventType)
	if err != nil {
		return nil, fmt.Errorf("failed to get channels for event: %w", err)
	}
	return channels, nil
}
