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

type GetLocationByIDQuery struct {
	cqrs.BaseQuery
	ID string `json:"id"`
}

func (q *GetLocationByIDQuery) QueryName() string {
	return "get_location_by_id"
}

func (q *GetLocationByIDQuery) Validate() error {
	if q.ID == "" {
		return errors.New("id is required")
	}
	return nil
}

type GetLocationBySlugQuery struct {
	cqrs.BaseQuery
	Slug string `json:"slug"`
}

func (q *GetLocationBySlugQuery) QueryName() string {
	return "get_location_by_slug"
}

func (q *GetLocationBySlugQuery) Validate() error {
	if q.Slug == "" {
		return errors.New("slug is required")
	}
	return nil
}

type ListLocationsQuery struct {
	cqrs.BaseQuery
	Page            int     `json:"page"`
	PageSize        int     `json:"page_size"`
	ParentID        *string `json:"parent_id"`
	IncludeChildren bool    `json:"include_children"`
	IncludeDeleted  bool    `json:"include_deleted"`
}

func (q *ListLocationsQuery) QueryName() string {
	return "list_locations"
}

func (q *ListLocationsQuery) Validate() error {
	return nil
}

// ============================================================================
// Query Results
// ============================================================================

type ListLocationsResult struct {
	Locations []*domain.Location
	Total     int64
}

// ============================================================================
// Query Handlers
// ============================================================================

type GetLocationByIDHandler struct {
	locationRepo interfaces.LocationRepository
}

func NewGetLocationByIDHandler(locationRepo interfaces.LocationRepository) *GetLocationByIDHandler {
	return &GetLocationByIDHandler{
		locationRepo: locationRepo,
	}
}

func (h *GetLocationByIDHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetLocationByIDQuery)

	locationEntity, err := h.locationRepo.GetByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	if locationEntity == nil {
		return nil, errors.New("location not found")
	}

	return locationEntity, nil
}

type GetLocationBySlugHandler struct {
	locationRepo interfaces.LocationRepository
}

func NewGetLocationBySlugHandler(locationRepo interfaces.LocationRepository) *GetLocationBySlugHandler {
	return &GetLocationBySlugHandler{
		locationRepo: locationRepo,
	}
}

func (h *GetLocationBySlugHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetLocationBySlugQuery)

	locationEntity, err := h.locationRepo.GetBySlug(ctx, q.Slug)
	if err != nil {
		return nil, err
	}

	if locationEntity == nil {
		return nil, errors.New("location not found")
	}

	return locationEntity, nil
}

type ListLocationsHandler struct {
	locationRepo interfaces.LocationRepository
}

func NewListLocationsHandler(locationRepo interfaces.LocationRepository) *ListLocationsHandler {
	return &ListLocationsHandler{
		locationRepo: locationRepo,
	}
}

func (h *ListLocationsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*ListLocationsQuery)

	page := q.Page
	pageSize := q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	locations, total, err := h.locationRepo.List(ctx, offset, pageSize, q.ParentID, q.IncludeChildren, q.IncludeDeleted)
	if err != nil {
		return nil, err
	}

	return &ListLocationsResult{
		Locations: locations,
		Total:     total,
	}, nil
}

type GetLocationTreeQuery struct {
	cqrs.BaseQuery
	RootID   *string `json:"root_id"`
	MaxDepth int     `json:"max_depth"`
}

func (q *GetLocationTreeQuery) QueryName() string {
	return "get_location_tree"
}

func (q *GetLocationTreeQuery) Validate() error {
	return nil
}

type GetLocationTreeHandler struct {
	locationRepo interfaces.LocationRepository
}

func NewGetLocationTreeHandler(locationRepo interfaces.LocationRepository) *GetLocationTreeHandler {
	return &GetLocationTreeHandler{
		locationRepo: locationRepo,
	}
}

func (h *GetLocationTreeHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetLocationTreeQuery)

	locations, err := h.locationRepo.GetTree(ctx, q.RootID, q.MaxDepth)
	if err != nil {
		return nil, err
	}

	return locations, nil
}
