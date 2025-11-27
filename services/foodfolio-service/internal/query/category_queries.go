package query

import (
	"context"
	"errors"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"

	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/interfaces"
)

// ============================================================================
// Queries
// ============================================================================

type GetCategoryByIDQuery struct {
	cqrs.BaseQuery
	ID string `json:"id"`
}

func (q *GetCategoryByIDQuery) QueryName() string {
	return "get_category_by_id"
}

func (q *GetCategoryByIDQuery) Validate() error {
	if q.ID == "" {
		return errors.New("id is required")
	}
	return nil
}

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

type ListCategoriesQuery struct {
	cqrs.BaseQuery
	Page            int     `json:"page"`
	PageSize        int     `json:"page_size"`
	ParentID        *string `json:"parent_id"`
	IncludeChildren bool    `json:"include_children"`
	IncludeDeleted  bool    `json:"include_deleted"`
}

func (q *ListCategoriesQuery) QueryName() string {
	return "list_categories"
}

func (q *ListCategoriesQuery) Validate() error {
	return nil
}

// ============================================================================
// Query Results
// ============================================================================

type ListCategoriesResult struct {
	Categories []*domain.Category
	Total      int64
}

// ============================================================================
// Query Handlers
// ============================================================================

type GetCategoryByIDHandler struct {
	categoryRepo interfaces.CategoryRepository
}

func NewGetCategoryByIDHandler(categoryRepo interfaces.CategoryRepository) *GetCategoryByIDHandler {
	return &GetCategoryByIDHandler{
		categoryRepo: categoryRepo,
	}
}

func (h *GetCategoryByIDHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetCategoryByIDQuery)

	categoryEntity, err := h.categoryRepo.GetByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	if categoryEntity == nil {
		return nil, errors.New("category not found")
	}

	return categoryEntity, nil
}

type GetCategoryBySlugHandler struct {
	categoryRepo interfaces.CategoryRepository
}

func NewGetCategoryBySlugHandler(categoryRepo interfaces.CategoryRepository) *GetCategoryBySlugHandler {
	return &GetCategoryBySlugHandler{
		categoryRepo: categoryRepo,
	}
}

func (h *GetCategoryBySlugHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetCategoryBySlugQuery)

	categoryEntity, err := h.categoryRepo.GetBySlug(ctx, q.Slug)
	if err != nil {
		return nil, err
	}

	if categoryEntity == nil {
		return nil, errors.New("category not found")
	}

	return categoryEntity, nil
}

type ListCategoriesHandler struct {
	categoryRepo interfaces.CategoryRepository
}

func NewListCategoriesHandler(categoryRepo interfaces.CategoryRepository) *ListCategoriesHandler {
	return &ListCategoriesHandler{
		categoryRepo: categoryRepo,
	}
}

func (h *ListCategoriesHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*ListCategoriesQuery)

	page := q.Page
	pageSize := q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	categories, total, err := h.categoryRepo.List(ctx, offset, pageSize, q.ParentID, q.IncludeChildren, q.IncludeDeleted)
	if err != nil {
		return nil, err
	}

	return &ListCategoriesResult{
		Categories: categories,
		Total:      total,
	}, nil
}

type GetCategoryTreeQuery struct {
	cqrs.BaseQuery
	RootID   *string `json:"root_id"`
	MaxDepth int     `json:"max_depth"`
}

func (q *GetCategoryTreeQuery) QueryName() string {
	return "get_category_tree"
}

func (q *GetCategoryTreeQuery) Validate() error {
	return nil
}

type GetCategoryTreeHandler struct {
	categoryRepo interfaces.CategoryRepository
}

func NewGetCategoryTreeHandler(categoryRepo interfaces.CategoryRepository) *GetCategoryTreeHandler {
	return &GetCategoryTreeHandler{
		categoryRepo: categoryRepo,
	}
}

func (h *GetCategoryTreeHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetCategoryTreeQuery)

	categories, err := h.categoryRepo.GetTree(ctx, q.RootID, q.MaxDepth)
	if err != nil {
		return nil, err
	}

	return categories, nil
}
