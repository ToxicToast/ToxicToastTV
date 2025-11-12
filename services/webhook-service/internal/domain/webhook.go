package domain

import (
	"time"

)

// Webhook represents a registered webhook endpoint
type Webhook struct {
	ID          string         
	URL         string         
	Secret      string          // HMAC secret (hidden from JSON)
	EventTypes  string                  // Comma-separated (e.g., "blog.*,twitchbot.message.*")
	Description string         
	Active      bool           
	CreatedAt   time.Time      
	UpdatedAt   time.Time      
	DeletedAt   *time.Time 

	// Statistics
	TotalDeliveries    int       
	SuccessDeliveries  int       
	FailedDeliveries   int       
	LastDeliveryAt     time.Time `json:"last_delivery_at,omitempty"`
	LastSuccessAt      time.Time `json:"last_success_at,omitempty"`
	LastFailureAt      time.Time `json:"last_failure_at,omitempty"`
}


// MatchesEvent checks if the webhook should receive this event type
func (w *Webhook) MatchesEvent(eventType string) bool {
	if w.EventTypes == "" || w.EventTypes == "*" {
		return true
	}

	// Split event types
	types := splitEventTypes(w.EventTypes)
	for _, t := range types {
		if matchesPattern(eventType, t) {
			return true
		}
	}

	return false
}

// splitEventTypes splits comma-separated event types
func splitEventTypes(eventTypes string) []string {
	if eventTypes == "" {
		return []string{}
	}

	result := []string{}
	for _, t := range splitAndTrim(eventTypes, ",") {
		if t != "" {
			result = append(result, t)
		}
	}
	return result
}

// splitAndTrim splits string and trims whitespace
func splitAndTrim(s, sep string) []string {
	parts := []string{}
	for _, part := range splitString(s, sep) {
		trimmed := trimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

// Simple string split (no strings.Split import needed for simplicity)
func splitString(s, sep string) []string {
	result := []string{}
	current := ""

	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, current)
			current = ""
			i += len(sep) - 1
		} else {
			current += string(s[i])
		}
	}
	result = append(result, current)
	return result
}

// Simple trim (no strings.TrimSpace import needed)
func trimSpace(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}

// matchesPattern checks if event type matches pattern (e.g., "blog.*" matches "blog.post.created")
func matchesPattern(eventType, pattern string) bool {
	if pattern == "*" {
		return true
	}

	// Wildcard matching
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(eventType) >= len(prefix) && eventType[:len(prefix)] == prefix
	}

	return eventType == pattern
}
