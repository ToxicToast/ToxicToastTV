package broker

import (
	"log"
	"sync"
	"time"

	"toxictoast/services/sse-service/internal/domain"
)

// Broker manages all SSE client connections and distributes events
type Broker struct {
	clients         map[string]*domain.Client
	clientsMu       sync.RWMutex
	newClients      chan *domain.Client
	closingClients  chan string
	events          chan *domain.Event
	maxClients      int
	heartbeatTicker *time.Ticker
	stopChan        chan struct{}
	history         *EventHistory
}

// NewBroker creates a new SSE broker
func NewBroker(maxClients int, heartbeatSeconds int, historySize int) *Broker {
	b := &Broker{
		clients:        make(map[string]*domain.Client),
		newClients:     make(chan *domain.Client),
		closingClients: make(chan string),
		events:         make(chan *domain.Event, 1000),
		maxClients:     maxClients,
		stopChan:       make(chan struct{}),
		history:        NewEventHistory(historySize),
	}

	if heartbeatSeconds > 0 {
		b.heartbeatTicker = time.NewTicker(time.Duration(heartbeatSeconds) * time.Second)
	}

	return b
}

// Start starts the broker's event loop
func (b *Broker) Start() {
	log.Println("🚀 SSE Broker started")

	go func() {
		for {
			select {
			case client := <-b.newClients:
				b.clientsMu.Lock()
				if len(b.clients) >= b.maxClients {
					log.Printf("⚠️  Max clients reached (%d), rejecting new connection", b.maxClients)
					client.Close()
					b.clientsMu.Unlock()
					continue
				}
				b.clients[client.ID] = client
				clientCount := len(b.clients)
				b.clientsMu.Unlock()
				log.Printf("✅ Client connected: %s (Total: %d)", client.ID, clientCount)

			case clientID := <-b.closingClients:
				b.clientsMu.Lock()
				if client, ok := b.clients[clientID]; ok {
					client.Close()
					delete(b.clients, clientID)
					clientCount := len(b.clients)
					b.clientsMu.Unlock()
					log.Printf("👋 Client disconnected: %s (Total: %d)", clientID, clientCount)
				} else {
					b.clientsMu.Unlock()
				}

			case event := <-b.events:
				// Broadcast event to all matching clients
				b.broadcastEvent(event)

			case <-b.heartbeatTicker.C:
				// Send heartbeat to all clients
				b.sendHeartbeat()

			case <-b.stopChan:
				log.Println("🛑 SSE Broker stopped")
				return
			}
		}
	}()
}

// Stop stops the broker
func (b *Broker) Stop() {
	// Signal stop first
	close(b.stopChan)

	// Give goroutine time to exit
	time.Sleep(100 * time.Millisecond)

	// Stop heartbeat ticker
	if b.heartbeatTicker != nil {
		b.heartbeatTicker.Stop()
	}

	// Close all clients
	b.clientsMu.Lock()
	for _, client := range b.clients {
		client.Close()
	}
	b.clients = make(map[string]*domain.Client)
	b.clientsMu.Unlock()

	// Close channels after goroutine has exited
	close(b.newClients)
	close(b.closingClients)
	close(b.events)

	log.Println("✅ SSE Broker channels closed")
}

// RegisterClient registers a new SSE client
func (b *Broker) RegisterClient(client *domain.Client) {
	b.newClients <- client
}

// UnregisterClient unregisters an SSE client
func (b *Broker) UnregisterClient(clientID string) {
	b.closingClients <- clientID
}

// PublishEvent publishes an event to all subscribed clients
func (b *Broker) PublishEvent(event *domain.Event) {
	select {
	case b.events <- event:
		// Event queued successfully
	default:
		log.Printf("⚠️  Event buffer full, dropping event: %s", event.Type)
	}
}

// broadcastEvent sends an event to all matching clients
func (b *Broker) broadcastEvent(event *domain.Event) {
	// Add to history first
	b.history.Add(event)

	b.clientsMu.RLock()
	defer b.clientsMu.RUnlock()

	sentCount := 0
	for _, client := range b.clients {
		if client.SendEvent(event) {
			sentCount++
		}
	}

	if sentCount > 0 {
		log.Printf("📤 Event broadcasted: %s (type: %s, sent to %d clients)", event.ID, event.Type, sentCount)
	}
}

// sendHeartbeat sends a heartbeat event to all clients
func (b *Broker) sendHeartbeat() {
	heartbeat := &domain.Event{
		ID:        "heartbeat",
		Type:      "heartbeat",
		Source:    "sse-service",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"message": "ping"},
	}

	b.clientsMu.RLock()
	defer b.clientsMu.RUnlock()

	for _, client := range b.clients {
		select {
		case client.Channel <- heartbeat:
		default:
			// Skip if channel is full
		}
	}
}

// GetStats returns broker statistics
func (b *Broker) GetStats() BrokerStats {
	b.clientsMu.RLock()
	defer b.clientsMu.RUnlock()

	clients := make([]domain.ClientStats, 0, len(b.clients))
	for _, client := range b.clients {
		clients = append(clients, client.GetStats())
	}

	return BrokerStats{
		TotalClients:     len(b.clients),
		MaxClients:       b.maxClients,
		ConnectedClients: clients,
		HistorySize:      b.history.Size(),
	}
}

// GetClientCount returns the current number of connected clients
func (b *Broker) GetClientCount() int {
	b.clientsMu.RLock()
	defer b.clientsMu.RUnlock()
	return len(b.clients)
}

// GetHistory returns all events in history
func (b *Broker) GetHistory() []*domain.Event {
	return b.history.GetAll()
}

// GetHistorySince returns events since a specific event ID
func (b *Broker) GetHistorySince(lastEventID string) []*domain.Event {
	return b.history.GetSince(lastEventID)
}

// BrokerStats represents broker statistics
type BrokerStats struct {
	TotalClients     int                  `json:"total_clients"`
	MaxClients       int                  `json:"max_clients"`
	ConnectedClients []domain.ClientStats `json:"connected_clients"`
	HistorySize      int                  `json:"history_size"`
}
