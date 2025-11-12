package domain

import (
	"encoding/json"
	"time"
)

// Event represents a generic event from any service
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`      // e.g., "blog.post.created", "twitchbot.message.received"
	Source    string                 `json:"source"`    // Service name: "blog-service", "twitchbot-service", etc.
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// EventType represents an event type category
type EventType string

const (
	// Blog Service Events
	EventBlogPostCreated      EventType = "blog.post.created"
	EventBlogPostUpdated      EventType = "blog.post.updated"
	EventBlogPostDeleted      EventType = "blog.post.deleted"
	EventBlogCommentCreated   EventType = "blog.comment.created"
	EventBlogCategoryCreated  EventType = "blog.category.created"
	EventBlogTagCreated       EventType = "blog.tag.created"
	EventBlogMediaUploaded    EventType = "blog.media.uploaded"

	// Twitchbot Service Events
	EventTwitchStreamStarted  EventType = "twitchbot.stream.started"
	EventTwitchStreamEnded    EventType = "twitchbot.stream.ended"
	EventTwitchMessageCreated EventType = "twitchbot.message.created"
	EventTwitchViewerJoined   EventType = "twitchbot.viewer.joined"
	EventTwitchClipCreated    EventType = "twitchbot.clip.created"
	EventTwitchCommandExecuted EventType = "twitchbot.command.executed"

	// Link Service Events
	EventLinkCreated          EventType = "link.created"
	EventLinkDeleted          EventType = "link.deleted"
	EventLinkClicked          EventType = "link.clicked"
)

// ParseEvent parses a Kafka message into an Event
func ParseEvent(value []byte, topic string) (*Event, error) {
	// First, try to parse as raw data (map)
	var rawData map[string]interface{}
	if err := json.Unmarshal(value, &rawData); err != nil {
		return nil, err
	}

	// Create SSE Event with topic as type
	event := &Event{
		Type:      topic, // Use topic as event type (e.g., "blog.post.created")
		Source:    inferSourceFromTopic(topic),
		Timestamp: time.Now(),
		Data:      rawData,
	}

	// Try to extract ID from common ID fields
	event.ID = extractID(rawData)

	// Try to extract timestamp from event data
	if timestamp := extractTimestamp(rawData); !timestamp.IsZero() {
		event.Timestamp = timestamp
	}

	return event, nil
}

// extractID tries to find an ID from common field names
func extractID(data map[string]interface{}) string {
	// Try common ID field names
	idFields := []string{
		"post_id", "category_id", "tag_id", "comment_id", "media_id", "author_id",
		"link_id", "click_id",
		"stream_id", "message_id", "viewer_id", "clip_id", "command_id",
		"item_id", "variant_id", "detail_id", "location_id", "warehouse_id",
		"receipt_id", "shoppinglist_id",
		"webhook_id", "delivery_id",
		"notification_id", "channel_id",
		"id", "ID",
	}

	for _, field := range idFields {
		if val, ok := data[field]; ok {
			if strVal, ok := val.(string); ok && strVal != "" {
				return strVal
			}
		}
	}

	return ""
}

// extractTimestamp tries to find a timestamp from common field names
func extractTimestamp(data map[string]interface{}) time.Time {
	timestampFields := []string{"created_at", "updated_at", "timestamp", "occurred_at", "event_time"}

	for _, field := range timestampFields {
		if val, ok := data[field]; ok {
			// Try as string (RFC3339 format)
			if strVal, ok := val.(string); ok {
				if t, err := time.Parse(time.RFC3339, strVal); err == nil {
					return t
				}
				// Try alternative formats
				if t, err := time.Parse(time.RFC3339Nano, strVal); err == nil {
					return t
				}
			}
			// Try as float64 (Unix timestamp)
			if floatVal, ok := val.(float64); ok {
				return time.Unix(int64(floatVal), 0)
			}
		}
	}

	return time.Time{}
}

// ToJSON converts event to JSON bytes
func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// MatchesFilter checks if event matches a subscription filter
func (e *Event) MatchesFilter(filter SubscriptionFilter) bool {
	// If no filters, match all
	if len(filter.EventTypes) == 0 && len(filter.Sources) == 0 {
		return true
	}

	// Check event type filter
	if len(filter.EventTypes) > 0 {
		matched := false
		for _, et := range filter.EventTypes {
			if string(et) == e.Type || matchesPattern(e.Type, string(et)) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check source filter
	if len(filter.Sources) > 0 {
		matched := false
		for _, source := range filter.Sources {
			if source == e.Source {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

// matchesPattern checks if event type matches pattern (e.g., "blog.*" matches "blog.post.created")
func matchesPattern(eventType, pattern string) bool {
	if pattern == "*" {
		return true
	}

	// Simple wildcard matching (e.g., "blog.*")
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(eventType) >= len(prefix) && eventType[:len(prefix)] == prefix
	}

	return eventType == pattern
}

// inferSourceFromTopic infers service name from Kafka topic
func inferSourceFromTopic(topic string) string {
	// Topics are usually like: "blog.posts", "twitchbot.messages", etc.
	if len(topic) == 0 {
		return "unknown"
	}

	// Extract first part before dot
	for i, c := range topic {
		if c == '.' {
			return topic[:i] + "-service"
		}
	}

	return topic + "-service"
}
