package query

import (
	"context"
	"errors"
	"fmt"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"

	"toxictoast/services/notification-service/internal/domain"
	"toxictoast/services/notification-service/internal/repository/interfaces"
)

// ============================================================================
// Queries
// ============================================================================

type GetChannelByIDQuery struct {
	cqrs.BaseQuery
	ID string `json:"id"`
}

func (q *GetChannelByIDQuery) QueryName() string {
	return "get_channel_by_id"
}

func (q *GetChannelByIDQuery) Validate() error {
	if q.ID == "" {
		return errors.New("channel ID is required")
	}
	return nil
}

type ListChannelsQuery struct {
	cqrs.BaseQuery
	Limit      int  `json:"limit"`
	Offset     int  `json:"offset"`
	ActiveOnly bool `json:"active_only"`
}

func (q *ListChannelsQuery) QueryName() string {
	return "list_channels"
}

func (q *ListChannelsQuery) Validate() error {
	if q.Limit < 0 {
		return errors.New("limit must be non-negative")
	}
	if q.Offset < 0 {
		return errors.New("offset must be non-negative")
	}
	return nil
}

type ListChannelsResult struct {
	Channels []*domain.DiscordChannel
	Total    int64
}

type GetActiveChannelsForEventQuery struct {
	cqrs.BaseQuery
	EventType string `json:"event_type"`
}

func (q *GetActiveChannelsForEventQuery) QueryName() string {
	return "get_active_channels_for_event"
}

func (q *GetActiveChannelsForEventQuery) Validate() error {
	if q.EventType == "" {
		return errors.New("event type is required")
	}
	return nil
}

// ============================================================================
// Query Handlers
// ============================================================================

type GetChannelByIDHandler struct {
	channelRepo interfaces.DiscordChannelRepository
}

func NewGetChannelByIDHandler(channelRepo interfaces.DiscordChannelRepository) *GetChannelByIDHandler {
	return &GetChannelByIDHandler{
		channelRepo: channelRepo,
	}
}

func (h *GetChannelByIDHandler) Handle(ctx context.Context, qry cqrs.Query) (interface{}, error) {
	query := qry.(*GetChannelByIDQuery)

	channel, err := h.channelRepo.GetByID(ctx, query.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}
	return channel, nil
}

type ListChannelsHandler struct {
	channelRepo interfaces.DiscordChannelRepository
}

func NewListChannelsHandler(channelRepo interfaces.DiscordChannelRepository) *ListChannelsHandler {
	return &ListChannelsHandler{
		channelRepo: channelRepo,
	}
}

func (h *ListChannelsHandler) Handle(ctx context.Context, qry cqrs.Query) (interface{}, error) {
	query := qry.(*ListChannelsQuery)

	channels, total, err := h.channelRepo.List(ctx, query.Limit, query.Offset, query.ActiveOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to list channels: %w", err)
	}

	return &ListChannelsResult{
		Channels: channels,
		Total:    total,
	}, nil
}

type GetActiveChannelsForEventHandler struct {
	channelRepo interfaces.DiscordChannelRepository
}

func NewGetActiveChannelsForEventHandler(channelRepo interfaces.DiscordChannelRepository) *GetActiveChannelsForEventHandler {
	return &GetActiveChannelsForEventHandler{
		channelRepo: channelRepo,
	}
}

func (h *GetActiveChannelsForEventHandler) Handle(ctx context.Context, qry cqrs.Query) (interface{}, error) {
	query := qry.(*GetActiveChannelsForEventQuery)

	channels, err := h.channelRepo.GetActiveChannelsForEvent(ctx, query.EventType)
	if err != nil {
		return nil, fmt.Errorf("failed to get channels for event: %w", err)
	}
	return channels, nil
}
