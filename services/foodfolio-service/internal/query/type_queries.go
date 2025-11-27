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

// GetTypeByIDQuery retrieves a type by ID
type GetTypeByIDQuery struct {
	cqrs.BaseQuery
	ID string `json:"id"`
}

func (q *GetTypeByIDQuery) QueryName() string {
	return "get_type_by_id"
}

func (q *GetTypeByIDQuery) Validate() error {
	if q.ID == "" {
		return errors.New("id is required")
	}
	return nil
}

// GetTypeBySlugQuery retrieves a type by slug
type GetTypeBySlugQuery struct {
	cqrs.BaseQuery
	Slug string `json:"slug"`
}

func (q *GetTypeBySlugQuery) QueryName() string {
	return "get_type_by_slug"
}

func (q *GetTypeBySlugQuery) Validate() error {
	if q.Slug == "" {
		return errors.New("slug is required")
	}
	return nil
}

// ListTypesQuery retrieves a list of types
type ListTypesQuery struct {
	cqrs.BaseQuery
	Page           int    `json:"page"`
	PageSize       int    `json:"page_size"`
	Search         string `json:"search"`
	IncludeDeleted bool   `json:"include_deleted"`
}

func (q *ListTypesQuery) QueryName() string {
	return "list_types"
}

func (q *ListTypesQuery) Validate() error {
	return nil
}

// ============================================================================
// Query Results
// ============================================================================

// ListTypesResult contains the result of listing types
type ListTypesResult struct {
	Types []*domain.Type
	Total int64
}

// ============================================================================
// Query Handlers
// ============================================================================

// GetTypeByIDHandler handles type retrieval by ID
type GetTypeByIDHandler struct {
	typeRepo interfaces.TypeRepository
}

func NewGetTypeByIDHandler(typeRepo interfaces.TypeRepository) *GetTypeByIDHandler {
	return &GetTypeByIDHandler{
		typeRepo: typeRepo,
	}
}

func (h *GetTypeByIDHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetTypeByIDQuery)

	typeEntity, err := h.typeRepo.GetByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	if typeEntity == nil {
		return nil, errors.New("type not found")
	}

	return typeEntity, nil
}

// GetTypeBySlugHandler handles type retrieval by slug
type GetTypeBySlugHandler struct {
	typeRepo interfaces.TypeRepository
}

func NewGetTypeBySlugHandler(typeRepo interfaces.TypeRepository) *GetTypeBySlugHandler {
	return &GetTypeBySlugHandler{
		typeRepo: typeRepo,
	}
}

func (h *GetTypeBySlugHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetTypeBySlugQuery)

	typeEntity, err := h.typeRepo.GetBySlug(ctx, q.Slug)
	if err != nil {
		return nil, err
	}

	if typeEntity == nil {
		return nil, errors.New("type not found")
	}

	return typeEntity, nil
}

// ListTypesHandler handles type listing
type ListTypesHandler struct {
	typeRepo interfaces.TypeRepository
}

func NewListTypesHandler(typeRepo interfaces.TypeRepository) *ListTypesHandler {
	return &ListTypesHandler{
		typeRepo: typeRepo,
	}
}

func (h *ListTypesHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*ListTypesQuery)

	// Default pagination
	page := q.Page
	pageSize := q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	types, total, err := h.typeRepo.List(ctx, offset, pageSize, q.Search, q.IncludeDeleted)
	if err != nil {
		return nil, err
	}

	return &ListTypesResult{
		Types: types,
		Total: total,
	}, nil
}
