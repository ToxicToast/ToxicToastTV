package events

import (
	"fmt"
	"time"

	"github.com/toxictoast/toxictoastgo/shared/kafka"
)

// EventPublisher handles publishing events to Kafka using shared event types
type EventPublisher struct {
	producer    *kafka.Producer
	channelID   string
	channelName string
}

// NewEventPublisher creates a new event publisher
func NewEventPublisher(producer *kafka.Producer) *EventPublisher {
	return &EventPublisher{
		producer: producer,
	}
}

// SetChannelInfo sets the channel context for events
func (p *EventPublisher) SetChannelInfo(channelID, channelName string) {
	p.channelID = channelID
	p.channelName = channelName
}

// PublishStreamStarted publishes a stream started event
func (p *EventPublisher) PublishStreamStarted(streamID, title, gameName string, viewerCount int) error {
	if p.producer == nil {
		return nil
	}

	event := kafka.TwitchbotStreamStartedEvent{
		StreamID:    streamID,
		ChannelID:   p.channelID,
		ChannelName: p.channelName,
		Title:       title,
		GameName:    gameName,
		ViewerCount: viewerCount,
		StartedAt:   time.Now(),
	}

	if err := p.producer.PublishTwitchbotStreamStarted("twitchbot.stream.started", event); err != nil {
		return fmt.Errorf("failed to publish stream started event: %w", err)
	}
	return nil
}

// PublishStreamEnded publishes a stream ended event
func (p *EventPublisher) PublishStreamEnded(streamID string, duration int) error {
	if p.producer == nil {
		return nil
	}

	event := kafka.TwitchbotStreamEndedEvent{
		StreamID:    streamID,
		ChannelID:   p.channelID,
		ChannelName: p.channelName,
		Duration:    duration,
		EndedAt:     time.Now(),
	}

	if err := p.producer.PublishTwitchbotStreamEnded("twitchbot.stream.ended", event); err != nil {
		return fmt.Errorf("failed to publish stream ended event: %w", err)
	}
	return nil
}

// PublishStreamUpdated publishes a stream updated event
func (p *EventPublisher) PublishStreamUpdated(streamID, title, gameName string, viewerCount int) error {
	if p.producer == nil {
		return nil
	}

	event := kafka.TwitchbotStreamUpdatedEvent{
		StreamID:    streamID,
		ChannelID:   p.channelID,
		ChannelName: p.channelName,
		Title:       title,
		GameName:    gameName,
		ViewerCount: viewerCount,
		UpdatedAt:   time.Now(),
	}

	if err := p.producer.PublishTwitchbotStreamUpdated("twitchbot.stream.updated", event); err != nil {
		return fmt.Errorf("failed to publish stream updated event: %w", err)
	}
	return nil
}

// PublishMessageReceived publishes a message received event
func (p *EventPublisher) PublishMessageReceived(messageID, userID, username, message string) error {
	if p.producer == nil {
		return nil
	}

	event := kafka.TwitchbotMessageReceivedEvent{
		MessageID:   messageID,
		ChannelID:   p.channelID,
		ChannelName: p.channelName,
		UserID:      userID,
		Username:    username,
		Message:     message,
		ReceivedAt:  time.Now(),
	}

	if err := p.producer.PublishTwitchbotMessageReceived("twitchbot.message.received", event); err != nil {
		return fmt.Errorf("failed to publish message received event: %w", err)
	}
	return nil
}

// PublishMessageDeleted publishes a message deleted event
func (p *EventPublisher) PublishMessageDeleted(messageID, deletedBy string) error {
	if p.producer == nil {
		return nil
	}

	event := kafka.TwitchbotMessageDeletedEvent{
		MessageID:   messageID,
		ChannelID:   p.channelID,
		ChannelName: p.channelName,
		DeletedBy:   deletedBy,
		DeletedAt:   time.Now(),
	}

	if err := p.producer.PublishTwitchbotMessageDeleted("twitchbot.message.deleted", event); err != nil {
		return fmt.Errorf("failed to publish message deleted event: %w", err)
	}
	return nil
}

// PublishMessageTimeout publishes a message timeout event
func (p *EventPublisher) PublishMessageTimeout(userID, username, reason string, duration int) error {
	if p.producer == nil {
		return nil
	}

	event := kafka.TwitchbotMessageTimeoutEvent{
		UserID:      userID,
		Username:    username,
		ChannelID:   p.channelID,
		ChannelName: p.channelName,
		Duration:    duration,
		Reason:      reason,
		TimeoutAt:   time.Now(),
	}

	if err := p.producer.PublishTwitchbotMessageTimeout("twitchbot.message.timeout", event); err != nil {
		return fmt.Errorf("failed to publish message timeout event: %w", err)
	}
	return nil
}

// PublishViewerJoined publishes a viewer joined event
func (p *EventPublisher) PublishViewerJoined(userID, username string) error {
	if p.producer == nil {
		return nil
	}

	event := kafka.TwitchbotViewerJoinedEvent{
		UserID:      userID,
		Username:    username,
		ChannelID:   p.channelID,
		ChannelName: p.channelName,
		JoinedAt:    time.Now(),
	}

	if err := p.producer.PublishTwitchbotViewerJoined("twitchbot.viewer.joined", event); err != nil {
		return fmt.Errorf("failed to publish viewer joined event: %w", err)
	}
	return nil
}

// PublishViewerLeft publishes a viewer left event
func (p *EventPublisher) PublishViewerLeft(userID, username string) error {
	if p.producer == nil {
		return nil
	}

	event := kafka.TwitchbotViewerLeftEvent{
		UserID:      userID,
		Username:    username,
		ChannelID:   p.channelID,
		ChannelName: p.channelName,
		LeftAt:      time.Now(),
	}

	if err := p.producer.PublishTwitchbotViewerLeft("twitchbot.viewer.left", event); err != nil {
		return fmt.Errorf("failed to publish viewer left event: %w", err)
	}
	return nil
}

// PublishViewerBanned publishes a viewer banned event
func (p *EventPublisher) PublishViewerBanned(userID, username, reason, bannedBy string) error {
	if p.producer == nil {
		return nil
	}

	event := kafka.TwitchbotViewerBannedEvent{
		UserID:      userID,
		Username:    username,
		ChannelID:   p.channelID,
		ChannelName: p.channelName,
		Reason:      reason,
		BannedBy:    bannedBy,
		BannedAt:    time.Now(),
	}

	if err := p.producer.PublishTwitchbotViewerBanned("twitchbot.viewer.banned", event); err != nil {
		return fmt.Errorf("failed to publish viewer banned event: %w", err)
	}
	return nil
}

// PublishViewerUnbanned publishes a viewer unbanned event
func (p *EventPublisher) PublishViewerUnbanned(userID, username, unbannedBy string) error {
	if p.producer == nil {
		return nil
	}

	event := kafka.TwitchbotViewerUnbannedEvent{
		UserID:      userID,
		Username:    username,
		ChannelID:   p.channelID,
		ChannelName: p.channelName,
		UnbannedBy:  unbannedBy,
		UnbannedAt:  time.Now(),
	}

	if err := p.producer.PublishTwitchbotViewerUnbanned("twitchbot.viewer.unbanned", event); err != nil {
		return fmt.Errorf("failed to publish viewer unbanned event: %w", err)
	}
	return nil
}

// PublishViewerModAdded publishes a viewer mod added event
func (p *EventPublisher) PublishViewerModAdded(userID, username, addedBy string) error {
	if p.producer == nil {
		return nil
	}

	event := kafka.TwitchbotViewerModAddedEvent{
		UserID:      userID,
		Username:    username,
		ChannelID:   p.channelID,
		ChannelName: p.channelName,
		AddedBy:     addedBy,
		AddedAt:     time.Now(),
	}

	if err := p.producer.PublishTwitchbotViewerModAdded("twitchbot.viewer.mod.added", event); err != nil {
		return fmt.Errorf("failed to publish viewer mod added event: %w", err)
	}
	return nil
}

// PublishViewerModRemoved publishes a viewer mod removed event
func (p *EventPublisher) PublishViewerModRemoved(userID, username, removedBy string) error {
	if p.producer == nil {
		return nil
	}

	event := kafka.TwitchbotViewerModRemovedEvent{
		UserID:      userID,
		Username:    username,
		ChannelID:   p.channelID,
		ChannelName: p.channelName,
		RemovedBy:   removedBy,
		RemovedAt:   time.Now(),
	}

	if err := p.producer.PublishTwitchbotViewerModRemoved("twitchbot.viewer.mod.removed", event); err != nil {
		return fmt.Errorf("failed to publish viewer mod removed event: %w", err)
	}
	return nil
}

// PublishViewerVipAdded publishes a viewer VIP added event
func (p *EventPublisher) PublishViewerVipAdded(userID, username, addedBy string) error {
	if p.producer == nil {
		return nil
	}

	event := kafka.TwitchbotViewerVipAddedEvent{
		UserID:      userID,
		Username:    username,
		ChannelID:   p.channelID,
		ChannelName: p.channelName,
		AddedBy:     addedBy,
		AddedAt:     time.Now(),
	}

	if err := p.producer.PublishTwitchbotViewerVipAdded("twitchbot.viewer.vip.added", event); err != nil {
		return fmt.Errorf("failed to publish viewer VIP added event: %w", err)
	}
	return nil
}

// PublishViewerVipRemoved publishes a viewer VIP removed event
func (p *EventPublisher) PublishViewerVipRemoved(userID, username, removedBy string) error {
	if p.producer == nil {
		return nil
	}

	event := kafka.TwitchbotViewerVipRemovedEvent{
		UserID:      userID,
		Username:    username,
		ChannelID:   p.channelID,
		ChannelName: p.channelName,
		RemovedBy:   removedBy,
		RemovedAt:   time.Now(),
	}

	if err := p.producer.PublishTwitchbotViewerVipRemoved("twitchbot.viewer.vip.removed", event); err != nil {
		return fmt.Errorf("failed to publish viewer VIP removed event: %w", err)
	}
	return nil
}

// PublishClipCreated publishes a clip created event
func (p *EventPublisher) PublishClipCreated(clipID, title, url, createdBy string) error {
	if p.producer == nil {
		return nil
	}

	event := kafka.TwitchbotClipCreatedEvent{
		ClipID:      clipID,
		ChannelID:   p.channelID,
		ChannelName: p.channelName,
		Title:       title,
		URL:         url,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
	}

	if err := p.producer.PublishTwitchbotClipCreated("twitchbot.clip.created", event); err != nil {
		return fmt.Errorf("failed to publish clip created event: %w", err)
	}
	return nil
}

// PublishClipUpdated publishes a clip updated event
func (p *EventPublisher) PublishClipUpdated(clipID, title string) error {
	if p.producer == nil {
		return nil
	}

	event := kafka.TwitchbotClipUpdatedEvent{
		ClipID:    clipID,
		Title:     title,
		UpdatedAt: time.Now(),
	}

	if err := p.producer.PublishTwitchbotClipUpdated("twitchbot.clip.updated", event); err != nil {
		return fmt.Errorf("failed to publish clip updated event: %w", err)
	}
	return nil
}

// PublishClipDeleted publishes a clip deleted event
func (p *EventPublisher) PublishClipDeleted(clipID string) error {
	if p.producer == nil {
		return nil
	}

	event := kafka.TwitchbotClipDeletedEvent{
		ClipID:    clipID,
		DeletedAt: time.Now(),
	}

	if err := p.producer.PublishTwitchbotClipDeleted("twitchbot.clip.deleted", event); err != nil {
		return fmt.Errorf("failed to publish clip deleted event: %w", err)
	}
	return nil
}

// PublishCommandCreated publishes a command created event
func (p *EventPublisher) PublishCommandCreated(commandID, name, response string) error {
	if p.producer == nil {
		return nil
	}

	event := kafka.TwitchbotCommandCreatedEvent{
		CommandID:   commandID,
		Name:        name,
		ChannelID:   p.channelID,
		ChannelName: p.channelName,
		Response:    response,
		CreatedAt:   time.Now(),
	}

	if err := p.producer.PublishTwitchbotCommandCreated("twitchbot.command.created", event); err != nil {
		return fmt.Errorf("failed to publish command created event: %w", err)
	}
	return nil
}

// PublishCommandUpdated publishes a command updated event
func (p *EventPublisher) PublishCommandUpdated(commandID, name, response string) error {
	if p.producer == nil {
		return nil
	}

	event := kafka.TwitchbotCommandUpdatedEvent{
		CommandID: commandID,
		Name:      name,
		Response:  response,
		UpdatedAt: time.Now(),
	}

	if err := p.producer.PublishTwitchbotCommandUpdated("twitchbot.command.updated", event); err != nil {
		return fmt.Errorf("failed to publish command updated event: %w", err)
	}
	return nil
}

// PublishCommandDeleted publishes a command deleted event
func (p *EventPublisher) PublishCommandDeleted(commandID, name string) error {
	if p.producer == nil {
		return nil
	}

	event := kafka.TwitchbotCommandDeletedEvent{
		CommandID: commandID,
		Name:      name,
		DeletedAt: time.Now(),
	}

	if err := p.producer.PublishTwitchbotCommandDeleted("twitchbot.command.deleted", event); err != nil {
		return fmt.Errorf("failed to publish command deleted event: %w", err)
	}
	return nil
}

// PublishCommandExecuted publishes a command executed event
func (p *EventPublisher) PublishCommandExecuted(commandID, name, userID, username string) error {
	if p.producer == nil {
		return nil
	}

	event := kafka.TwitchbotCommandExecutedEvent{
		CommandID:   commandID,
		Name:        name,
		ChannelID:   p.channelID,
		ChannelName: p.channelName,
		UserID:      userID,
		Username:    username,
		ExecutedAt:  time.Now(),
	}

	if err := p.producer.PublishTwitchbotCommandExecuted("twitchbot.command.executed", event); err != nil {
		return fmt.Errorf("failed to publish command executed event: %w", err)
	}
	return nil
}
