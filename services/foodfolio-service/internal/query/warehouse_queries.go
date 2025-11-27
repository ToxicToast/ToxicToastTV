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

type GetWarehouseByIDQuery struct {
	cqrs.BaseQuery
	ID string `json:"id"`
}

func (q *GetWarehouseByIDQuery) QueryName() string {
	return "get_warehouse_by_id"
}

func (q *GetWarehouseByIDQuery) Validate() error {
	if q.ID == "" {
		return errors.New("id is required")
	}
	return nil
}

type GetWarehouseBySlugQuery struct {
	cqrs.BaseQuery
	Slug string `json:"slug"`
}

func (q *GetWarehouseBySlugQuery) QueryName() string {
	return "get_warehouse_by_slug"
}

func (q *GetWarehouseBySlugQuery) Validate() error {
	if q.Slug == "" {
		return errors.New("slug is required")
	}
	return nil
}

type ListWarehousesQuery struct {
	cqrs.BaseQuery
	Page           int    `json:"page"`
	PageSize       int    `json:"page_size"`
	Search         string `json:"search"`
	IncludeDeleted bool   `json:"include_deleted"`
}

func (q *ListWarehousesQuery) QueryName() string {
	return "list_warehouses"
}

func (q *ListWarehousesQuery) Validate() error {
	return nil
}

// ============================================================================
// Query Results
// ============================================================================

type ListWarehousesResult struct {
	Warehouses []*domain.Warehouse
	Total      int64
}

// ============================================================================
// Query Handlers
// ============================================================================

type GetWarehouseByIDHandler struct {
	warehouseRepo interfaces.WarehouseRepository
}

func NewGetWarehouseByIDHandler(warehouseRepo interfaces.WarehouseRepository) *GetWarehouseByIDHandler {
	return &GetWarehouseByIDHandler{
		warehouseRepo: warehouseRepo,
	}
}

func (h *GetWarehouseByIDHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetWarehouseByIDQuery)

	warehouseEntity, err := h.warehouseRepo.GetByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	if warehouseEntity == nil {
		return nil, errors.New("warehouse not found")
	}

	return warehouseEntity, nil
}

type GetWarehouseBySlugHandler struct {
	warehouseRepo interfaces.WarehouseRepository
}

func NewGetWarehouseBySlugHandler(warehouseRepo interfaces.WarehouseRepository) *GetWarehouseBySlugHandler {
	return &GetWarehouseBySlugHandler{
		warehouseRepo: warehouseRepo,
	}
}

func (h *GetWarehouseBySlugHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetWarehouseBySlugQuery)

	warehouseEntity, err := h.warehouseRepo.GetBySlug(ctx, q.Slug)
	if err != nil {
		return nil, err
	}

	if warehouseEntity == nil {
		return nil, errors.New("warehouse not found")
	}

	return warehouseEntity, nil
}

type ListWarehousesHandler struct {
	warehouseRepo interfaces.WarehouseRepository
}

func NewListWarehousesHandler(warehouseRepo interfaces.WarehouseRepository) *ListWarehousesHandler {
	return &ListWarehousesHandler{
		warehouseRepo: warehouseRepo,
	}
}

func (h *ListWarehousesHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*ListWarehousesQuery)

	page := q.Page
	pageSize := q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	warehouses, total, err := h.warehouseRepo.List(ctx, offset, pageSize, q.Search, q.IncludeDeleted)
	if err != nil {
		return nil, err
	}

	return &ListWarehousesResult{
		Warehouses: warehouses,
		Total:      total,
	}, nil
}
