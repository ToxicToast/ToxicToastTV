package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"toxictoast/services/notification-service/internal/discord"
	"toxictoast/services/notification-service/internal/domain"
	"toxictoast/services/notification-service/internal/repository/interfaces"

	"github.com/google/uuid"
	"github.com/toxictoast/toxictoastgo/shared/logger"
)

type NotificationUseCase struct {
	notificationRepo interfaces.NotificationRepository
	channelRepo      interfaces.DiscordChannelRepository
	discordClient    *discord.Client
}

func NewNotificationUseCase(
	notificationRepo interfaces.NotificationRepository,
	channelRepo interfaces.DiscordChannelRepository,
	discordClient *discord.Client,
) *NotificationUseCase {
	return &NotificationUseCase{
		notificationRepo: notificationRepo,
		channelRepo:      channelRepo,
		discordClient:    discordClient,
	}
}

// ProcessEvent processes an event and sends notifications to matching Discord channels
func (uc *NotificationUseCase) ProcessEvent(ctx context.Context, event *domain.Event) error {
	logger.Info(fmt.Sprintf("Processing event %s (type: %s)", event.ID, event.Type))

	// Find all active channels that match this event type
	channels, err := uc.channelRepo.GetActiveChannelsForEvent(ctx, event.Type)
	if err != nil {
		return fmt.Errorf("failed to get channels for event: %w", err)
	}

	if len(channels) == 0 {
		logger.Info(fmt.Sprintf("No Discord channels found for event type %s", event.Type))
		return nil
	}

	logger.Info(fmt.Sprintf("Found %d Discord channels for event %s", len(channels), event.Type))

	// Convert event to JSON payload
	payloadBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	payload := string(payloadBytes)

	// Send to each matching channel
	for _, channel := range channels {
		notification := &domain.Notification{
			ID:           uuid.New().String(),
			ChannelID:    channel.ID,
			EventID:      event.ID,
			EventType:    event.Type,
			EventPayload: payload,
			Status:       domain.NotificationStatusPending,
			AttemptCount: 0,
		}

		// Save notification to database
		if err := uc.notificationRepo.Create(ctx, notification); err != nil {
			logger.Error(fmt.Sprintf("Failed to create notification for channel %s: %v", channel.ID, err))
			continue
		}

		// Send notification
		uc.sendNotification(ctx, notification, channel, event)
	}

	return nil
}

// sendNotification sends a single notification to Discord
func (uc *NotificationUseCase) sendNotification(ctx context.Context, notification *domain.Notification, channel *domain.DiscordChannel, event *domain.Event) {
	startTime := time.Now()

	// Build Discord embed
	color := channel.Color
	if color == 0 {
		color = discord.GetEmbedColor(event.Type)
	}
	embed := discord.BuildEmbedFromEvent(event, color)

	// Send to Discord
	messageID, responseStatus, responseBody, err := uc.discordClient.SendNotification(ctx, channel.WebhookURL, embed)

	duration := time.Since(startTime)

	// Record attempt
	attempt := &domain.NotificationAttempt{
		ID:               uuid.New().String(),
		NotificationID:   notification.ID,
		AttemptNumber:    1,
		ResponseStatus:   responseStatus,
		ResponseBody:     responseBody,
		DiscordMessageID: messageID,
		Success:          err == nil,
		Error:            "",
		DurationMs:       int(duration.Milliseconds()),
	}

	if err != nil {
		attempt.Error = err.Error()
	}

	if err := uc.notificationRepo.CreateAttempt(ctx, attempt); err != nil {
		logger.Error(fmt.Sprintf("Failed to record notification attempt: %v", err))
	}

	// Update notification status
	notification.AttemptCount = 1
	if err == nil {
		notification.Status = domain.NotificationStatusSuccess
		notification.DiscordMessageID = messageID
		now := time.Now()
		notification.SentAt = &now
		logger.Info(fmt.Sprintf("Successfully sent notification %s to channel %s", notification.ID, channel.Name))
	} else {
		notification.Status = domain.NotificationStatusFailed
		notification.LastError = err.Error()
		logger.Error(fmt.Sprintf("Failed to send notification %s: %v", notification.ID, err))
	}

	if err := uc.notificationRepo.Update(ctx, notification); err != nil {
		logger.Error(fmt.Sprintf("Failed to update notification: %v", err))
	}

	// Update channel statistics
	if err := uc.channelRepo.UpdateStatistics(ctx, channel.ID, err == nil); err != nil {
		logger.Error(fmt.Sprintf("Failed to update channel statistics: %v", err))
	}
}

// GetNotification gets a notification by ID with all attempts
func (uc *NotificationUseCase) GetNotification(ctx context.Context, id string) (*domain.Notification, []*domain.NotificationAttempt, error) {
	notification, err := uc.notificationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get notification: %w", err)
	}

	attempts, err := uc.notificationRepo.GetAttempts(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get notification attempts: %w", err)
	}

	return notification, attempts, nil
}

// ListNotifications lists notifications with filters
func (uc *NotificationUseCase) ListNotifications(ctx context.Context, channelID string, status domain.NotificationStatus, limit, offset int) ([]*domain.Notification, int64, error) {
	notifications, total, err := uc.notificationRepo.List(ctx, channelID, status, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list notifications: %w", err)
	}
	return notifications, total, nil
}

// DeleteNotification soft deletes a notification
func (uc *NotificationUseCase) DeleteNotification(ctx context.Context, id string) error {
	if err := uc.notificationRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	logger.Info(fmt.Sprintf("Deleted notification %s", id))
	return nil
}

// CleanupOldNotifications removes old completed/failed notifications
func (uc *NotificationUseCase) CleanupOldNotifications(ctx context.Context, olderThanDays int) error {
	if olderThanDays <= 0 {
		return fmt.Errorf("olderThanDays must be positive")
	}

	duration := time.Duration(olderThanDays) * 24 * time.Hour

	if err := uc.notificationRepo.CleanupOldNotifications(ctx, duration); err != nil {
		return fmt.Errorf("failed to cleanup old notifications: %w", err)
	}

	logger.Info(fmt.Sprintf("Cleaned up notifications older than %d days", olderThanDays))
	return nil
}

// RetryNotification retries sending a failed notification
func (uc *NotificationUseCase) RetryNotification(ctx context.Context, notification *domain.Notification) error {
	// Get the Discord channel
	channel, err := uc.channelRepo.GetByID(ctx, notification.ChannelID)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	// Unmarshal event from stored payload
	var event domain.Event
	if err := json.Unmarshal([]byte(notification.EventPayload), &event); err != nil {
		return fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	startTime := time.Now()

	// Build Discord embed
	color := channel.Color
	if color == 0 {
		color = discord.GetEmbedColor(event.Type)
	}
	embed := discord.BuildEmbedFromEvent(&event, color)

	// Send to Discord
	messageID, responseStatus, responseBody, err := uc.discordClient.SendNotification(ctx, channel.WebhookURL, embed)

	duration := time.Since(startTime)

	// Increment attempt count
	notification.AttemptCount++

	// Record attempt
	attempt := &domain.NotificationAttempt{
		ID:               uuid.New().String(),
		NotificationID:   notification.ID,
		AttemptNumber:    notification.AttemptCount,
		ResponseStatus:   responseStatus,
		ResponseBody:     responseBody,
		DiscordMessageID: messageID,
		Success:          err == nil,
		Error:            "",
		DurationMs:       int(duration.Milliseconds()),
	}

	if err != nil {
		attempt.Error = err.Error()
	}

	if err := uc.notificationRepo.CreateAttempt(ctx, attempt); err != nil {
		logger.Error(fmt.Sprintf("Failed to record retry attempt: %v", err))
	}

	// Update notification status
	if err == nil {
		notification.Status = domain.NotificationStatusSuccess
		notification.DiscordMessageID = messageID
		now := time.Now()
		notification.SentAt = &now
		notification.LastError = ""
		logger.Info(fmt.Sprintf("Successfully retried notification %s (attempt %d)", notification.ID, notification.AttemptCount))
	} else {
		notification.Status = domain.NotificationStatusFailed
		notification.LastError = err.Error()
		logger.Error(fmt.Sprintf("Retry failed for notification %s (attempt %d): %v", notification.ID, notification.AttemptCount, err))
	}

	if err := uc.notificationRepo.Update(ctx, notification); err != nil {
		logger.Error(fmt.Sprintf("Failed to update notification after retry: %v", err))
		return fmt.Errorf("failed to update notification: %w", err)
	}

	// Update channel statistics
	if err := uc.channelRepo.UpdateStatistics(ctx, channel.ID, err == nil); err != nil {
		logger.Error(fmt.Sprintf("Failed to update channel statistics: %v", err))
	}

	return nil
}

// TestChannel sends a test notification to a Discord channel
func (uc *NotificationUseCase) TestChannel(ctx context.Context, channelID string) error {
	channel, err := uc.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	// Create test event
	testEvent := &domain.Event{
		ID:        uuid.New().String(),
		Type:      "test.notification",
		Source:    "notification-service",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"message":    "This is a test notification from ToxicToastGo",
			"channel_id": channelID,
			"test":       true,
		},
	}

	payloadBytes, err := json.Marshal(testEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal test event: %w", err)
	}

	// Create notification
	notification := &domain.Notification{
		ID:           uuid.New().String(),
		ChannelID:    channel.ID,
		EventID:      testEvent.ID,
		EventType:    testEvent.Type,
		EventPayload: string(payloadBytes),
		Status:       domain.NotificationStatusPending,
		AttemptCount: 0,
	}

	// Save to database
	if err := uc.notificationRepo.Create(ctx, notification); err != nil {
		return fmt.Errorf("failed to create test notification: %w", err)
	}

	// Send notification
	uc.sendNotification(ctx, notification, channel, testEvent)

	logger.Info(fmt.Sprintf("Sent test notification to Discord channel %s", channel.Name))
	return nil
}
