package domain

import (
	"encoding/json"
	"time"
)

// Event represents a generic event from Kafka
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// ParseEvent parses a Kafka message into an Event
func ParseEvent(value []byte, topic string) (*Event, error) {
	var event Event
	if err := json.Unmarshal(value, &event); err != nil {
		return nil, err
	}

	// If event doesn't have a source, infer from topic
	if event.Source == "" {
		event.Source = inferSourceFromTopic(topic)
	}

	// If event doesn't have a timestamp, use current time
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	return &event, nil
}

// ToJSON converts event to JSON bytes
func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// inferSourceFromTopic infers service name from Kafka topic
func inferSourceFromTopic(topic string) string {
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
