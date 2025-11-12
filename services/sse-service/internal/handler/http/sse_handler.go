package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"toxictoast/services/sse-service/internal/broker"
	"toxictoast/services/sse-service/internal/domain"
)

// SSEHandler handles SSE connections
type SSEHandler struct {
	broker          *broker.Broker
	eventBufferSize int
	corsOrigins     []string
	corsHeaders     []string
}

// NewSSEHandler creates a new SSE handler
func NewSSEHandler(b *broker.Broker, eventBufferSize int, corsOrigins []string, corsHeaders []string) *SSEHandler {
	return &SSEHandler{
		broker:          b,
		eventBufferSize: eventBufferSize,
		corsOrigins:     corsOrigins,
		corsHeaders:     corsHeaders,
	}
}

// HandleSSE handles SSE connections
// Query parameters:
// - event_types: comma-separated list of event types (e.g., "blog.*,twitchbot.message.*")
// - sources: comma-separated list of sources (e.g., "blog-service,twitchbot-service")
// Headers:
// - Last-Event-ID: Resume from this event ID (SSE reconnection standard)
func (h *SSEHandler) HandleSSE(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	h.setCORSHeaders(w, r)

	// Handle OPTIONS preflight
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Flush support
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Parse subscription filter from query params
	filter := parseSubscriptionFilter(r)

	// Check for Last-Event-ID header (SSE reconnection standard)
	lastEventID := r.Header.Get("Last-Event-ID")

	// Create client
	clientID := uuid.New().String()
	client := domain.NewClient(
		clientID,
		filter,
		r.UserAgent(),
		r.RemoteAddr,
		h.eventBufferSize,
	)

	// Register client
	h.broker.RegisterClient(client)
	defer h.broker.UnregisterClient(clientID)

	log.Printf("📡 SSE client connected: %s (Filter: %+v, Last-Event-ID: %s)", clientID, filter, lastEventID)

	// Send initial connection event
	connectionEvent := &domain.Event{
		ID:        clientID,
		Type:      "connection.established",
		Source:    "sse-service",
		Data: map[string]interface{}{
			"client_id": clientID,
			"message":   "Connected to SSE stream",
		},
	}
	h.sendEvent(w, flusher, connectionEvent)

	// Send event history
	var historyEvents []*domain.Event
	if lastEventID != "" {
		// Client is reconnecting, send events since last event
		historyEvents = h.broker.GetHistorySince(lastEventID)
		if len(historyEvents) > 0 {
			log.Printf("📜 Sending %d missed events to client %s (since %s)", len(historyEvents), clientID, lastEventID)
		}
	} else {
		// New connection, send recent history
		historyEvents = h.broker.GetHistory()
		if len(historyEvents) > 0 {
			log.Printf("📜 Sending %d history events to client %s", len(historyEvents), clientID)
		}
	}

	// Send history events that match the filter
	for _, event := range historyEvents {
		if event.MatchesFilter(filter) {
			if err := h.sendEvent(w, flusher, event); err != nil {
				log.Printf("⚠️  Error sending history event to client %s: %v", clientID, err)
				return
			}
		}
	}

	// Stream live events to client
	for event := range client.Channel {
		if err := h.sendEvent(w, flusher, event); err != nil {
			log.Printf("⚠️  Error sending event to client %s: %v", clientID, err)
			return
		}
	}

	log.Printf("📡 SSE client disconnected: %s", clientID)
}

// sendEvent sends an event to the client
func (h *SSEHandler) sendEvent(w http.ResponseWriter, flusher http.Flusher, event *domain.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// SSE format:
	// id: <event-id>
	// event: <event-type>
	// data: <json-data>
	// (blank line)

	fmt.Fprintf(w, "id: %s\n", event.ID)
	fmt.Fprintf(w, "event: %s\n", event.Type)
	fmt.Fprintf(w, "data: %s\n\n", string(data))

	flusher.Flush()
	return nil
}

// parseSubscriptionFilter parses subscription filter from query params
func parseSubscriptionFilter(r *http.Request) domain.SubscriptionFilter {
	filter := domain.SubscriptionFilter{}

	// Parse event_types
	if eventTypesStr := r.URL.Query().Get("event_types"); eventTypesStr != "" {
		eventTypesList := strings.Split(eventTypesStr, ",")
		filter.EventTypes = make([]domain.EventType, 0, len(eventTypesList))
		for _, et := range eventTypesList {
			filter.EventTypes = append(filter.EventTypes, domain.EventType(strings.TrimSpace(et)))
		}
	}

	// Parse sources
	if sourcesStr := r.URL.Query().Get("sources"); sourcesStr != "" {
		filter.Sources = strings.Split(sourcesStr, ",")
		for i, source := range filter.Sources {
			filter.Sources[i] = strings.TrimSpace(source)
		}
	}

	return filter
}

// setCORSHeaders sets CORS headers based on configuration
func (h *SSEHandler) setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")

	// Check if origin is allowed
	allowed := false
	for _, allowedOrigin := range h.corsOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			allowed = true
			break
		}
	}

	if allowed {
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", h.corsOrigins[0])
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(h.corsHeaders, ", "))
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")
	}
}

// HandleHealth handles health check requests
func (h *SSEHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Healthy"))
}

// HandleStats handles stats requests
func (h *SSEHandler) HandleStats(w http.ResponseWriter, r *http.Request) {
	stats := h.broker.GetStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
