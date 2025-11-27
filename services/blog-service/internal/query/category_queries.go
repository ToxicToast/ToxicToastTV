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

// GetCategoryByIDQuery retrieves a category by ID
type GetCategoryByIDQuery struct {
	cqrs.BaseQuery
	CategoryID string `json:"category_id"`
}

func (q *GetCategoryByIDQuery) QueryName() string {
	return "get_category_by_id"
}

func (q *GetCategoryByIDQuery) Validate() error {
	if q.CategoryID == "" {
		return errors.New("category_id is required")
	}
	return nil
}

// GetCategoryBySlugQuery retrieves a category by slug
type GetCategoryBySlugQuery struct {
	cqrs.BaseQuery
	Slug string `json:"slug"`
}

func (q *GetCategoryBySlugQuery) QueryName() string {
	return "get_category_by_slug"
}

func (q *GetCategoryBySlugQuery) Validate() error {
	if q.Slug == "" {
		return errors.New("slug is required")
	}
	return nil
}

// ListCategoriesQuery retrieves a list of categories with filtering
type ListCategoriesQuery struct {
	cqrs.BaseQuery
	Filters repository.CategoryFilters `json:"filters"`
}

func (q *ListCategoriesQuery) QueryName() string {
	return "list_categories"
}

func (q *ListCategoriesQuery) Validate() error {
	// No strict validation required for filtering
	return nil
}

// GetCategoryChildrenQuery retrieves child categories of a parent category
type GetCategoryChildrenQuery struct {
	cqrs.BaseQuery
	ParentID string `json:"parent_id"`
}

func (q *GetCategoryChildrenQuery) QueryName() string {
	return "get_category_children"
}

func (q *GetCategoryChildrenQuery) Validate() error {
	if q.ParentID == "" {
		return errors.New("parent_id is required")
	}
	return nil
}

// ============================================================================
// Query Results
// ============================================================================

// ListCategoriesResult contains the result of listing categories
type ListCategoriesResult struct {
	Categories []domain.Category
	Total      int64
}

// ============================================================================
// Query Handlers
// ============================================================================

// GetCategoryByIDHandler handles category retrieval by ID
type GetCategoryByIDHandler struct {
	categoryRepo repository.CategoryRepository
}

func NewGetCategoryByIDHandler(categoryRepo repository.CategoryRepository) *GetCategoryByIDHandler {
	return &GetCategoryByIDHandler{
		categoryRepo: categoryRepo,
	}
}

func (h *GetCategoryByIDHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetCategoryByIDQuery)

	category, err := h.categoryRepo.GetByID(ctx, q.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return category, nil
}

// GetCategoryBySlugHandler handles category retrieval by slug
type GetCategoryBySlugHandler struct {
	categoryRepo repository.CategoryRepository
}

func NewGetCategoryBySlugHandler(categoryRepo repository.CategoryRepository) *GetCategoryBySlugHandler {
	return &GetCategoryBySlugHandler{
		categoryRepo: categoryRepo,
	}
}

func (h *GetCategoryBySlugHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetCategoryBySlugQuery)

	category, err := h.categoryRepo.GetBySlug(ctx, q.Slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get category by slug: %w", err)
	}

	return category, nil
}

// ListCategoriesHandler handles category listing with filters
type ListCategoriesHandler struct {
	categoryRepo repository.CategoryRepository
}

func NewListCategoriesHandler(categoryRepo repository.CategoryRepository) *ListCategoriesHandler {
	return &ListCategoriesHandler{
		categoryRepo: categoryRepo,
	}
}

func (h *ListCategoriesHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*ListCategoriesQuery)

	categories, total, err := h.categoryRepo.List(ctx, q.Filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}

	return &ListCategoriesResult{
		Categories: categories,
		Total:      total,
	}, nil
}

// GetCategoryChildrenHandler handles retrieving child categories
type GetCategoryChildrenHandler struct {
	categoryRepo repository.CategoryRepository
}

func NewGetCategoryChildrenHandler(categoryRepo repository.CategoryRepository) *GetCategoryChildrenHandler {
	return &GetCategoryChildrenHandler{
		categoryRepo: categoryRepo,
	}
}

func (h *GetCategoryChildrenHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetCategoryChildrenQuery)

	children, err := h.categoryRepo.GetChildren(ctx, q.ParentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get category children: %w", err)
	}

	return children, nil
}
