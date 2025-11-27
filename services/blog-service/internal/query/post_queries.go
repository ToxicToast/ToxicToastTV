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

// GetPostByIDQuery retrieves a post by ID
type GetPostByIDQuery struct {
	cqrs.BaseQuery
	PostID string `json:"post_id"`
}

func (q *GetPostByIDQuery) QueryName() string {
	return "get_post_by_id"
}

func (q *GetPostByIDQuery) Validate() error {
	if q.PostID == "" {
		return errors.New("post_id is required")
	}
	return nil
}

// GetPostBySlugQuery retrieves a post by slug
type GetPostBySlugQuery struct {
	cqrs.BaseQuery
	Slug string `json:"slug"`
}

func (q *GetPostBySlugQuery) QueryName() string {
	return "get_post_by_slug"
}

func (q *GetPostBySlugQuery) Validate() error {
	if q.Slug == "" {
		return errors.New("slug is required")
	}
	return nil
}

// ListPostsQuery retrieves a list of posts with filtering
type ListPostsQuery struct {
	cqrs.BaseQuery
	Filters repository.PostFilters `json:"filters"`
}

func (q *ListPostsQuery) QueryName() string {
	return "list_posts"
}

func (q *ListPostsQuery) Validate() error {
	// No strict validation required for filtering
	return nil
}

// ============================================================================
// Query Results
// ============================================================================

// ListPostsResult contains the result of listing posts
type ListPostsResult struct {
	Posts []domain.Post
	Total int64
}

// ============================================================================
// Query Handlers
// ============================================================================

// GetPostByIDHandler handles post retrieval by ID
type GetPostByIDHandler struct {
	postRepo repository.PostRepository
}

func NewGetPostByIDHandler(postRepo repository.PostRepository) *GetPostByIDHandler {
	return &GetPostByIDHandler{
		postRepo: postRepo,
	}
}

func (h *GetPostByIDHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetPostByIDQuery)

	post, err := h.postRepo.GetByID(ctx, q.PostID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	return post, nil
}

// GetPostBySlugHandler handles post retrieval by slug
type GetPostBySlugHandler struct {
	postRepo repository.PostRepository
}

func NewGetPostBySlugHandler(postRepo repository.PostRepository) *GetPostBySlugHandler {
	return &GetPostBySlugHandler{
		postRepo: postRepo,
	}
}

func (h *GetPostBySlugHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetPostBySlugQuery)

	post, err := h.postRepo.GetBySlug(ctx, q.Slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get post by slug: %w", err)
	}

	return post, nil
}

// ListPostsHandler handles post listing with filters
type ListPostsHandler struct {
	postRepo repository.PostRepository
}

func NewListPostsHandler(postRepo repository.PostRepository) *ListPostsHandler {
	return &ListPostsHandler{
		postRepo: postRepo,
	}
}

func (h *ListPostsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*ListPostsQuery)

	posts, total, err := h.postRepo.List(ctx, q.Filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list posts: %w", err)
	}

	return &ListPostsResult{
		Posts: posts,
		Total: total,
	}, nil
}
