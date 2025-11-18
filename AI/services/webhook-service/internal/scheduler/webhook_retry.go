package scheduler

import (
	"context"
	"log"
	"math"
	"time"

	"toxictoast/services/webhook-service/internal/domain"
	"toxictoast/services/webhook-service/internal/repository/interfaces"
)

type WebhookRetryScheduler struct {
	deliveryRepo interfaces.DeliveryRepository
	interval     time.Duration
	maxRetries   int
	enabled      bool
	stopChan     chan struct{}
}

func NewWebhookRetryScheduler(
	deliveryRepo interfaces.DeliveryRepository,
	interval time.Duration,
	maxRetries int,
	enabled bool,
) *WebhookRetryScheduler {
	return &WebhookRetryScheduler{
		deliveryRepo: deliveryRepo,
		interval:     interval,
		maxRetries:   maxRetries,
		enabled:      enabled,
		stopChan:     make(chan struct{}),
	}
}

func (s *WebhookRetryScheduler) Start() {
	if !s.enabled {
		log.Println("Webhook retry scheduler is disabled")
		return
	}

	log.Printf("Webhook retry scheduler started (interval: %v, max retries: %d)", s.interval, s.maxRetries)

	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		// Run immediately on start
		s.scheduleFailedDeliveries()

		for {
			select {
			case <-ticker.C:
				s.scheduleFailedDeliveries()
			case <-s.stopChan:
				log.Println("Webhook retry scheduler stopped")
				return
			}
		}
	}()
}

func (s *WebhookRetryScheduler) Stop() {
	if s.enabled {
		close(s.stopChan)
	}
}

func (s *WebhookRetryScheduler) scheduleFailedDeliveries() {
	ctx := context.Background()
	log.Println("Checking for failed deliveries to schedule for retry...")

	// Get all failed deliveries
	failedDeliveries, _, err := s.deliveryRepo.List(ctx, "", domain.DeliveryStatusFailed, 100, 0)
	if err != nil {
		log.Printf("Error listing failed deliveries: %v", err)
		return
	}

	if len(failedDeliveries) == 0 {
		return
	}

	scheduledCount := 0
	skippedCount := 0

	for _, delivery := range failedDeliveries {
		// Skip if already at max retries
		if delivery.AttemptCount >= s.maxRetries {
			skippedCount++
			continue
		}

		// Calculate next retry time with exponential backoff
		// First retry after 1 min, then 2 min, 4 min, 8 min, etc.
		backoffMinutes := math.Pow(2, float64(delivery.AttemptCount))
		nextRetry := time.Now().Add(time.Duration(backoffMinutes) * time.Minute)

		// Update delivery status to retrying
		delivery.Status = domain.DeliveryStatusRetrying
		delivery.NextRetryAt = &nextRetry
		delivery.UpdatedAt = time.Now()

		if err := s.deliveryRepo.Update(ctx, delivery); err != nil {
			log.Printf("Error scheduling delivery %s for retry: %v", delivery.ID, err)
			continue
		}

		log.Printf("Scheduled delivery %s for retry at %v (attempt %d/%d)",
			delivery.ID, nextRetry.Format(time.RFC3339), delivery.AttemptCount+1, s.maxRetries)
		scheduledCount++
	}

	if scheduledCount > 0 || skippedCount > 0 {
		log.Printf("Scheduling completed: %d deliveries scheduled, %d skipped (max retries reached)", scheduledCount, skippedCount)
	}
}
