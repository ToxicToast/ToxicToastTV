package query

import (
	"context"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/interfaces"
)

type GetChannelViewerQuery struct {
	cqrs.BaseQuery
	Channel  string `json:"channel"`
	TwitchID string `json:"twitch_id"`
}

func (q *GetChannelViewerQuery) QueryName() string { return "get_channel_viewer" }
func (q *GetChannelViewerQuery) Validate() error  { return nil }

type ListChannelViewersQuery struct {
	cqrs.BaseQuery
	Channel string `json:"channel"`
	Limit   int    `json:"limit"`
	Offset  int    `json:"offset"`
}

func (q *ListChannelViewersQuery) QueryName() string { return "list_channel_viewers" }
func (q *ListChannelViewersQuery) Validate() error  { return nil }

type CountChannelViewersQuery struct {
	cqrs.BaseQuery
	Channel string `json:"channel"`
}

func (q *CountChannelViewersQuery) QueryName() string { return "count_channel_viewers" }
func (q *CountChannelViewersQuery) Validate() error  { return nil }

// Results

type GetChannelViewerResult struct {
	ChannelViewer *domain.ChannelViewer
}

type ListChannelViewersResult struct {
	ChannelViewers []*domain.ChannelViewer
	Total          int64
}

type CountChannelViewersResult struct {
	Count int64
}

// Handlers

type GetChannelViewerHandler struct {
	channelViewerRepo interfaces.ChannelViewerRepository
}

func NewGetChannelViewerHandler(channelViewerRepo interfaces.ChannelViewerRepository) *GetChannelViewerHandler {
	return &GetChannelViewerHandler{channelViewerRepo: channelViewerRepo}
}

func (h *GetChannelViewerHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	qry := query.(*GetChannelViewerQuery)

	viewer, err := h.channelViewerRepo.GetByChannelAndTwitchID(ctx, qry.Channel, qry.TwitchID)
	if err != nil {
		return nil, err
	}

	return &GetChannelViewerResult{ChannelViewer: viewer}, nil
}

type ListChannelViewersHandler struct {
	channelViewerRepo interfaces.ChannelViewerRepository
}

func NewListChannelViewersHandler(channelViewerRepo interfaces.ChannelViewerRepository) *ListChannelViewersHandler {
	return &ListChannelViewersHandler{channelViewerRepo: channelViewerRepo}
}

func (h *ListChannelViewersHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	qry := query.(*ListChannelViewersQuery)

	viewers, total, err := h.channelViewerRepo.ListByChannel(ctx, qry.Channel, qry.Limit, qry.Offset)
	if err != nil {
		return nil, err
	}

	return &ListChannelViewersResult{
		ChannelViewers: viewers,
		Total:          total,
	}, nil
}

type CountChannelViewersHandler struct {
	channelViewerRepo interfaces.ChannelViewerRepository
}

func NewCountChannelViewersHandler(channelViewerRepo interfaces.ChannelViewerRepository) *CountChannelViewersHandler {
	return &CountChannelViewersHandler{channelViewerRepo: channelViewerRepo}
}

func (h *CountChannelViewersHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	qry := query.(*CountChannelViewersQuery)

	count, err := h.channelViewerRepo.CountByChannel(ctx, qry.Channel)
	if err != nil {
		return nil, err
	}

	return &CountChannelViewersResult{Count: count}, nil
}
