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

type GetItemVariantByIDQuery struct {
	cqrs.BaseQuery
	ID string `json:"id"`
}

func (q *GetItemVariantByIDQuery) QueryName() string {
	return "get_item_variant_by_id"
}

func (q *GetItemVariantByIDQuery) Validate() error {
	if q.ID == "" {
		return errors.New("id is required")
	}
	return nil
}

type GetItemVariantByBarcodeQuery struct {
	cqrs.BaseQuery
	Barcode string `json:"barcode"`
}

func (q *GetItemVariantByBarcodeQuery) QueryName() string {
	return "get_item_variant_by_barcode"
}

func (q *GetItemVariantByBarcodeQuery) Validate() error {
	if q.Barcode == "" {
		return errors.New("barcode is required")
	}
	return nil
}

type ListItemVariantsQuery struct {
	cqrs.BaseQuery
	Page             int     `json:"page"`
	PageSize         int     `json:"page_size"`
	ItemID           *string `json:"item_id"`
	SizeID           *string `json:"size_id"`
	IsNormallyFrozen *bool   `json:"is_normally_frozen"`
	IncludeDeleted   bool    `json:"include_deleted"`
}

func (q *ListItemVariantsQuery) QueryName() string {
	return "list_item_variants"
}

func (q *ListItemVariantsQuery) Validate() error {
	return nil
}

type GetItemVariantsByItemQuery struct {
	cqrs.BaseQuery
	ItemID string `json:"item_id"`
}

func (q *GetItemVariantsByItemQuery) QueryName() string {
	return "get_item_variants_by_item"
}

func (q *GetItemVariantsByItemQuery) Validate() error {
	if q.ItemID == "" {
		return errors.New("item_id is required")
	}
	return nil
}

type GetLowStockVariantsQuery struct {
	cqrs.BaseQuery
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

func (q *GetLowStockVariantsQuery) QueryName() string {
	return "get_low_stock_variants"
}

func (q *GetLowStockVariantsQuery) Validate() error {
	return nil
}

type GetOverstockedVariantsQuery struct {
	cqrs.BaseQuery
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

func (q *GetOverstockedVariantsQuery) QueryName() string {
	return "get_overstocked_variants"
}

func (q *GetOverstockedVariantsQuery) Validate() error {
	return nil
}

type GetCurrentStockQuery struct {
	cqrs.BaseQuery
	VariantID string `json:"variant_id"`
}

func (q *GetCurrentStockQuery) QueryName() string {
	return "get_current_stock"
}

func (q *GetCurrentStockQuery) Validate() error {
	if q.VariantID == "" {
		return errors.New("variant_id is required")
	}
	return nil
}

// ============================================================================
// Query Results
// ============================================================================

type ListItemVariantsResult struct {
	ItemVariants []*domain.ItemVariant
	Total        int64
}

type CurrentStockResult struct {
	Stock         int
	NeedsRestock  bool
	IsOverstocked bool
}

// ============================================================================
// Query Handlers
// ============================================================================

type GetItemVariantByIDHandler struct {
	variantRepo interfaces.ItemVariantRepository
}

func NewGetItemVariantByIDHandler(variantRepo interfaces.ItemVariantRepository) *GetItemVariantByIDHandler {
	return &GetItemVariantByIDHandler{
		variantRepo: variantRepo,
	}
}

func (h *GetItemVariantByIDHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetItemVariantByIDQuery)

	variant, err := h.variantRepo.GetByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	if variant == nil {
		return nil, errors.New("item variant not found")
	}

	return variant, nil
}

type GetItemVariantByBarcodeHandler struct {
	variantRepo interfaces.ItemVariantRepository
}

func NewGetItemVariantByBarcodeHandler(variantRepo interfaces.ItemVariantRepository) *GetItemVariantByBarcodeHandler {
	return &GetItemVariantByBarcodeHandler{
		variantRepo: variantRepo,
	}
}

func (h *GetItemVariantByBarcodeHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetItemVariantByBarcodeQuery)

	variant, err := h.variantRepo.GetByBarcode(ctx, q.Barcode)
	if err != nil {
		return nil, err
	}

	if variant == nil {
		return nil, errors.New("item variant not found")
	}

	return variant, nil
}

type ListItemVariantsHandler struct {
	variantRepo interfaces.ItemVariantRepository
}

func NewListItemVariantsHandler(variantRepo interfaces.ItemVariantRepository) *ListItemVariantsHandler {
	return &ListItemVariantsHandler{
		variantRepo: variantRepo,
	}
}

func (h *ListItemVariantsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*ListItemVariantsQuery)

	page := q.Page
	pageSize := q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	variants, total, err := h.variantRepo.List(ctx, offset, pageSize, q.ItemID, q.SizeID, q.IsNormallyFrozen, q.IncludeDeleted)
	if err != nil {
		return nil, err
	}

	return &ListItemVariantsResult{
		ItemVariants: variants,
		Total:        total,
	}, nil
}

type GetItemVariantsByItemHandler struct {
	variantRepo interfaces.ItemVariantRepository
	itemRepo    interfaces.ItemRepository
}

func NewGetItemVariantsByItemHandler(
	variantRepo interfaces.ItemVariantRepository,
	itemRepo interfaces.ItemRepository,
) *GetItemVariantsByItemHandler {
	return &GetItemVariantsByItemHandler{
		variantRepo: variantRepo,
		itemRepo:    itemRepo,
	}
}

func (h *GetItemVariantsByItemHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetItemVariantsByItemQuery)

	// Validate item exists
	item, err := h.itemRepo.GetByID(ctx, q.ItemID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return errors.New("item not found"), nil
	}

	variants, err := h.variantRepo.GetByItem(ctx, q.ItemID)
	if err != nil {
		return nil, err
	}

	return variants, nil
}

type GetLowStockVariantsHandler struct {
	variantRepo interfaces.ItemVariantRepository
}

func NewGetLowStockVariantsHandler(variantRepo interfaces.ItemVariantRepository) *GetLowStockVariantsHandler {
	return &GetLowStockVariantsHandler{
		variantRepo: variantRepo,
	}
}

func (h *GetLowStockVariantsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetLowStockVariantsQuery)

	page := q.Page
	pageSize := q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	variants, total, err := h.variantRepo.GetLowStockVariants(ctx, offset, pageSize)
	if err != nil {
		return nil, err
	}

	return &ListItemVariantsResult{
		ItemVariants: variants,
		Total:        total,
	}, nil
}

type GetOverstockedVariantsHandler struct {
	variantRepo interfaces.ItemVariantRepository
}

func NewGetOverstockedVariantsHandler(variantRepo interfaces.ItemVariantRepository) *GetOverstockedVariantsHandler {
	return &GetOverstockedVariantsHandler{
		variantRepo: variantRepo,
	}
}

func (h *GetOverstockedVariantsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetOverstockedVariantsQuery)

	page := q.Page
	pageSize := q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	variants, total, err := h.variantRepo.GetOverstockedVariants(ctx, offset, pageSize)
	if err != nil {
		return nil, err
	}

	return &ListItemVariantsResult{
		ItemVariants: variants,
		Total:        total,
	}, nil
}

type GetCurrentStockHandler struct {
	variantRepo interfaces.ItemVariantRepository
}

func NewGetCurrentStockHandler(variantRepo interfaces.ItemVariantRepository) *GetCurrentStockHandler {
	return &GetCurrentStockHandler{
		variantRepo: variantRepo,
	}
}

func (h *GetCurrentStockHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetCurrentStockQuery)

	// Get variant to check thresholds
	variant, err := h.variantRepo.GetByID(ctx, q.VariantID)
	if err != nil {
		return nil, err
	}
	if variant == nil {
		return nil, errors.New("item variant not found")
	}

	// Get current stock
	stock, err := h.variantRepo.GetCurrentStock(ctx, q.VariantID)
	if err != nil {
		return nil, err
	}

	// Check if needs restock
	needsRestock := stock < variant.MinSKU

	// Check if overstocked
	isOverstocked := variant.MaxSKU > 0 && stock > variant.MaxSKU

	return &CurrentStockResult{
		Stock:         stock,
		NeedsRestock:  needsRestock,
		IsOverstocked: isOverstocked,
	}, nil
}
