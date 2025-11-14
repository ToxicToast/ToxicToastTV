package scheduler

import (
	"context"
	"log"
	"time"

	"toxictoast/services/twitchbot-service/internal/repository/interfaces"
)

type StreamSessionCloserScheduler struct {
	streamRepo      interfaces.StreamRepository
	interval        time.Duration
	inactiveTimeout time.Duration
	enabled         bool
	stopChan        chan struct{}
}

func NewStreamSessionCloserScheduler(
	streamRepo interfaces.StreamRepository,
	interval time.Duration,
	inactiveTimeout time.Duration,
	enabled bool,
) *StreamSessionCloserScheduler {
	return &StreamSessionCloserScheduler{
		streamRepo:      streamRepo,
		interval:        interval,
		inactiveTimeout: inactiveTimeout,
		enabled:         enabled,
		stopChan:        make(chan struct{}),
	}
}

func (s *StreamSessionCloserScheduler) Start() {
	if !s.enabled {
		log.Println("Stream session closer scheduler is disabled")
		return
	}

	log.Printf("Stream session closer scheduler started (interval: %v, timeout: %v)", s.interval, s.inactiveTimeout)

	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		// Run immediately on start
		s.closeInactiveSessions()

		for {
			select {
			case <-ticker.C:
				s.closeInactiveSessions()
			case <-s.stopChan:
				log.Println("Stream session closer scheduler stopped")
				return
			}
		}
	}()
}

func (s *StreamSessionCloserScheduler) Stop() {
	if s.enabled {
		close(s.stopChan)
	}
}

func (s *StreamSessionCloserScheduler) closeInactiveSessions() {
	ctx := context.Background()
	log.Println("Checking for inactive stream sessions...")

	// Get all active streams
	streams, total, err := s.streamRepo.List(ctx, 0, 100, true, "", false)
	if err != nil {
		log.Printf("Error listing active streams: %v", err)
		return
	}

	if total == 0 {
		return
	}

	now := time.Now()
	closedCount := 0

	for _, stream := range streams {
		// Check if stream hasn't been updated in a long time
		inactiveDuration := now.Sub(stream.UpdatedAt)

		if inactiveDuration > s.inactiveTimeout && stream.IsActive {
			log.Printf("Closing inactive stream session %s (inactive for %v)", stream.ID, inactiveDuration.Round(time.Second))

			if err := s.streamRepo.EndStream(ctx, stream.ID); err != nil {
				log.Printf("Error ending stream %s: %v", stream.ID, err)
				continue
			}
			closedCount++
		}
	}

	if closedCount > 0 {
		log.Printf("Session closer completed: closed %d inactive stream sessions", closedCount)
	}
}
