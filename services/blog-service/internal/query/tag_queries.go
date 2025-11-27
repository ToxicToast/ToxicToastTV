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

// GetTagByIDQuery retrieves a tag by ID
type GetTagByIDQuery struct {
	cqrs.BaseQuery
	TagID string `json:"tag_id"`
}

func (q *GetTagByIDQuery) QueryName() string {
	return "get_tag_by_id"
}

func (q *GetTagByIDQuery) Validate() error {
	if q.TagID == "" {
		return errors.New("tag_id is required")
	}
	return nil
}

// GetTagBySlugQuery retrieves a tag by slug
type GetTagBySlugQuery struct {
	cqrs.BaseQuery
	Slug string `json:"slug"`
}

func (q *GetTagBySlugQuery) QueryName() string {
	return "get_tag_by_slug"
}

func (q *GetTagBySlugQuery) Validate() error {
	if q.Slug == "" {
		return errors.New("slug is required")
	}
	return nil
}

// ListTagsQuery retrieves a list of tags with filtering
type ListTagsQuery struct {
	cqrs.BaseQuery
	Filters repository.TagFilters `json:"filters"`
}

func (q *ListTagsQuery) QueryName() string {
	return "list_tags"
}

func (q *ListTagsQuery) Validate() error {
	// No strict validation required for filtering
	return nil
}

// ============================================================================
// Query Results
// ============================================================================

// ListTagsResult contains the result of listing tags
type ListTagsResult struct {
	Tags  []domain.Tag
	Total int64
}

// ============================================================================
// Query Handlers
// ============================================================================

// GetTagByIDHandler handles tag retrieval by ID
type GetTagByIDHandler struct {
	tagRepo repository.TagRepository
}

func NewGetTagByIDHandler(tagRepo repository.TagRepository) *GetTagByIDHandler {
	return &GetTagByIDHandler{
		tagRepo: tagRepo,
	}
}

func (h *GetTagByIDHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetTagByIDQuery)

	tag, err := h.tagRepo.GetByID(ctx, q.TagID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	return tag, nil
}

// GetTagBySlugHandler handles tag retrieval by slug
type GetTagBySlugHandler struct {
	tagRepo repository.TagRepository
}

func NewGetTagBySlugHandler(tagRepo repository.TagRepository) *GetTagBySlugHandler {
	return &GetTagBySlugHandler{
		tagRepo: tagRepo,
	}
}

func (h *GetTagBySlugHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetTagBySlugQuery)

	tag, err := h.tagRepo.GetBySlug(ctx, q.Slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag by slug: %w", err)
	}

	return tag, nil
}

// ListTagsHandler handles tag listing with filters
type ListTagsHandler struct {
	tagRepo repository.TagRepository
}

func NewListTagsHandler(tagRepo repository.TagRepository) *ListTagsHandler {
	return &ListTagsHandler{
		tagRepo: tagRepo,
	}
}

func (h *ListTagsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*ListTagsQuery)

	tags, total, err := h.tagRepo.List(ctx, q.Filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	return &ListTagsResult{
		Tags:  tags,
		Total: total,
	}, nil
}
