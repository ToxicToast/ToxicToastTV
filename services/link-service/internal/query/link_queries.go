package query

import (
	"context"
	"errors"
	"fmt"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"

	"toxictoast/services/link-service/internal/domain"
	"toxictoast/services/link-service/internal/repository"
)

// ============================================================================
// Queries
// ============================================================================

// GetLinkByIDQuery retrieves a link by ID
type GetLinkByIDQuery struct {
	cqrs.BaseQuery
	LinkID string `json:"link_id"`
}

func (q *GetLinkByIDQuery) QueryName() string {
	return "get_link_by_id"
}

func (q *GetLinkByIDQuery) Validate() error {
	if q.LinkID == "" {
		return errors.New("link_id is required")
	}
	return nil
}

// GetLinkByShortCodeQuery retrieves a link by short code
type GetLinkByShortCodeQuery struct {
	cqrs.BaseQuery
	ShortCode string `json:"short_code"`
}

func (q *GetLinkByShortCodeQuery) QueryName() string {
	return "get_link_by_short_code"
}

func (q *GetLinkByShortCodeQuery) Validate() error {
	if q.ShortCode == "" {
		return errors.New("short_code is required")
	}
	return nil
}

// ListLinksQuery retrieves a list of links with filtering
type ListLinksQuery struct {
	cqrs.BaseQuery
	Filters repository.LinkFilters `json:"filters"`
}

func (q *ListLinksQuery) QueryName() string {
	return "list_links"
}

func (q *ListLinksQuery) Validate() error {
	// No strict validation required for filtering
	return nil
}

// GetLinkStatsQuery retrieves statistics for a link
type GetLinkStatsQuery struct {
	cqrs.BaseQuery
	LinkID string `json:"link_id"`
}

func (q *GetLinkStatsQuery) QueryName() string {
	return "get_link_stats"
}

func (q *GetLinkStatsQuery) Validate() error {
	if q.LinkID == "" {
		return errors.New("link_id is required")
	}
	return nil
}

// ============================================================================
// Query Results
// ============================================================================

// ListLinksResult contains the result of listing links
type ListLinksResult struct {
	Links []domain.Link
	Total int64
}

// ============================================================================
// Query Handlers
// ============================================================================

// GetLinkByIDHandler handles link retrieval by ID
type GetLinkByIDHandler struct {
	linkRepo repository.LinkRepository
}

func NewGetLinkByIDHandler(linkRepo repository.LinkRepository) *GetLinkByIDHandler {
	return &GetLinkByIDHandler{
		linkRepo: linkRepo,
	}
}

func (h *GetLinkByIDHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetLinkByIDQuery)

	link, err := h.linkRepo.GetByID(ctx, q.LinkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get link: %w", err)
	}

	return link, nil
}

// GetLinkByShortCodeHandler handles link retrieval by short code
type GetLinkByShortCodeHandler struct {
	linkRepo repository.LinkRepository
}

func NewGetLinkByShortCodeHandler(linkRepo repository.LinkRepository) *GetLinkByShortCodeHandler {
	return &GetLinkByShortCodeHandler{
		linkRepo: linkRepo,
	}
}

func (h *GetLinkByShortCodeHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetLinkByShortCodeQuery)

	link, err := h.linkRepo.GetByShortCode(ctx, q.ShortCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get link: %w", err)
	}

	return link, nil
}

// ListLinksHandler handles link listing with filters
type ListLinksHandler struct {
	linkRepo repository.LinkRepository
}

func NewListLinksHandler(linkRepo repository.LinkRepository) *ListLinksHandler {
	return &ListLinksHandler{
		linkRepo: linkRepo,
	}
}

func (h *ListLinksHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*ListLinksQuery)

	links, total, err := h.linkRepo.List(ctx, q.Filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list links: %w", err)
	}

	return &ListLinksResult{
		Links: links,
		Total: total,
	}, nil
}

// GetLinkStatsHandler handles link statistics retrieval
type GetLinkStatsHandler struct {
	linkRepo repository.LinkRepository
}

func NewGetLinkStatsHandler(linkRepo repository.LinkRepository) *GetLinkStatsHandler {
	return &GetLinkStatsHandler{
		linkRepo: linkRepo,
	}
}

func (h *GetLinkStatsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetLinkStatsQuery)

	stats, err := h.linkRepo.GetStats(ctx, q.LinkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get link stats: %w", err)
	}

	return stats, nil
}
