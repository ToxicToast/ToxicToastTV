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

type GetShoppinglistByIDQuery struct {
	cqrs.BaseQuery
	ID string `json:"id"`
}

func (q *GetShoppinglistByIDQuery) QueryName() string {
	return "get_shoppinglist_by_id"
}

func (q *GetShoppinglistByIDQuery) Validate() error {
	if q.ID == "" {
		return errors.New("id is required")
	}
	return nil
}

type ListShoppinglistsQuery struct {
	cqrs.BaseQuery
	Page           int  `json:"page"`
	PageSize       int  `json:"page_size"`
	IncludeDeleted bool `json:"include_deleted"`
}

func (q *ListShoppinglistsQuery) QueryName() string {
	return "list_shoppinglists"
}

func (q *ListShoppinglistsQuery) Validate() error {
	return nil
}

// ============================================================================
// Query Results
// ============================================================================

type ListShoppinglistsResult struct {
	Shoppinglists []*domain.Shoppinglist
	Total         int64
}

// ============================================================================
// Query Handlers
// ============================================================================

type GetShoppinglistByIDHandler struct {
	shoppinglistRepo interfaces.ShoppinglistRepository
}

func NewGetShoppinglistByIDHandler(shoppinglistRepo interfaces.ShoppinglistRepository) *GetShoppinglistByIDHandler {
	return &GetShoppinglistByIDHandler{
		shoppinglistRepo: shoppinglistRepo,
	}
}

func (h *GetShoppinglistByIDHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetShoppinglistByIDQuery)

	shoppinglist, err := h.shoppinglistRepo.GetByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	if shoppinglist == nil {
		return nil, errors.New("shoppinglist not found")
	}

	return shoppinglist, nil
}

type ListShoppinglistsHandler struct {
	shoppinglistRepo interfaces.ShoppinglistRepository
}

func NewListShoppinglistsHandler(shoppinglistRepo interfaces.ShoppinglistRepository) *ListShoppinglistsHandler {
	return &ListShoppinglistsHandler{
		shoppinglistRepo: shoppinglistRepo,
	}
}

func (h *ListShoppinglistsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*ListShoppinglistsQuery)

	page := q.Page
	pageSize := q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	shoppinglists, total, err := h.shoppinglistRepo.List(ctx, offset, pageSize, q.IncludeDeleted)
	if err != nil {
		return nil, err
	}

	return &ListShoppinglistsResult{
		Shoppinglists: shoppinglists,
		Total:         total,
	}, nil
}
