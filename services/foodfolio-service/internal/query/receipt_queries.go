package query

import (
	"context"
	"errors"
	"time"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"

	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/interfaces"
)

// ============================================================================
// Queries
// ============================================================================

type GetReceiptByIDQuery struct {
	cqrs.BaseQuery
	ID string `json:"id"`
}

func (q *GetReceiptByIDQuery) QueryName() string {
	return "get_receipt_by_id"
}

func (q *GetReceiptByIDQuery) Validate() error {
	if q.ID == "" {
		return errors.New("id is required")
	}
	return nil
}

type ListReceiptsQuery struct {
	cqrs.BaseQuery
	Page           int        `json:"page"`
	PageSize       int        `json:"page_size"`
	WarehouseID    *string    `json:"warehouse_id"`
	StartDate      *time.Time `json:"start_date"`
	EndDate        *time.Time `json:"end_date"`
	IncludeDeleted bool       `json:"include_deleted"`
}

func (q *ListReceiptsQuery) QueryName() string {
	return "list_receipts"
}

func (q *ListReceiptsQuery) Validate() error {
	return nil
}

type GetUnmatchedItemsQuery struct {
	cqrs.BaseQuery
	ReceiptID string `json:"receipt_id"`
}

func (q *GetUnmatchedItemsQuery) QueryName() string {
	return "get_unmatched_items"
}

func (q *GetUnmatchedItemsQuery) Validate() error {
	if q.ReceiptID == "" {
		return errors.New("receipt_id is required")
	}
	return nil
}

type GetStatisticsQuery struct {
	cqrs.BaseQuery
	WarehouseID *string    `json:"warehouse_id"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
}

func (q *GetStatisticsQuery) QueryName() string {
	return "get_statistics"
}

func (q *GetStatisticsQuery) Validate() error {
	return nil
}

// ============================================================================
// Query Results
// ============================================================================

type ListReceiptsResult struct {
	Receipts []*domain.Receipt
	Total    int64
}

// ============================================================================
// Query Handlers
// ============================================================================

type GetReceiptByIDHandler struct {
	receiptRepo interfaces.ReceiptRepository
}

func NewGetReceiptByIDHandler(receiptRepo interfaces.ReceiptRepository) *GetReceiptByIDHandler {
	return &GetReceiptByIDHandler{
		receiptRepo: receiptRepo,
	}
}

func (h *GetReceiptByIDHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetReceiptByIDQuery)

	receipt, err := h.receiptRepo.GetByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	if receipt == nil {
		return nil, errors.New("receipt not found")
	}

	return receipt, nil
}

type ListReceiptsHandler struct {
	receiptRepo interfaces.ReceiptRepository
}

func NewListReceiptsHandler(receiptRepo interfaces.ReceiptRepository) *ListReceiptsHandler {
	return &ListReceiptsHandler{
		receiptRepo: receiptRepo,
	}
}

func (h *ListReceiptsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*ListReceiptsQuery)

	page := q.Page
	pageSize := q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	receipts, total, err := h.receiptRepo.List(ctx, offset, pageSize, q.WarehouseID, q.StartDate, q.EndDate, q.IncludeDeleted)
	if err != nil {
		return nil, err
	}

	return &ListReceiptsResult{
		Receipts: receipts,
		Total:    total,
	}, nil
}

type GetUnmatchedItemsHandler struct {
	receiptRepo interfaces.ReceiptRepository
}

func NewGetUnmatchedItemsHandler(receiptRepo interfaces.ReceiptRepository) *GetUnmatchedItemsHandler {
	return &GetUnmatchedItemsHandler{
		receiptRepo: receiptRepo,
	}
}

func (h *GetUnmatchedItemsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetUnmatchedItemsQuery)

	// Validate receipt exists
	receipt, err := h.receiptRepo.GetByID(ctx, q.ReceiptID)
	if err != nil {
		return nil, err
	}
	if receipt == nil {
		return nil, errors.New("receipt not found")
	}

	items, err := h.receiptRepo.GetUnmatchedItems(ctx, q.ReceiptID)
	if err != nil {
		return nil, err
	}

	return items, nil
}

type GetStatisticsHandler struct {
	receiptRepo interfaces.ReceiptRepository
}

func NewGetStatisticsHandler(receiptRepo interfaces.ReceiptRepository) *GetStatisticsHandler {
	return &GetStatisticsHandler{
		receiptRepo: receiptRepo,
	}
}

func (h *GetStatisticsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetStatisticsQuery)

	stats, err := h.receiptRepo.GetStatistics(ctx, q.WarehouseID, q.StartDate, q.EndDate)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
