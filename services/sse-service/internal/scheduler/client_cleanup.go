package scheduler

import (
	"log"
	"time"

	"toxictoast/services/sse-service/internal/broker"
)

type ClientCleanupScheduler struct {
	broker          *broker.Broker
	interval        time.Duration
	inactiveTimeout time.Duration
	enabled         bool
	stopChan        chan struct{}
}

func NewClientCleanupScheduler(
	broker *broker.Broker,
	interval time.Duration,
	inactiveTimeout time.Duration,
	enabled bool,
) *ClientCleanupScheduler {
	return &ClientCleanupScheduler{
		broker:          broker,
		interval:        interval,
		inactiveTimeout: inactiveTimeout,
		enabled:         enabled,
		stopChan:        make(chan struct{}),
	}
}

func (s *ClientCleanupScheduler) Start() {
	if !s.enabled {
		log.Println("Client cleanup scheduler is disabled")
		return
	}

	log.Printf("Client cleanup scheduler started (interval: %v, timeout: %v)", s.interval, s.inactiveTimeout)

	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		// Run immediately on start
		s.cleanupInactiveClients()

		for {
			select {
			case <-ticker.C:
				s.cleanupInactiveClients()
			case <-s.stopChan:
				log.Println("Client cleanup scheduler stopped")
				return
			}
		}
	}()
}

func (s *ClientCleanupScheduler) Stop() {
	if s.enabled {
		close(s.stopChan)
	}
}

func (s *ClientCleanupScheduler) cleanupInactiveClients() {
	log.Println("Checking for inactive clients...")

	stats := s.broker.GetStats()
	now := time.Now()
	cleanedCount := 0

	for _, client := range stats.ConnectedClients {
		inactiveDuration := now.Sub(client.LastEventAt)

		if inactiveDuration > s.inactiveTimeout {
			log.Printf("Disconnecting inactive client %s (inactive for %v)", client.ID, inactiveDuration.Round(time.Second))
			s.broker.UnregisterClient(client.ID)
			cleanedCount++
		}
	}

	if cleanedCount > 0 {
		log.Printf("Cleanup completed: disconnected %d inactive clients", cleanedCount)
	}
}
