package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"

	"toxictoast/services/link-service/internal/command"
	"toxictoast/services/link-service/internal/repository"
)

type LinkExpirationScheduler struct {
	commandBus *cqrs.CommandBus
	linkRepo   repository.LinkRepository
	interval   time.Duration
	enabled    bool
	stopChan   chan struct{}
}

func NewLinkExpirationScheduler(
	commandBus *cqrs.CommandBus,
	linkRepo repository.LinkRepository,
	interval time.Duration,
	enabled bool,
) *LinkExpirationScheduler {
	return &LinkExpirationScheduler{
		commandBus: commandBus,
		linkRepo:   linkRepo,
		interval:   interval,
		enabled:    enabled,
		stopChan:   make(chan struct{}),
	}
}

func (s *LinkExpirationScheduler) Start() {
	if !s.enabled {
		log.Println("Link expiration scheduler is disabled")
		return
	}

	log.Printf("Link expiration scheduler started (interval: %v)", s.interval)

	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		// Run immediately on start
		s.checkExpiredLinks()

		for {
			select {
			case <-ticker.C:
				s.checkExpiredLinks()
			case <-s.stopChan:
				log.Println("Link expiration scheduler stopped")
				return
			}
		}
	}()
}

func (s *LinkExpirationScheduler) Stop() {
	if s.enabled {
		close(s.stopChan)
	}
}

func (s *LinkExpirationScheduler) checkExpiredLinks() {
	ctx := context.Background()
	log.Println("Checking for expired links...")

	// Get all active links (no pagination limit for background job)
	filters := repository.LinkFilters{
		IsActive: boolPtr(true),
		Page:     1,
		PageSize: 1000, // Process in batches of 1000
	}

	links, total, err := s.linkRepo.List(ctx, filters)
	if err != nil {
		log.Printf("Error listing links for expiration check: %v", err)
		return
	}

	log.Printf("Found %d active links to check", total)

	expiredCount := 0
	errorCount := 0

	for i := range links {
		// Check if link is expired
		if links[i].IsExpired() {
			// Create command to deactivate the link
			cmd := &command.DeactivateExpiredLinkCommand{
				BaseCommand: cqrs.BaseCommand{},
				Link:        &links[i],
			}

			// Dispatch command
			err := s.commandBus.Dispatch(ctx, cmd)
			if err != nil {
				log.Printf("Error deactivating expired link %s: %v", links[i].ShortCode, err)
				errorCount++
				continue
			}

			log.Printf("Deactivated expired link: %s (expired at: %v)", links[i].ShortCode, links[i].ExpiresAt)
			expiredCount++

			// Small delay to avoid overwhelming the system
			time.Sleep(10 * time.Millisecond)
		}
	}

	if expiredCount > 0 || errorCount > 0 {
		log.Printf("Link expiration check completed: %d expired, %d errors", expiredCount, errorCount)
	}
}

func boolPtr(b bool) *bool {
	return &b
}
