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

type GetItemDetailByIDQuery struct {
	cqrs.BaseQuery
	ID string `json:"id"`
}

func (q *GetItemDetailByIDQuery) QueryName() string {
	return "get_item_detail_by_id"
}

func (q *GetItemDetailByIDQuery) Validate() error {
	if q.ID == "" {
		return errors.New("id is required")
	}
	return nil
}

type ListItemDetailsQuery struct {
	cqrs.BaseQuery
	Page           int     `json:"page"`
	PageSize       int     `json:"page_size"`
	VariantID      *string `json:"variant_id"`
	WarehouseID    *string `json:"warehouse_id"`
	LocationID     *string `json:"location_id"`
	IsOpened       *bool   `json:"is_opened"`
	HasDeposit     *bool   `json:"has_deposit"`
	IsFrozen       *bool   `json:"is_frozen"`
	IncludeDeleted bool    `json:"include_deleted"`
}

func (q *ListItemDetailsQuery) QueryName() string {
	return "list_item_details"
}

func (q *ListItemDetailsQuery) Validate() error {
	return nil
}

type GetItemDetailsByVariantQuery struct {
	cqrs.BaseQuery
	VariantID string `json:"variant_id"`
}

func (q *GetItemDetailsByVariantQuery) QueryName() string {
	return "get_item_details_by_variant"
}

func (q *GetItemDetailsByVariantQuery) Validate() error {
	if q.VariantID == "" {
		return errors.New("variant_id is required")
	}
	return nil
}

type GetItemDetailsByLocationQuery struct {
	cqrs.BaseQuery
	LocationID      string `json:"location_id"`
	IncludeChildren bool   `json:"include_children"`
}

func (q *GetItemDetailsByLocationQuery) QueryName() string {
	return "get_item_details_by_location"
}

func (q *GetItemDetailsByLocationQuery) Validate() error {
	if q.LocationID == "" {
		return errors.New("location_id is required")
	}
	return nil
}

type GetExpiringItemsQuery struct {
	cqrs.BaseQuery
	Days     int `json:"days"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

func (q *GetExpiringItemsQuery) QueryName() string {
	return "get_expiring_items"
}

func (q *GetExpiringItemsQuery) Validate() error {
	if q.Days <= 0 {
		return errors.New("days must be greater than 0")
	}
	return nil
}

type GetExpiredItemsQuery struct {
	cqrs.BaseQuery
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

func (q *GetExpiredItemsQuery) QueryName() string {
	return "get_expired_items"
}

func (q *GetExpiredItemsQuery) Validate() error {
	return nil
}

type GetItemsWithDepositQuery struct {
	cqrs.BaseQuery
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

func (q *GetItemsWithDepositQuery) QueryName() string {
	return "get_items_with_deposit"
}

func (q *GetItemsWithDepositQuery) Validate() error {
	return nil
}

// ============================================================================
// Query Results
// ============================================================================

type ListItemDetailsResult struct {
	ItemDetails []*domain.ItemDetail
	Total       int64
}

// ============================================================================
// Query Handlers
// ============================================================================

type GetItemDetailByIDHandler struct {
	detailRepo interfaces.ItemDetailRepository
}

func NewGetItemDetailByIDHandler(detailRepo interfaces.ItemDetailRepository) *GetItemDetailByIDHandler {
	return &GetItemDetailByIDHandler{
		detailRepo: detailRepo,
	}
}

func (h *GetItemDetailByIDHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetItemDetailByIDQuery)

	detail, err := h.detailRepo.GetByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	if detail == nil {
		return nil, errors.New("item detail not found")
	}

	return detail, nil
}

type ListItemDetailsHandler struct {
	detailRepo interfaces.ItemDetailRepository
}

func NewListItemDetailsHandler(detailRepo interfaces.ItemDetailRepository) *ListItemDetailsHandler {
	return &ListItemDetailsHandler{
		detailRepo: detailRepo,
	}
}

func (h *ListItemDetailsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*ListItemDetailsQuery)

	page := q.Page
	pageSize := q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	details, total, err := h.detailRepo.List(ctx, offset, pageSize, q.VariantID, q.WarehouseID, q.LocationID, q.IsOpened, q.HasDeposit, q.IsFrozen, q.IncludeDeleted)
	if err != nil {
		return nil, err
	}

	return &ListItemDetailsResult{
		ItemDetails: details,
		Total:       total,
	}, nil
}

type GetItemDetailsByVariantHandler struct {
	detailRepo  interfaces.ItemDetailRepository
	variantRepo interfaces.ItemVariantRepository
}

func NewGetItemDetailsByVariantHandler(
	detailRepo interfaces.ItemDetailRepository,
	variantRepo interfaces.ItemVariantRepository,
) *GetItemDetailsByVariantHandler {
	return &GetItemDetailsByVariantHandler{
		detailRepo:  detailRepo,
		variantRepo: variantRepo,
	}
}

func (h *GetItemDetailsByVariantHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetItemDetailsByVariantQuery)

	// Validate variant exists
	variant, err := h.variantRepo.GetByID(ctx, q.VariantID)
	if err != nil {
		return nil, err
	}
	if variant == nil {
		return nil, errors.New("item variant not found")
	}

	details, err := h.detailRepo.GetByVariant(ctx, q.VariantID)
	if err != nil {
		return nil, err
	}

	return details, nil
}

type GetItemDetailsByLocationHandler struct {
	detailRepo   interfaces.ItemDetailRepository
	locationRepo interfaces.LocationRepository
}

func NewGetItemDetailsByLocationHandler(
	detailRepo interfaces.ItemDetailRepository,
	locationRepo interfaces.LocationRepository,
) *GetItemDetailsByLocationHandler {
	return &GetItemDetailsByLocationHandler{
		detailRepo:   detailRepo,
		locationRepo: locationRepo,
	}
}

func (h *GetItemDetailsByLocationHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetItemDetailsByLocationQuery)

	// Validate location exists
	location, err := h.locationRepo.GetByID(ctx, q.LocationID)
	if err != nil {
		return nil, err
	}
	if location == nil {
		return nil, errors.New("location not found")
	}

	details, err := h.detailRepo.GetByLocation(ctx, q.LocationID, q.IncludeChildren)
	if err != nil {
		return nil, err
	}

	return details, nil
}

type GetExpiringItemsHandler struct {
	detailRepo interfaces.ItemDetailRepository
}

func NewGetExpiringItemsHandler(detailRepo interfaces.ItemDetailRepository) *GetExpiringItemsHandler {
	return &GetExpiringItemsHandler{
		detailRepo: detailRepo,
	}
}

func (h *GetExpiringItemsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetExpiringItemsQuery)

	page := q.Page
	pageSize := q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	details, total, err := h.detailRepo.GetExpiringItems(ctx, q.Days, offset, pageSize)
	if err != nil {
		return nil, err
	}

	return &ListItemDetailsResult{
		ItemDetails: details,
		Total:       total,
	}, nil
}

type GetExpiredItemsHandler struct {
	detailRepo interfaces.ItemDetailRepository
}

func NewGetExpiredItemsHandler(detailRepo interfaces.ItemDetailRepository) *GetExpiredItemsHandler {
	return &GetExpiredItemsHandler{
		detailRepo: detailRepo,
	}
}

func (h *GetExpiredItemsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetExpiredItemsQuery)

	page := q.Page
	pageSize := q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	details, total, err := h.detailRepo.GetExpiredItems(ctx, offset, pageSize)
	if err != nil {
		return nil, err
	}

	return &ListItemDetailsResult{
		ItemDetails: details,
		Total:       total,
	}, nil
}

type GetItemsWithDepositHandler struct {
	detailRepo interfaces.ItemDetailRepository
}

func NewGetItemsWithDepositHandler(detailRepo interfaces.ItemDetailRepository) *GetItemsWithDepositHandler {
	return &GetItemsWithDepositHandler{
		detailRepo: detailRepo,
	}
}

func (h *GetItemsWithDepositHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetItemsWithDepositQuery)

	page := q.Page
	pageSize := q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	details, total, err := h.detailRepo.GetItemsWithDeposit(ctx, offset, pageSize)
	if err != nil {
		return nil, err
	}

	return &ListItemDetailsResult{
		ItemDetails: details,
		Total:       total,
	}, nil
}
