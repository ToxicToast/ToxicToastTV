package scheduler

import (
	"context"
	"log"
	"time"

	"toxictoast/services/foodfolio-service/internal/repository/interfaces"
	"toxictoast/services/foodfolio-service/internal/usecase"
)

type ItemExpirationScheduler struct {
	itemDetailUseCase usecase.ItemDetailUseCase
	itemDetailRepo    interfaces.ItemDetailRepository
	interval          time.Duration
	enabled           bool
	stopChan          chan struct{}
}

func NewItemExpirationScheduler(
	itemDetailUseCase usecase.ItemDetailUseCase,
	itemDetailRepo interfaces.ItemDetailRepository,
	interval time.Duration,
	enabled bool,
) *ItemExpirationScheduler {
	return &ItemExpirationScheduler{
		itemDetailUseCase: itemDetailUseCase,
		itemDetailRepo:    itemDetailRepo,
		interval:          interval,
		enabled:           enabled,
		stopChan:          make(chan struct{}),
	}
}

func (s *ItemExpirationScheduler) Start() {
	if !s.enabled {
		log.Println("Item expiration scheduler is disabled")
		return
	}

	log.Printf("Item expiration scheduler started (interval: %v)", s.interval)

	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		// Run immediately on start
		s.checkExpirations()

		for {
			select {
			case <-ticker.C:
				s.checkExpirations()
			case <-s.stopChan:
				log.Println("Item expiration scheduler stopped")
				return
			}
		}
	}()
}

func (s *ItemExpirationScheduler) Stop() {
	if s.enabled {
		close(s.stopChan)
	}
}

func (s *ItemExpirationScheduler) checkExpirations() {
	ctx := context.Background()
	log.Println("Checking for expired and expiring items...")

	// Get all active item details
	offset := 0
	limit := 1000

	details, total, err := s.itemDetailRepo.List(ctx, offset, limit, nil, nil, nil, nil, nil, nil, false)
	if err != nil {
		log.Printf("Error listing item details for expiration check: %v", err)
		return
	}

	log.Printf("Found %d item details to check", total)

	expiredCount := 0
	expiringSoonCount := 0
	errorCount := 0

	for i := range details {
		detail := details[i]

		// Check if expired
		if detail.IsExpired() {
			err := s.itemDetailUseCase.NotifyItemExpired(ctx, detail)
			if err != nil {
				log.Printf("Error notifying expired item %s: %v", detail.ID, err)
				errorCount++
				continue
			}

			log.Printf("Notified expired item: %s (expired at: %v)", detail.ID, detail.ExpiryDate)
			expiredCount++
		} else if detail.IsExpiringSoon(7) { // Check if expiring within 7 days
			err := s.itemDetailUseCase.NotifyItemExpiringSoon(ctx, detail)
			if err != nil {
				log.Printf("Error notifying expiring soon item %s: %v", detail.ID, err)
				errorCount++
				continue
			}

			log.Printf("Notified expiring soon item: %s (expires at: %v)", detail.ID, detail.ExpiryDate)
			expiringSoonCount++
		}

		time.Sleep(10 * time.Millisecond)
	}

	if expiredCount > 0 || expiringSoonCount > 0 || errorCount > 0 {
		log.Printf("Expiration check completed: %d expired, %d expiring soon, %d errors",
			expiredCount, expiringSoonCount, errorCount)
	}
}
