package scheduler

import (
	"context"
	"log"
	"time"

	"toxictoast/services/notification-service/internal/repository/interfaces"
)

type NotificationCleanupScheduler struct {
	notificationRepo interfaces.NotificationRepository
	interval         time.Duration
	retentionDays    int
	enabled          bool
	stopChan         chan struct{}
}

func NewNotificationCleanupScheduler(
	notificationRepo interfaces.NotificationRepository,
	interval time.Duration,
	retentionDays int,
	enabled bool,
) *NotificationCleanupScheduler {
	return &NotificationCleanupScheduler{
		notificationRepo: notificationRepo,
		interval:         interval,
		retentionDays:    retentionDays,
		enabled:          enabled,
		stopChan:         make(chan struct{}),
	}
}

func (s *NotificationCleanupScheduler) Start() {
	if !s.enabled {
		log.Println("Notification cleanup scheduler is disabled")
		return
	}

	log.Printf("Notification cleanup scheduler started (interval: %v, retention: %d days)", s.interval, s.retentionDays)

	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		// Run immediately on start
		s.cleanupOldNotifications()

		for {
			select {
			case <-ticker.C:
				s.cleanupOldNotifications()
			case <-s.stopChan:
				log.Println("Notification cleanup scheduler stopped")
				return
			}
		}
	}()
}

func (s *NotificationCleanupScheduler) Stop() {
	if s.enabled {
		close(s.stopChan)
	}
}

func (s *NotificationCleanupScheduler) cleanupOldNotifications() {
	ctx := context.Background()
	log.Println("Cleaning up old notifications...")

	// Calculate cutoff date
	cutoffDate := time.Now().AddDate(0, 0, -s.retentionDays)

	// Delete old successful notifications
	deletedCount, err := s.notificationRepo.DeleteOldSuccessfulNotifications(ctx, cutoffDate)
	if err != nil {
		log.Printf("Error cleaning up old notifications: %v", err)
		return
	}

	if deletedCount > 0 {
		log.Printf("Cleanup completed: deleted %d old notifications (older than %d days)", deletedCount, s.retentionDays)
	}
}
