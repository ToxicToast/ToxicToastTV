package scheduler

import (
	"context"
	"log"
	"time"

	"toxictoast/services/webhook-service/internal/repository/interfaces"
)

type WebhookCleanupScheduler struct {
	deliveryRepo  interfaces.DeliveryRepository
	interval      time.Duration
	retentionDays int
	enabled       bool
	stopChan      chan struct{}
}

func NewWebhookCleanupScheduler(
	deliveryRepo interfaces.DeliveryRepository,
	interval time.Duration,
	retentionDays int,
	enabled bool,
) *WebhookCleanupScheduler {
	return &WebhookCleanupScheduler{
		deliveryRepo:  deliveryRepo,
		interval:      interval,
		retentionDays: retentionDays,
		enabled:       enabled,
		stopChan:      make(chan struct{}),
	}
}

func (s *WebhookCleanupScheduler) Start() {
	if !s.enabled {
		log.Println("Webhook cleanup scheduler is disabled")
		return
	}

	log.Printf("Webhook cleanup scheduler started (interval: %v, retention: %d days)", s.interval, s.retentionDays)

	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		// Run immediately on start
		s.cleanupOldDeliveries()

		for {
			select {
			case <-ticker.C:
				s.cleanupOldDeliveries()
			case <-s.stopChan:
				log.Println("Webhook cleanup scheduler stopped")
				return
			}
		}
	}()
}

func (s *WebhookCleanupScheduler) Stop() {
	if s.enabled {
		close(s.stopChan)
	}
}

func (s *WebhookCleanupScheduler) cleanupOldDeliveries() {
	ctx := context.Background()
	log.Println("Cleaning up old webhook deliveries...")

	// Calculate cutoff duration
	duration := time.Duration(s.retentionDays) * 24 * time.Hour

	// Delete old deliveries (both successful and failed)
	err := s.deliveryRepo.CleanupOldDeliveries(ctx, duration)
	if err != nil {
		log.Printf("Error cleaning up old deliveries: %v", err)
		return
	}

	log.Printf("Cleanup completed: removed deliveries older than %d days", s.retentionDays)
}
