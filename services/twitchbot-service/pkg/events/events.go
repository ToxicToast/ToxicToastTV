package events

import (
	"time"

	"github.com/toxictoast/toxictoastgo/shared/kafka"
)

// EventType represents the type of event
type EventType string

const (
	// Stream events
	StreamStarted EventType = "twitchbot.stream.started"
	StreamEnded   EventType = "twitchbot.stream.ended"
	StreamUpdated EventType = "twitchbot.stream.updated"

	// Message events
	MessageReceived EventType = "twitchbot.message.received"
	MessageDeleted  EventType = "twitchbot.message.deleted"
	MessageTimeout  EventType = "twitchbot.message.timeout"

	// Viewer events
	ViewerJoined      EventType = "twitchbot.viewer.joined"
	ViewerLeft        EventType = "twitchbot.viewer.left"
	ViewerBanned      EventType = "twitchbot.viewer.banned"
	ViewerUnbanned    EventType = "twitchbot.viewer.unbanned"
	ViewerModAdded    EventType = "twitchbot.viewer.mod.added"
	ViewerModRemoved  EventType = "twitchbot.viewer.mod.removed"
	ViewerVIPAdded    EventType = "twitchbot.viewer.vip.added"
	ViewerVIPRemoved  EventType = "twitchbot.viewer.vip.removed"

	// Clip events
	ClipCreated EventType = "twitchbot.clip.created"
	ClipUpdated EventType = "twitchbot.clip.updated"
	ClipDeleted EventType = "twitchbot.clip.deleted"

	// Command events
	CommandCreated  EventType = "twitchbot.command.created"
	CommandUpdated  EventType = "twitchbot.command.updated"
	CommandDeleted  EventType = "twitchbot.command.deleted"
	CommandExecuted EventType = "twitchbot.command.executed"
)

// BaseEvent contains common event fields
type BaseEvent struct {
	Type      EventType `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

// StreamEvent represents a stream-related event
type StreamEvent struct {
	BaseEvent
	StreamID       string `json:"stream_id"`
	Title          string `json:"title"`
	GameName       string `json:"game_name"`
	GameID         string `json:"game_id"`
	ViewerCount    int    `json:"viewer_count,omitempty"`
	IsActive       bool   `json:"is_active"`
	StartedAt      string `json:"started_at,omitempty"`
	EndedAt        string `json:"ended_at,omitempty"`
	PeakViewers    int    `json:"peak_viewers,omitempty"`
	AverageViewers int    `json:"average_viewers,omitempty"`
	TotalMessages  int    `json:"total_messages,omitempty"`
}

// MessageEvent represents a message-related event
type MessageEvent struct {
	BaseEvent
	MessageID     string `json:"message_id"`
	StreamID      string `json:"stream_id"`
	UserID        string `json:"user_id"`
	Username      string `json:"username"`
	DisplayName   string `json:"display_name"`
	Message       string `json:"message"`
	IsModerator   bool   `json:"is_moderator"`
	IsSubscriber  bool   `json:"is_subscriber"`
	IsVIP         bool   `json:"is_vip"`
	IsBroadcaster bool   `json:"is_broadcaster"`
}

// ViewerEvent represents a viewer-related event
type ViewerEvent struct {
	BaseEvent
	ViewerID            string `json:"viewer_id"`
	TwitchID            string `json:"twitch_id"`
	Username            string `json:"username"`
	DisplayName         string `json:"display_name"`
	TotalMessages       int    `json:"total_messages,omitempty"`
	TotalStreamsWatched int    `json:"total_streams_watched,omitempty"`
}

// ClipEvent represents a clip-related event
type ClipEvent struct {
	BaseEvent
	ClipID          string `json:"clip_id"`
	StreamID        string `json:"stream_id"`
	TwitchClipID    string `json:"twitch_clip_id"`
	Title           string `json:"title"`
	URL             string `json:"url"`
	CreatorName     string `json:"creator_name"`
	CreatorID       string `json:"creator_id"`
	ViewCount       int    `json:"view_count,omitempty"`
	DurationSeconds int    `json:"duration_seconds"`
}

// CommandEvent represents a command-related event
type CommandEvent struct {
	BaseEvent
	CommandID      string `json:"command_id"`
	CommandName    string `json:"command_name"`
	UserID         string `json:"user_id,omitempty"`
	Username       string `json:"username,omitempty"`
	Response       string `json:"response,omitempty"`
	Success        bool   `json:"success,omitempty"`
	IsActive       bool   `json:"is_active,omitempty"`
	ModeratorOnly  bool   `json:"moderator_only,omitempty"`
	SubscriberOnly bool   `json:"subscriber_only,omitempty"`
}

// EventPublisher handles publishing events to Kafka
type EventPublisher struct {
	producer *kafka.Producer
	topics   map[EventType]string
}

// NewEventPublisher creates a new event publisher
func NewEventPublisher(producer *kafka.Producer) *EventPublisher {
	return &EventPublisher{
		producer: producer,
		topics: map[EventType]string{
			// Stream events
			StreamStarted: "twitchbot.stream.started",
			StreamEnded:   "twitchbot.stream.ended",
			StreamUpdated: "twitchbot.stream.updated",

			// Message events
			MessageReceived: "twitchbot.message.received",
			MessageDeleted:  "twitchbot.message.deleted",
			MessageTimeout:  "twitchbot.message.timeout",

			// Viewer events
			ViewerJoined:     "twitchbot.viewer.joined",
			ViewerLeft:       "twitchbot.viewer.left",
			ViewerBanned:     "twitchbot.viewer.banned",
			ViewerUnbanned:   "twitchbot.viewer.unbanned",
			ViewerModAdded:   "twitchbot.viewer.mod.added",
			ViewerModRemoved: "twitchbot.viewer.mod.removed",
			ViewerVIPAdded:   "twitchbot.viewer.vip.added",
			ViewerVIPRemoved: "twitchbot.viewer.vip.removed",

			// Clip events
			ClipCreated: "twitchbot.clip.created",
			ClipUpdated: "twitchbot.clip.updated",
			ClipDeleted: "twitchbot.clip.deleted",

			// Command events
			CommandCreated:  "twitchbot.command.created",
			CommandUpdated:  "twitchbot.command.updated",
			CommandDeleted:  "twitchbot.command.deleted",
			CommandExecuted: "twitchbot.command.executed",
		},
	}
}

// PublishStreamEvent publishes a stream event
func (p *EventPublisher) PublishStreamEvent(event StreamEvent) error {
	if p.producer == nil {
		return nil // Silently skip if Kafka is not configured
	}

	event.Timestamp = time.Now()
	topic := p.topics[event.Type]
	return p.producer.PublishEvent(topic, event.StreamID, event)
}

// PublishMessageEvent publishes a message event
func (p *EventPublisher) PublishMessageEvent(event MessageEvent) error {
	if p.producer == nil {
		return nil
	}

	event.Timestamp = time.Now()
	topic := p.topics[event.Type]
	return p.producer.PublishEvent(topic, event.MessageID, event)
}

// PublishViewerEvent publishes a viewer event
func (p *EventPublisher) PublishViewerEvent(event ViewerEvent) error {
	if p.producer == nil {
		return nil
	}

	event.Timestamp = time.Now()
	topic := p.topics[event.Type]
	return p.producer.PublishEvent(topic, event.ViewerID, event)
}

// PublishClipEvent publishes a clip event
func (p *EventPublisher) PublishClipEvent(event ClipEvent) error {
	if p.producer == nil {
		return nil
	}

	event.Timestamp = time.Now()
	topic := p.topics[event.Type]
	return p.producer.PublishEvent(topic, event.ClipID, event)
}

// PublishCommandEvent publishes a command event
func (p *EventPublisher) PublishCommandEvent(event CommandEvent) error {
	if p.producer == nil {
		return nil
	}

	event.Timestamp = time.Now()
	topic := p.topics[event.Type]
	return p.producer.PublishEvent(topic, event.CommandID, event)
}
