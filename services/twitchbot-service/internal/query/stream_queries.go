package query

import (
	"context"
	"errors"
	"time"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/interfaces"
)

type GetStreamByIDQuery struct {
	cqrs.BaseQuery
	ID string `json:"id"`
}

func (q *GetStreamByIDQuery) QueryName() string { return "get_stream_by_id" }
func (q *GetStreamByIDQuery) Validate() error {
	if q.ID == "" {
		return errors.New("stream ID is required")
	}
	return nil
}

type ListStreamsQuery struct {
	cqrs.BaseQuery
	Page           int    `json:"page"`
	PageSize       int    `json:"page_size"`
	OnlyActive     bool   `json:"only_active"`
	GameName       string `json:"game_name"`
	IncludeDeleted bool   `json:"include_deleted"`
}

func (q *ListStreamsQuery) QueryName() string { return "list_streams" }
func (q *ListStreamsQuery) Validate() error {
	if q.PageSize <= 0 || q.PageSize > 100 {
		q.PageSize = 20
	}
	return nil
}

type GetActiveStreamQuery struct {
	cqrs.BaseQuery
}

func (q *GetActiveStreamQuery) QueryName() string { return "get_active_stream" }
func (q *GetActiveStreamQuery) Validate() error  { return nil }

type GetStreamStatsQuery struct {
	cqrs.BaseQuery
	ID string `json:"id"`
}

func (q *GetStreamStatsQuery) QueryName() string { return "get_stream_stats" }
func (q *GetStreamStatsQuery) Validate() error {
	if q.ID == "" {
		return errors.New("stream ID is required")
	}
	return nil
}

// Results

type GetStreamResult struct {
	Stream *domain.Stream
}

type ListStreamsResult struct {
	Streams []*domain.Stream
	Total   int64
}

type GetActiveStreamResult struct {
	Stream *domain.Stream
}

type GetStreamStatsResult struct {
	PeakViewers     int
	AverageViewers  int
	TotalMessages   int
	UniqueViewers   int64
	DurationSeconds int64
}

// Handlers

type GetStreamByIDHandler struct {
	streamRepo interfaces.StreamRepository
}

func NewGetStreamByIDHandler(streamRepo interfaces.StreamRepository) *GetStreamByIDHandler {
	return &GetStreamByIDHandler{streamRepo: streamRepo}
}

func (h *GetStreamByIDHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	qry := query.(*GetStreamByIDQuery)

	stream, err := h.streamRepo.GetByID(ctx, qry.ID)
	if err != nil {
		return nil, err
	}
	if stream == nil {
		return nil, errors.New("stream not found")
	}

	return &GetStreamResult{Stream: stream}, nil
}

type ListStreamsHandler struct {
	streamRepo interfaces.StreamRepository
}

func NewListStreamsHandler(streamRepo interfaces.StreamRepository) *ListStreamsHandler {
	return &ListStreamsHandler{streamRepo: streamRepo}
}

func (h *ListStreamsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	qry := query.(*ListStreamsQuery)

	offset := (qry.Page - 1) * qry.PageSize

	streams, total, err := h.streamRepo.List(ctx, offset, qry.PageSize, qry.OnlyActive, qry.GameName, qry.IncludeDeleted)
	if err != nil {
		return nil, err
	}

	return &ListStreamsResult{
		Streams: streams,
		Total:   total,
	}, nil
}

type GetActiveStreamHandler struct {
	streamRepo interfaces.StreamRepository
}

func NewGetActiveStreamHandler(streamRepo interfaces.StreamRepository) *GetActiveStreamHandler {
	return &GetActiveStreamHandler{streamRepo: streamRepo}
}

func (h *GetActiveStreamHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	stream, err := h.streamRepo.GetActive(ctx)
	if err != nil {
		return nil, err
	}
	if stream == nil {
		return nil, errors.New("no active stream found")
	}

	return &GetActiveStreamResult{Stream: stream}, nil
}

type GetStreamStatsHandler struct {
	streamRepo  interfaces.StreamRepository
	messageRepo interfaces.MessageRepository
}

func NewGetStreamStatsHandler(streamRepo interfaces.StreamRepository, messageRepo interfaces.MessageRepository) *GetStreamStatsHandler {
	return &GetStreamStatsHandler{
		streamRepo:  streamRepo,
		messageRepo: messageRepo,
	}
}

func (h *GetStreamStatsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	qry := query.(*GetStreamStatsQuery)

	stream, err := h.streamRepo.GetByID(ctx, qry.ID)
	if err != nil {
		return nil, err
	}
	if stream == nil {
		return nil, errors.New("stream not found")
	}

	// Get message stats
	totalMsg, uniqueUsers, _, _, err := h.messageRepo.GetStats(ctx, qry.ID)
	if err != nil {
		return nil, err
	}

	// Calculate duration
	var duration int64
	if stream.EndedAt != nil {
		duration = int64(stream.EndedAt.Sub(stream.StartedAt).Seconds())
	} else {
		duration = int64(time.Since(stream.StartedAt).Seconds())
	}

	return &GetStreamStatsResult{
		PeakViewers:     stream.PeakViewers,
		AverageViewers:  stream.AverageViewers,
		TotalMessages:   int(totalMsg),
		UniqueViewers:   uniqueUsers,
		DurationSeconds: duration,
	}, nil
}
