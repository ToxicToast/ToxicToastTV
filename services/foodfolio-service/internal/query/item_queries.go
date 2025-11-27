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

type GetItemByIDQuery struct {
	cqrs.BaseQuery
	ID string `json:"id"`
}

func (q *GetItemByIDQuery) QueryName() string {
	return "get_item_by_id"
}

func (q *GetItemByIDQuery) Validate() error {
	if q.ID == "" {
		return errors.New("id is required")
	}
	return nil
}

type GetItemBySlugQuery struct {
	cqrs.BaseQuery
	Slug string `json:"slug"`
}

func (q *GetItemBySlugQuery) QueryName() string {
	return "get_item_by_slug"
}

func (q *GetItemBySlugQuery) Validate() error {
	if q.Slug == "" {
		return errors.New("slug is required")
	}
	return nil
}

type GetItemWithVariantsQuery struct {
	cqrs.BaseQuery
	ID             string `json:"id"`
	IncludeDetails bool   `json:"include_details"`
}

func (q *GetItemWithVariantsQuery) QueryName() string {
	return "get_item_with_variants"
}

func (q *GetItemWithVariantsQuery) Validate() error {
	if q.ID == "" {
		return errors.New("id is required")
	}
	return nil
}

type ListItemsQuery struct {
	cqrs.BaseQuery
	Page           int     `json:"page"`
	PageSize       int     `json:"page_size"`
	CategoryID     *string `json:"category_id"`
	CompanyID      *string `json:"company_id"`
	TypeID         *string `json:"type_id"`
	Search         *string `json:"search"`
	IncludeDeleted bool    `json:"include_deleted"`
}

func (q *ListItemsQuery) QueryName() string {
	return "list_items"
}

func (q *ListItemsQuery) Validate() error {
	return nil
}

type SearchItemsQuery struct {
	cqrs.BaseQuery
	Query      string  `json:"query"`
	Page       int     `json:"page"`
	PageSize   int     `json:"page_size"`
	CategoryID *string `json:"category_id"`
	CompanyID  *string `json:"company_id"`
}

func (q *SearchItemsQuery) QueryName() string {
	return "search_items"
}

func (q *SearchItemsQuery) Validate() error {
	if q.Query == "" {
		return errors.New("search query is required")
	}
	return nil
}

// ============================================================================
// Query Results
// ============================================================================

type ListItemsResult struct {
	Items []*domain.Item
	Total int64
}

// ============================================================================
// Query Handlers
// ============================================================================

type GetItemByIDHandler struct {
	itemRepo interfaces.ItemRepository
}

func NewGetItemByIDHandler(itemRepo interfaces.ItemRepository) *GetItemByIDHandler {
	return &GetItemByIDHandler{
		itemRepo: itemRepo,
	}
}

func (h *GetItemByIDHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetItemByIDQuery)

	item, err := h.itemRepo.GetByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	if item == nil {
		return nil, errors.New("item not found")
	}

	return item, nil
}

type GetItemBySlugHandler struct {
	itemRepo interfaces.ItemRepository
}

func NewGetItemBySlugHandler(itemRepo interfaces.ItemRepository) *GetItemBySlugHandler {
	return &GetItemBySlugHandler{
		itemRepo: itemRepo,
	}
}

func (h *GetItemBySlugHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetItemBySlugQuery)

	item, err := h.itemRepo.GetBySlug(ctx, q.Slug)
	if err != nil {
		return nil, err
	}

	if item == nil {
		return nil, errors.New("item not found")
	}

	return item, nil
}

type GetItemWithVariantsHandler struct {
	itemRepo interfaces.ItemRepository
}

func NewGetItemWithVariantsHandler(itemRepo interfaces.ItemRepository) *GetItemWithVariantsHandler {
	return &GetItemWithVariantsHandler{
		itemRepo: itemRepo,
	}
}

func (h *GetItemWithVariantsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetItemWithVariantsQuery)

	item, err := h.itemRepo.GetWithVariants(ctx, q.ID, q.IncludeDetails)
	if err != nil {
		return nil, err
	}

	if item == nil {
		return nil, errors.New("item not found")
	}

	return item, nil
}

type ListItemsHandler struct {
	itemRepo interfaces.ItemRepository
}

func NewListItemsHandler(itemRepo interfaces.ItemRepository) *ListItemsHandler {
	return &ListItemsHandler{
		itemRepo: itemRepo,
	}
}

func (h *ListItemsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*ListItemsQuery)

	page := q.Page
	pageSize := q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	items, total, err := h.itemRepo.List(ctx, offset, pageSize, q.CategoryID, q.CompanyID, q.TypeID, q.Search, q.IncludeDeleted)
	if err != nil {
		return nil, err
	}

	return &ListItemsResult{
		Items: items,
		Total: total,
	}, nil
}

type SearchItemsHandler struct {
	itemRepo interfaces.ItemRepository
}

func NewSearchItemsHandler(itemRepo interfaces.ItemRepository) *SearchItemsHandler {
	return &SearchItemsHandler{
		itemRepo: itemRepo,
	}
}

func (h *SearchItemsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*SearchItemsQuery)

	page := q.Page
	pageSize := q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	items, total, err := h.itemRepo.Search(ctx, q.Query, offset, pageSize, q.CategoryID, q.CompanyID)
	if err != nil {
		return nil, err
	}

	return &ListItemsResult{
		Items: items,
		Total: total,
	}, nil
}
