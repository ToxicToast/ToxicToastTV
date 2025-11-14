package scheduler

import (
	"context"
	"log"
	"time"

	"toxictoast/services/twitchbot-service/internal/repository/interfaces"
)

type MessageCleanupScheduler struct {
	messageRepo   interfaces.MessageRepository
	interval      time.Duration
	retentionDays int
	enabled       bool
	stopChan      chan struct{}
}

func NewMessageCleanupScheduler(
	messageRepo interfaces.MessageRepository,
	interval time.Duration,
	retentionDays int,
	enabled bool,
) *MessageCleanupScheduler {
	return &MessageCleanupScheduler{
		messageRepo:   messageRepo,
		interval:      interval,
		retentionDays: retentionDays,
		enabled:       enabled,
		stopChan:      make(chan struct{}),
	}
}

func (s *MessageCleanupScheduler) Start() {
	if !s.enabled {
		log.Println("Message cleanup scheduler is disabled")
		return
	}

	log.Printf("Message cleanup scheduler started (interval: %v, retention: %d days)", s.interval, s.retentionDays)

	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		// Run immediately on start
		s.cleanupOldMessages()

		for {
			select {
			case <-ticker.C:
				s.cleanupOldMessages()
			case <-s.stopChan:
				log.Println("Message cleanup scheduler stopped")
				return
			}
		}
	}()
}

func (s *MessageCleanupScheduler) Stop() {
	if s.enabled {
		close(s.stopChan)
	}
}

func (s *MessageCleanupScheduler) cleanupOldMessages() {
	ctx := context.Background()
	log.Println("Cleaning up old chat messages...")

	// Get all messages older than retention period
	// List with large limit to get old messages
	messages, total, err := s.messageRepo.List(ctx, 0, 10000, "", "", true)
	if err != nil {
		log.Printf("Error listing messages: %v", err)
		return
	}

	if total == 0 {
		return
	}

	cutoff := time.Now().AddDate(0, 0, -s.retentionDays)
	deletedCount := 0

	for _, message := range messages {
		if message.CreatedAt.Before(cutoff) {
			if err := s.messageRepo.HardDelete(ctx, message.ID); err != nil {
				log.Printf("Error deleting message %s: %v", message.ID, err)
				continue
			}
			deletedCount++
		}
	}

	if deletedCount > 0 {
		log.Printf("Cleanup completed: deleted %d old messages (older than %d days)", deletedCount, s.retentionDays)
	}
}
