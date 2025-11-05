package usecase

import (
	"context"
	"fmt"
	"time"

	"toxictoast/services/link-service/internal/domain"
	"toxictoast/services/link-service/internal/repository"
)

// ClickUseCase defines the interface for click analytics business logic
type ClickUseCase interface {
	RecordClick(ctx context.Context, input RecordClickInput) (*domain.Click, error)
	GetLinkClicks(ctx context.Context, linkID string, page, pageSize int, startDate, endDate *time.Time) ([]domain.Click, int64, error)
	GetLinkAnalytics(ctx context.Context, linkID string) (*repository.ClickStats, error)
	GetClicksByDate(ctx context.Context, linkID string, startDate, endDate time.Time) (map[string]int, int, error)
}

// RecordClickInput defines the input for recording a click
type RecordClickInput struct {
	LinkID     string
	IPAddress  string
	UserAgent  string
	Referer    *string
	Country    *string
	City       *string
	DeviceType *string
}

type clickUseCase struct {
	clickRepo repository.ClickRepository
	linkRepo  repository.LinkRepository
}

// NewClickUseCase creates a new instance of ClickUseCase
func NewClickUseCase(clickRepo repository.ClickRepository, linkRepo repository.LinkRepository) ClickUseCase {
	return &clickUseCase{
		clickRepo: clickRepo,
		linkRepo:  linkRepo,
	}
}

func (uc *clickUseCase) RecordClick(ctx context.Context, input RecordClickInput) (*domain.Click, error) {
	// Verify link exists
	link, err := uc.linkRepo.GetByID(ctx, input.LinkID)
	if err != nil {
		return nil, fmt.Errorf("link not found: %w", err)
	}

	// Check if link is available
	if !link.IsAvailable() {
		return nil, fmt.Errorf("link is not available")
	}

	// Create click entity
	click := &domain.Click{
		LinkID:     input.LinkID,
		IPAddress:  input.IPAddress,
		UserAgent:  input.UserAgent,
		Referer:    input.Referer,
		Country:    input.Country,
		City:       input.City,
		DeviceType: input.DeviceType,
		ClickedAt:  time.Now(),
	}

	// Save to database
	if err := uc.clickRepo.Create(ctx, click); err != nil {
		return nil, fmt.Errorf("failed to record click: %w", err)
	}

	return click, nil
}

func (uc *clickUseCase) GetLinkClicks(ctx context.Context, linkID string, page, pageSize int, startDate, endDate *time.Time) ([]domain.Click, int64, error) {
	// Verify link exists
	_, err := uc.linkRepo.GetByID(ctx, linkID)
	if err != nil {
		return nil, 0, fmt.Errorf("link not found: %w", err)
	}

	// Default pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// Get clicks with or without date range
	if startDate != nil && endDate != nil {
		return uc.clickRepo.GetByDateRange(ctx, linkID, *startDate, *endDate, page, pageSize)
	}

	return uc.clickRepo.GetByLinkID(ctx, linkID, page, pageSize)
}

func (uc *clickUseCase) GetLinkAnalytics(ctx context.Context, linkID string) (*repository.ClickStats, error) {
	// Verify link exists
	_, err := uc.linkRepo.GetByID(ctx, linkID)
	if err != nil {
		return nil, fmt.Errorf("link not found: %w", err)
	}

	return uc.clickRepo.GetStats(ctx, linkID)
}

func (uc *clickUseCase) GetClicksByDate(ctx context.Context, linkID string, startDate, endDate time.Time) (map[string]int, int, error) {
	// Verify link exists
	_, err := uc.linkRepo.GetByID(ctx, linkID)
	if err != nil {
		return nil, 0, fmt.Errorf("link not found: %w", err)
	}

	// Get clicks by date
	clicksByDate, err := uc.clickRepo.GetClicksByDate(ctx, linkID, startDate, endDate)
	if err != nil {
		return nil, 0, err
	}

	// Calculate total clicks
	totalClicks := 0
	for _, count := range clicksByDate {
		totalClicks += count
	}

	return clicksByDate, totalClicks, nil
}
