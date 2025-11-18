package scheduler

import (
	"context"
	"log"
	"time"

	"toxictoast/services/blog-service/internal/domain"
	"toxictoast/services/blog-service/internal/repository"
	"toxictoast/services/blog-service/internal/usecase"
)

type PostPublisherScheduler struct {
	postUseCase usecase.PostUseCase
	postRepo    repository.PostRepository
	interval    time.Duration
	enabled     bool
	stopChan    chan struct{}
}

func NewPostPublisherScheduler(
	postUseCase usecase.PostUseCase,
	postRepo repository.PostRepository,
	interval time.Duration,
	enabled bool,
) *PostPublisherScheduler {
	return &PostPublisherScheduler{
		postUseCase: postUseCase,
		postRepo:    postRepo,
		interval:    interval,
		enabled:     enabled,
		stopChan:    make(chan struct{}),
	}
}

func (s *PostPublisherScheduler) Start() {
	if !s.enabled {
		log.Println("Post publisher scheduler is disabled")
		return
	}

	log.Printf("Post publisher scheduler started (interval: %v)", s.interval)

	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		// Run immediately on start
		s.checkScheduledPosts()

		for {
			select {
			case <-ticker.C:
				s.checkScheduledPosts()
			case <-s.stopChan:
				log.Println("Post publisher scheduler stopped")
				return
			}
		}
	}()
}

func (s *PostPublisherScheduler) Stop() {
	if s.enabled {
		close(s.stopChan)
	}
}

func (s *PostPublisherScheduler) checkScheduledPosts() {
	ctx := context.Background()
	log.Println("Checking for scheduled posts ready to publish...")

	// Find all draft posts with PublishedAt in the past
	draftStatus := domain.PostStatusDraft
	filters := repository.PostFilters{
		Status:   &draftStatus,
		Page:     1,
		PageSize: 100,
	}

	posts, total, err := s.postRepo.List(ctx, filters)
	if err != nil {
		log.Printf("Error listing posts for scheduled publishing: %v", err)
		return
	}

	log.Printf("Found %d draft posts to check", total)

	publishedCount := 0
	errorCount := 0
	now := time.Now()

	for i := range posts {
		post := &posts[i]

		// Check if post has PublishedAt set and it's in the past
		if post.PublishedAt != nil && post.PublishedAt.Before(now) {
			err := s.postUseCase.PublishScheduledPost(ctx, post)
			if err != nil {
				log.Printf("Error publishing scheduled post %s: %v", post.ID, err)
				errorCount++
				continue
			}

			log.Printf("Published scheduled post: %s (scheduled for: %v)", post.Title, post.PublishedAt)
			publishedCount++

			time.Sleep(10 * time.Millisecond)
		}
	}

	if publishedCount > 0 || errorCount > 0 {
		log.Printf("Scheduled post check completed: %d published, %d errors", publishedCount, errorCount)
	}
}
