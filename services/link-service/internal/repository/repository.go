package repository

import (
	"context"
	"time"

	"toxictoast/services/link-service/internal/domain"
)

// LinkRepository defines the interface for link data operations
type LinkRepository interface {
	Create(ctx context.Context, link *domain.Link) error
	GetByID(ctx context.Context, id string) (*domain.Link, error)
	GetByShortCode(ctx context.Context, shortCode string) (*domain.Link, error)
	List(ctx context.Context, filters LinkFilters) ([]domain.Link, int64, error)
	Update(ctx context.Context, link *domain.Link) error
	Delete(ctx context.Context, id string) error
	IncrementClicks(ctx context.Context, id string) error
	ShortCodeExists(ctx context.Context, shortCode string) (bool, error)
	GetStats(ctx context.Context, linkID string) (*LinkStats, error)
}

// LinkFilters defines filters for listing links
type LinkFilters struct {
	Page           int
	PageSize       int
	IsActive       *bool
	IncludeExpired bool
	Search         *string
	SortBy         string
	SortOrder      string
}

// LinkStats holds statistics for a link
type LinkStats struct {
	LinkID       string
	TotalClicks  int
	UniqueIPs    int
	ClicksToday  int
	ClicksWeek   int
	ClicksMonth  int
}

// ClickRepository defines the interface for click analytics data operations
type ClickRepository interface {
	Create(ctx context.Context, click *domain.Click) error
	GetByLinkID(ctx context.Context, linkID string, page, pageSize int) ([]domain.Click, int64, error)
	GetByDateRange(ctx context.Context, linkID string, startDate, endDate time.Time, page, pageSize int) ([]domain.Click, int64, error)
	GetStats(ctx context.Context, linkID string) (*ClickStats, error)
	GetClicksByDate(ctx context.Context, linkID string, startDate, endDate time.Time) (map[string]int, error)
}

// ClickStats holds aggregated click statistics
type ClickStats struct {
	LinkID           string
	TotalClicks      int
	UniqueIPs        int
	ClicksToday      int
	ClicksThisWeek   int
	ClicksThisMonth  int
	ClicksByCountry  map[string]int
	ClicksByDevice   map[string]int
	TopReferers      []string
}
