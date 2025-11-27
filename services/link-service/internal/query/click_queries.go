package query

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"

	"toxictoast/services/link-service/internal/domain"
	"toxictoast/services/link-service/internal/repository"
)

// ============================================================================
// Queries
// ============================================================================

// GetLinkClicksQuery retrieves clicks for a link with optional date filtering
type GetLinkClicksQuery struct {
	cqrs.BaseQuery
	LinkID    string     `json:"link_id"`
	Page      int        `json:"page"`
	PageSize  int        `json:"page_size"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
}

func (q *GetLinkClicksQuery) QueryName() string {
	return "get_link_clicks"
}

func (q *GetLinkClicksQuery) Validate() error {
	if q.LinkID == "" {
		return errors.New("link_id is required")
	}
	return nil
}

// GetLinkAnalyticsQuery retrieves analytics/statistics for a link
type GetLinkAnalyticsQuery struct {
	cqrs.BaseQuery
	LinkID string `json:"link_id"`
}

func (q *GetLinkAnalyticsQuery) QueryName() string {
	return "get_link_analytics"
}

func (q *GetLinkAnalyticsQuery) Validate() error {
	if q.LinkID == "" {
		return errors.New("link_id is required")
	}
	return nil
}

// GetClicksByDateQuery retrieves clicks grouped by date for a link
type GetClicksByDateQuery struct {
	cqrs.BaseQuery
	LinkID    string    `json:"link_id"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

func (q *GetClicksByDateQuery) QueryName() string {
	return "get_clicks_by_date"
}

func (q *GetClicksByDateQuery) Validate() error {
	if q.LinkID == "" {
		return errors.New("link_id is required")
	}
	if q.StartDate.IsZero() {
		return errors.New("start_date is required")
	}
	if q.EndDate.IsZero() {
		return errors.New("end_date is required")
	}
	return nil
}

// ============================================================================
// Query Results
// ============================================================================

// GetLinkClicksResult contains the result of listing clicks
type GetLinkClicksResult struct {
	Clicks []domain.Click
	Total  int64
}

// GetClicksByDateResult contains clicks grouped by date
type GetClicksByDateResult struct {
	ClicksByDate map[string]int
	TotalClicks  int
}

// ============================================================================
// Query Handlers
// ============================================================================

// GetLinkClicksHandler handles click listing for a link
type GetLinkClicksHandler struct {
	clickRepo repository.ClickRepository
	linkRepo  repository.LinkRepository
}

func NewGetLinkClicksHandler(clickRepo repository.ClickRepository, linkRepo repository.LinkRepository) *GetLinkClicksHandler {
	return &GetLinkClicksHandler{
		clickRepo: clickRepo,
		linkRepo:  linkRepo,
	}
}

func (h *GetLinkClicksHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetLinkClicksQuery)

	// Verify link exists
	_, err := h.linkRepo.GetByID(ctx, q.LinkID)
	if err != nil {
		return nil, fmt.Errorf("link not found: %w", err)
	}

	// Default pagination
	page := q.Page
	pageSize := q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// Get clicks with or without date range
	var clicks []domain.Click
	var total int64

	if q.StartDate != nil && q.EndDate != nil {
		clicks, total, err = h.clickRepo.GetByDateRange(ctx, q.LinkID, *q.StartDate, *q.EndDate, page, pageSize)
	} else {
		clicks, total, err = h.clickRepo.GetByLinkID(ctx, q.LinkID, page, pageSize)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get clicks: %w", err)
	}

	return &GetLinkClicksResult{
		Clicks: clicks,
		Total:  total,
	}, nil
}

// GetLinkAnalyticsHandler handles analytics retrieval for a link
type GetLinkAnalyticsHandler struct {
	clickRepo repository.ClickRepository
	linkRepo  repository.LinkRepository
}

func NewGetLinkAnalyticsHandler(clickRepo repository.ClickRepository, linkRepo repository.LinkRepository) *GetLinkAnalyticsHandler {
	return &GetLinkAnalyticsHandler{
		clickRepo: clickRepo,
		linkRepo:  linkRepo,
	}
}

func (h *GetLinkAnalyticsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetLinkAnalyticsQuery)

	// Verify link exists
	_, err := h.linkRepo.GetByID(ctx, q.LinkID)
	if err != nil {
		return nil, fmt.Errorf("link not found: %w", err)
	}

	stats, err := h.clickRepo.GetStats(ctx, q.LinkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}

	return stats, nil
}

// GetClicksByDateHandler handles click counts grouped by date
type GetClicksByDateHandler struct {
	clickRepo repository.ClickRepository
	linkRepo  repository.LinkRepository
}

func NewGetClicksByDateHandler(clickRepo repository.ClickRepository, linkRepo repository.LinkRepository) *GetClicksByDateHandler {
	return &GetClicksByDateHandler{
		clickRepo: clickRepo,
		linkRepo:  linkRepo,
	}
}

func (h *GetClicksByDateHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetClicksByDateQuery)

	// Verify link exists
	_, err := h.linkRepo.GetByID(ctx, q.LinkID)
	if err != nil {
		return nil, fmt.Errorf("link not found: %w", err)
	}

	// Get clicks by date
	clicksByDate, err := h.clickRepo.GetClicksByDate(ctx, q.LinkID, q.StartDate, q.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get clicks by date: %w", err)
	}

	// Calculate total clicks
	totalClicks := 0
	for _, count := range clicksByDate {
		totalClicks += count
	}

	return &GetClicksByDateResult{
		ClicksByDate: clicksByDate,
		TotalClicks:  totalClicks,
	}, nil
}
