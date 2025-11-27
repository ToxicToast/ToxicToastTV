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

	"toxictoast/services/notification-service/internal/discord"
	"toxictoast/services/notification-service/internal/domain"
	"toxictoast/services/notification-service/internal/repository/interfaces"
)

// ============================================================================
// Commands
// ============================================================================

type ProcessEventCommand struct {
	cqrs.BaseCommand
	Event *domain.Event `json:"event"`
}

func (c *ProcessEventCommand) CommandName() string {
	return "process_event"
}

func (c *ProcessEventCommand) Validate() error {
	if c.Event == nil {
		return errors.New("event is required")
	}
	if c.Event.Type == "" {
		return errors.New("event type is required")
	}
	return nil
}

type DeleteNotificationCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteNotificationCommand) CommandName() string {
	return "delete_notification"
}

func (c *DeleteNotificationCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("notification ID is required")
	}
	return nil
}

type CleanupOldNotificationsCommand struct {
	cqrs.BaseCommand
	OlderThanDays int `json:"older_than_days"`
}

func (c *CleanupOldNotificationsCommand) CommandName() string {
	return "cleanup_old_notifications"
}

func (c *CleanupOldNotificationsCommand) Validate() error {
	if c.OlderThanDays <= 0 {
		return errors.New("olderThanDays must be positive")
	}
	return nil
}

type RetryNotificationCommand struct {
	cqrs.BaseCommand
	NotificationID string `json:"notification_id"`
}

func (c *RetryNotificationCommand) CommandName() string {
	return "retry_notification"
}

func (c *RetryNotificationCommand) Validate() error {
	if c.NotificationID == "" {
		return errors.New("notification ID is required")
	}
	return nil
}

type TestChannelCommand struct {
	cqrs.BaseCommand
	ChannelID string `json:"channel_id"`
}

func (c *TestChannelCommand) CommandName() string {
	return "test_channel"
}

func (c *TestChannelCommand) Validate() error {
	if c.ChannelID == "" {
		return errors.New("channel ID is required")
	}
	return nil
}

// ============================================================================
// Command Handlers
// ============================================================================

type ProcessEventHandler struct {
	notificationRepo interfaces.NotificationRepository
	channelRepo      interfaces.DiscordChannelRepository
	discordClient    *discord.Client
}

func NewProcessEventHandler(
	notificationRepo interfaces.NotificationRepository,
	channelRepo interfaces.DiscordChannelRepository,
	discordClient *discord.Client,
) *ProcessEventHandler {
	return &ProcessEventHandler{
		notificationRepo: notificationRepo,
		channelRepo:      channelRepo,
		discordClient:    discordClient,
	}
}

func (h *ProcessEventHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	processCmd := cmd.(*ProcessEventCommand)
	event := processCmd.Event

	logger.Info(fmt.Sprintf("Processing event %s (type: %s)", event.ID, event.Type))

	// Find all active channels that match this event type
	channels, err := h.channelRepo.GetActiveChannelsForEvent(ctx, event.Type)
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
		if err := h.notificationRepo.Create(ctx, notification); err != nil {
			logger.Error(fmt.Sprintf("Failed to create notification for channel %s: %v", channel.ID, err))
			continue
		}

		// Send notification
		h.sendNotification(ctx, notification, channel, event)
	}

	return nil
}

func (h *ProcessEventHandler) sendNotification(ctx context.Context, notification *domain.Notification, channel *domain.DiscordChannel, event *domain.Event) {
	startTime := time.Now()

	// Build Discord embed
	color := channel.Color
	if color == 0 {
		color = discord.GetEmbedColor(event.Type)
	}
	embed := discord.BuildEmbedFromEvent(event, color)

	// Send to Discord
	messageID, responseStatus, responseBody, err := h.discordClient.SendNotification(ctx, channel.WebhookURL, embed)

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

	if err := h.notificationRepo.CreateAttempt(ctx, attempt); err != nil {
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

	if err := h.notificationRepo.Update(ctx, notification); err != nil {
		logger.Error(fmt.Sprintf("Failed to update notification: %v", err))
	}

	// Update channel statistics
	if err := h.channelRepo.UpdateStatistics(ctx, channel.ID, err == nil); err != nil {
		logger.Error(fmt.Sprintf("Failed to update channel statistics: %v", err))
	}
}

type DeleteNotificationHandler struct {
	notificationRepo interfaces.NotificationRepository
}

func NewDeleteNotificationHandler(notificationRepo interfaces.NotificationRepository) *DeleteNotificationHandler {
	return &DeleteNotificationHandler{
		notificationRepo: notificationRepo,
	}
}

func (h *DeleteNotificationHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteNotificationCommand)

	if err := h.notificationRepo.Delete(ctx, deleteCmd.AggregateID); err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	logger.Info(fmt.Sprintf("Deleted notification %s", deleteCmd.AggregateID))
	return nil
}

type CleanupOldNotificationsHandler struct {
	notificationRepo interfaces.NotificationRepository
}

func NewCleanupOldNotificationsHandler(notificationRepo interfaces.NotificationRepository) *CleanupOldNotificationsHandler {
	return &CleanupOldNotificationsHandler{
		notificationRepo: notificationRepo,
	}
}

func (h *CleanupOldNotificationsHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	cleanupCmd := cmd.(*CleanupOldNotificationsCommand)

	duration := time.Duration(cleanupCmd.OlderThanDays) * 24 * time.Hour

	if err := h.notificationRepo.CleanupOldNotifications(ctx, duration); err != nil {
		return fmt.Errorf("failed to cleanup old notifications: %w", err)
	}

	logger.Info(fmt.Sprintf("Cleaned up notifications older than %d days", cleanupCmd.OlderThanDays))
	return nil
}

type RetryNotificationHandler struct {
	notificationRepo interfaces.NotificationRepository
	channelRepo      interfaces.DiscordChannelRepository
	discordClient    *discord.Client
}

func NewRetryNotificationHandler(
	notificationRepo interfaces.NotificationRepository,
	channelRepo interfaces.DiscordChannelRepository,
	discordClient *discord.Client,
) *RetryNotificationHandler {
	return &RetryNotificationHandler{
		notificationRepo: notificationRepo,
		channelRepo:      channelRepo,
		discordClient:    discordClient,
	}
}

func (h *RetryNotificationHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	retryCmd := cmd.(*RetryNotificationCommand)

	// Get the notification
	notification, err := h.notificationRepo.GetByID(ctx, retryCmd.NotificationID)
	if err != nil {
		return fmt.Errorf("failed to get notification: %w", err)
	}

	// Get the Discord channel
	channel, err := h.channelRepo.GetByID(ctx, notification.ChannelID)
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
	messageID, responseStatus, responseBody, err := h.discordClient.SendNotification(ctx, channel.WebhookURL, embed)

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

	if err := h.notificationRepo.CreateAttempt(ctx, attempt); err != nil {
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

	if err := h.notificationRepo.Update(ctx, notification); err != nil {
		logger.Error(fmt.Sprintf("Failed to update notification after retry: %v", err))
		return fmt.Errorf("failed to update notification: %w", err)
	}

	// Update channel statistics
	if err := h.channelRepo.UpdateStatistics(ctx, channel.ID, err == nil); err != nil {
		logger.Error(fmt.Sprintf("Failed to update channel statistics: %v", err))
	}

	return nil
}

type TestChannelHandler struct {
	notificationRepo interfaces.NotificationRepository
	channelRepo      interfaces.DiscordChannelRepository
	discordClient    *discord.Client
}

func NewTestChannelHandler(
	notificationRepo interfaces.NotificationRepository,
	channelRepo interfaces.DiscordChannelRepository,
	discordClient *discord.Client,
) *TestChannelHandler {
	return &TestChannelHandler{
		notificationRepo: notificationRepo,
		channelRepo:      channelRepo,
		discordClient:    discordClient,
	}
}

func (h *TestChannelHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	testCmd := cmd.(*TestChannelCommand)

	channel, err := h.channelRepo.GetByID(ctx, testCmd.ChannelID)
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
			"channel_id": testCmd.ChannelID,
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
	if err := h.notificationRepo.Create(ctx, notification); err != nil {
		return fmt.Errorf("failed to create test notification: %w", err)
	}

	// Send notification
	h.sendNotification(ctx, notification, channel, testEvent)

	logger.Info(fmt.Sprintf("Sent test notification to Discord channel %s", channel.Name))
	return nil
}

func (h *TestChannelHandler) sendNotification(ctx context.Context, notification *domain.Notification, channel *domain.DiscordChannel, event *domain.Event) {
	startTime := time.Now()

	// Build Discord embed
	color := channel.Color
	if color == 0 {
		color = discord.GetEmbedColor(event.Type)
	}
	embed := discord.BuildEmbedFromEvent(event, color)

	// Send to Discord
	messageID, responseStatus, responseBody, err := h.discordClient.SendNotification(ctx, channel.WebhookURL, embed)

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

	if err := h.notificationRepo.CreateAttempt(ctx, attempt); err != nil {
		logger.Error(fmt.Sprintf("Failed to record notification attempt: %v", err))
	}

	// Update notification status
	notification.AttemptCount = 1
	if err == nil {
		notification.Status = domain.NotificationStatusSuccess
		notification.DiscordMessageID = messageID
		now := time.Now()
		notification.SentAt = &now
		logger.Info(fmt.Sprintf("Successfully sent test notification %s to channel %s", notification.ID, channel.Name))
	} else {
		notification.Status = domain.NotificationStatusFailed
		notification.LastError = err.Error()
		logger.Error(fmt.Sprintf("Failed to send test notification %s: %v", notification.ID, err))
	}

	if err := h.notificationRepo.Update(ctx, notification); err != nil {
		logger.Error(fmt.Sprintf("Failed to update notification: %v", err))
	}

	// Update channel statistics
	if err := h.channelRepo.UpdateStatistics(ctx, channel.ID, err == nil); err != nil {
		logger.Error(fmt.Sprintf("Failed to update channel statistics: %v", err))
	}
}
