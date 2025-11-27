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

type GetSizeByIDQuery struct {
	cqrs.BaseQuery
	ID string `json:"id"`
}

func (q *GetSizeByIDQuery) QueryName() string {
	return "get_size_by_id"
}

func (q *GetSizeByIDQuery) Validate() error {
	if q.ID == "" {
		return errors.New("id is required")
	}
	return nil
}

type ListSizesQuery struct {
	cqrs.BaseQuery
	Page           int      `json:"page"`
	PageSize       int      `json:"page_size"`
	Unit           string   `json:"unit"`
	MinValue       *float64 `json:"min_value"`
	MaxValue       *float64 `json:"max_value"`
	IncludeDeleted bool     `json:"include_deleted"`
}

func (q *ListSizesQuery) QueryName() string {
	return "list_sizes"
}

func (q *ListSizesQuery) Validate() error {
	return nil
}

// ============================================================================
// Query Results
// ============================================================================

type ListSizesResult struct {
	Sizes []*domain.Size
	Total int64
}

// ============================================================================
// Query Handlers
// ============================================================================

type GetSizeByIDHandler struct {
	sizeRepo interfaces.SizeRepository
}

func NewGetSizeByIDHandler(sizeRepo interfaces.SizeRepository) *GetSizeByIDHandler {
	return &GetSizeByIDHandler{
		sizeRepo: sizeRepo,
	}
}

func (h *GetSizeByIDHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetSizeByIDQuery)

	sizeEntity, err := h.sizeRepo.GetByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	if sizeEntity == nil {
		return nil, errors.New("size not found")
	}

	return sizeEntity, nil
}

type ListSizesHandler struct {
	sizeRepo interfaces.SizeRepository
}

func NewListSizesHandler(sizeRepo interfaces.SizeRepository) *ListSizesHandler {
	return &ListSizesHandler{
		sizeRepo: sizeRepo,
	}
}

func (h *ListSizesHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*ListSizesQuery)

	page := q.Page
	pageSize := q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	sizes, total, err := h.sizeRepo.List(ctx, offset, pageSize, q.Unit, q.MinValue, q.MaxValue, q.IncludeDeleted)
	if err != nil {
		return nil, err
	}

	return &ListSizesResult{
		Sizes: sizes,
		Total: total,
	}, nil
}
