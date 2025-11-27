package query

import (
	"context"
	"errors"
	"fmt"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"

	"toxictoast/services/blog-service/internal/domain"
	"toxictoast/services/blog-service/internal/repository"
)

// ============================================================================
// Queries
// ============================================================================

// GetMediaByIDQuery retrieves a media file by ID
type GetMediaByIDQuery struct {
	cqrs.BaseQuery
	MediaID string `json:"media_id"`
}

func (q *GetMediaByIDQuery) QueryName() string {
	return "get_media_by_id"
}

func (q *GetMediaByIDQuery) Validate() error {
	if q.MediaID == "" {
		return errors.New("media_id is required")
	}
	return nil
}

// ListMediaQuery retrieves a list of media files with filtering
type ListMediaQuery struct {
	cqrs.BaseQuery
	Filters repository.MediaFilters `json:"filters"`
}

func (q *ListMediaQuery) QueryName() string {
	return "list_media"
}

func (q *ListMediaQuery) Validate() error {
	// No strict validation required for filtering
	return nil
}

// ============================================================================
// Query Results
// ============================================================================

// ListMediaResult contains the result of listing media
type ListMediaResult struct {
	Media []domain.Media
	Total int64
}

// ============================================================================
// Query Handlers
// ============================================================================

// GetMediaByIDHandler handles media retrieval by ID
type GetMediaByIDHandler struct {
	mediaRepo repository.MediaRepository
}

func NewGetMediaByIDHandler(mediaRepo repository.MediaRepository) *GetMediaByIDHandler {
	return &GetMediaByIDHandler{
		mediaRepo: mediaRepo,
	}
}

func (h *GetMediaByIDHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetMediaByIDQuery)

	media, err := h.mediaRepo.GetByID(ctx, q.MediaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get media: %w", err)
	}

	return media, nil
}

// ListMediaHandler handles media listing with filters
type ListMediaHandler struct {
	mediaRepo repository.MediaRepository
}

func NewListMediaHandler(mediaRepo repository.MediaRepository) *ListMediaHandler {
	return &ListMediaHandler{
		mediaRepo: mediaRepo,
	}
}

func (h *ListMediaHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*ListMediaQuery)

	media, total, err := h.mediaRepo.List(ctx, q.Filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list media: %w", err)
	}

	return &ListMediaResult{
		Media: media,
		Total: total,
	}, nil
}
