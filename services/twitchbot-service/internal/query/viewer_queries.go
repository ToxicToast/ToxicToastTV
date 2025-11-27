package query

import (
	"context"
	"errors"
	"time"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/interfaces"
)

type GetViewerByIDQuery struct {
	cqrs.BaseQuery
	ID string `json:"id"`
}

func (q *GetViewerByIDQuery) QueryName() string { return "get_viewer_by_id" }
func (q *GetViewerByIDQuery) Validate() error {
	if q.ID == "" {
		return errors.New("viewer ID is required")
	}
	return nil
}

type GetViewerByTwitchIDQuery struct {
	cqrs.BaseQuery
	TwitchID string `json:"twitch_id"`
}

func (q *GetViewerByTwitchIDQuery) QueryName() string { return "get_viewer_by_twitch_id" }
func (q *GetViewerByTwitchIDQuery) Validate() error {
	if q.TwitchID == "" {
		return errors.New("twitch ID is required")
	}
	return nil
}

type ListViewersQuery struct {
	cqrs.BaseQuery
	Page           int    `json:"page"`
	PageSize       int    `json:"page_size"`
	OrderBy        string `json:"order_by"`
	IncludeDeleted bool   `json:"include_deleted"`
}

func (q *ListViewersQuery) QueryName() string { return "list_viewers" }
func (q *ListViewersQuery) Validate() error {
	if q.PageSize <= 0 || q.PageSize > 100 {
		q.PageSize = 20
	}
	return nil
}

type GetViewerStatsQuery struct {
	cqrs.BaseQuery
	ID string `json:"id"`
}

func (q *GetViewerStatsQuery) QueryName() string { return "get_viewer_stats" }
func (q *GetViewerStatsQuery) Validate() error {
	if q.ID == "" {
		return errors.New("viewer ID is required")
	}
	return nil
}

// Results

type GetViewerResult struct {
	Viewer *domain.Viewer
}

type ListViewersResult struct {
	Viewers []*domain.Viewer
	Total   int64
}

type GetViewerStatsResult struct {
	TotalMessages       int
	TotalStreamsWatched int
	DaysSinceFirstSeen  int
	DaysSinceLastSeen   int
}

// Handlers

type GetViewerByIDHandler struct {
	viewerRepo interfaces.ViewerRepository
}

func NewGetViewerByIDHandler(viewerRepo interfaces.ViewerRepository) *GetViewerByIDHandler {
	return &GetViewerByIDHandler{viewerRepo: viewerRepo}
}

func (h *GetViewerByIDHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	qry := query.(*GetViewerByIDQuery)

	viewer, err := h.viewerRepo.GetByID(ctx, qry.ID)
	if err != nil {
		return nil, err
	}
	if viewer == nil {
		return nil, errors.New("viewer not found")
	}

	return &GetViewerResult{Viewer: viewer}, nil
}

type GetViewerByTwitchIDHandler struct {
	viewerRepo interfaces.ViewerRepository
}

func NewGetViewerByTwitchIDHandler(viewerRepo interfaces.ViewerRepository) *GetViewerByTwitchIDHandler {
	return &GetViewerByTwitchIDHandler{viewerRepo: viewerRepo}
}

func (h *GetViewerByTwitchIDHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	qry := query.(*GetViewerByTwitchIDQuery)

	viewer, err := h.viewerRepo.GetByTwitchID(ctx, qry.TwitchID)
	if err != nil {
		return nil, err
	}
	if viewer == nil {
		return nil, errors.New("viewer not found")
	}

	return &GetViewerResult{Viewer: viewer}, nil
}

type ListViewersHandler struct {
	viewerRepo interfaces.ViewerRepository
}

func NewListViewersHandler(viewerRepo interfaces.ViewerRepository) *ListViewersHandler {
	return &ListViewersHandler{viewerRepo: viewerRepo}
}

func (h *ListViewersHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	qry := query.(*ListViewersQuery)

	offset := (qry.Page - 1) * qry.PageSize

	viewers, total, err := h.viewerRepo.List(ctx, offset, qry.PageSize, qry.OrderBy, qry.IncludeDeleted)
	if err != nil {
		return nil, err
	}

	return &ListViewersResult{
		Viewers: viewers,
		Total:   total,
	}, nil
}

type GetViewerStatsHandler struct {
	viewerRepo interfaces.ViewerRepository
}

func NewGetViewerStatsHandler(viewerRepo interfaces.ViewerRepository) *GetViewerStatsHandler {
	return &GetViewerStatsHandler{viewerRepo: viewerRepo}
}

func (h *GetViewerStatsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	qry := query.(*GetViewerStatsQuery)

	viewer, err := h.viewerRepo.GetByID(ctx, qry.ID)
	if err != nil {
		return nil, err
	}
	if viewer == nil {
		return nil, errors.New("viewer not found")
	}

	daysSinceFirstSeen := int(time.Since(viewer.FirstSeen).Hours() / 24)
	daysSinceLastSeen := int(time.Since(viewer.LastSeen).Hours() / 24)

	return &GetViewerStatsResult{
		TotalMessages:       viewer.TotalMessages,
		TotalStreamsWatched: viewer.TotalStreamsWatched,
		DaysSinceFirstSeen:  daysSinceFirstSeen,
		DaysSinceLastSeen:   daysSinceLastSeen,
	}, nil
}
