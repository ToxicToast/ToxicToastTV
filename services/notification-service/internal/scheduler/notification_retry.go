package scheduler

import (
	"context"
	"log"
	"time"

	"toxictoast/services/notification-service/internal/domain"
	"toxictoast/services/notification-service/internal/repository/interfaces"
	"toxictoast/services/notification-service/internal/usecase"
)

type NotificationRetryScheduler struct {
	notificationUseCase *usecase.NotificationUseCase
	notificationRepo    interfaces.NotificationRepository
	interval            time.Duration
	maxRetries          int
	enabled             bool
	stopChan            chan struct{}
}

func NewNotificationRetryScheduler(
	notificationUseCase *usecase.NotificationUseCase,
	notificationRepo interfaces.NotificationRepository,
	interval time.Duration,
	maxRetries int,
	enabled bool,
) *NotificationRetryScheduler {
	return &NotificationRetryScheduler{
		notificationUseCase: notificationUseCase,
		notificationRepo:    notificationRepo,
		interval:            interval,
		maxRetries:          maxRetries,
		enabled:             enabled,
		stopChan:            make(chan struct{}),
	}
}

func (s *NotificationRetryScheduler) Start() {
	if !s.enabled {
		log.Println("Notification retry scheduler is disabled")
		return
	}

	log.Printf("Notification retry scheduler started (interval: %v, max retries: %d)", s.interval, s.maxRetries)

	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		// Run immediately on start
		s.retryFailedNotifications()

		for {
			select {
			case <-ticker.C:
				s.retryFailedNotifications()
			case <-s.stopChan:
				log.Println("Notification retry scheduler stopped")
				return
			}
		}
	}()
}

func (s *NotificationRetryScheduler) Stop() {
	if s.enabled {
		close(s.stopChan)
	}
}

func (s *NotificationRetryScheduler) retryFailedNotifications() {
	ctx := context.Background()
	log.Println("Checking for failed notifications to retry...")

	// Get all failed notifications with attempts < maxRetries
	notifications, err := s.notificationRepo.GetFailedNotifications(ctx, s.maxRetries)
	if err != nil {
		log.Printf("Error listing failed notifications: %v", err)
		return
	}

	if len(notifications) == 0 {
		return
	}

	log.Printf("Found %d failed notifications to retry", len(notifications))

	retriedCount := 0
	errorCount := 0

	for i := range notifications {
		notification := notifications[i]

		// Skip if already at max retries
		if notification.AttemptCount >= s.maxRetries {
			continue
		}

		// Retry the notification
		err := s.notificationUseCase.RetryNotification(ctx, &notification)
		if err != nil {
			log.Printf("Error retrying notification %s: %v", notification.ID, err)
			errorCount++
			continue
		}

		if notification.Status == domain.NotificationStatusSuccess {
			log.Printf("Successfully retried notification: %s (attempt %d/%d)", notification.ID, notification.AttemptCount, s.maxRetries)
			retriedCount++
		} else {
			log.Printf("Retry failed for notification: %s (attempt %d/%d): %s", notification.ID, notification.AttemptCount, s.maxRetries, notification.LastError)
		}

		time.Sleep(100 * time.Millisecond) // Rate limiting
	}

	if retriedCount > 0 || errorCount > 0 {
		log.Printf("Notification retry completed: %d succeeded, %d errors", retriedCount, errorCount)
	}
}
