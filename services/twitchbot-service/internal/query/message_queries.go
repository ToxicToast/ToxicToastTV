package query

import (
	"context"
	"errors"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/interfaces"
)

type GetMessageByIDQuery struct {
	cqrs.BaseQuery
	ID string `json:"id"`
}

func (q *GetMessageByIDQuery) QueryName() string { return "get_message_by_id" }
func (q *GetMessageByIDQuery) Validate() error {
	if q.ID == "" {
		return errors.New("message ID is required")
	}
	return nil
}

type ListMessagesQuery struct {
	cqrs.BaseQuery
	Page           int    `json:"page"`
	PageSize       int    `json:"page_size"`
	StreamID       string `json:"stream_id"`
	UserID         string `json:"user_id"`
	IncludeDeleted bool   `json:"include_deleted"`
}

func (q *ListMessagesQuery) QueryName() string { return "list_messages" }
func (q *ListMessagesQuery) Validate() error {
	if q.PageSize <= 0 || q.PageSize > 100 {
		q.PageSize = 20
	}
	return nil
}

type SearchMessagesQuery struct {
	cqrs.BaseQuery
	Query    string `json:"query"`
	StreamID string `json:"stream_id"`
	UserID   string `json:"user_id"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

func (q *SearchMessagesQuery) QueryName() string { return "search_messages" }
func (q *SearchMessagesQuery) Validate() error {
	if q.PageSize <= 0 || q.PageSize > 100 {
		q.PageSize = 20
	}
	return nil
}

type GetMessageStatsQuery struct {
	cqrs.BaseQuery
	StreamID string `json:"stream_id"`
}

func (q *GetMessageStatsQuery) QueryName() string { return "get_message_stats" }
func (q *GetMessageStatsQuery) Validate() error {
	if q.StreamID == "" {
		return errors.New("stream ID is required")
	}
	return nil
}

// Results

type GetMessageResult struct {
	Message *domain.Message
}

type ListMessagesResult struct {
	Messages []*domain.Message
	Total    int64
}

type SearchMessagesResult struct {
	Messages []*domain.Message
	Total    int64
}

type GetMessageStatsResult struct {
	TotalMessages       int64
	UniqueUsers         int64
	MostActiveUser      string
	MostActiveUserCount int64
}

// Handlers

type GetMessageByIDHandler struct {
	messageRepo interfaces.MessageRepository
}

func NewGetMessageByIDHandler(messageRepo interfaces.MessageRepository) *GetMessageByIDHandler {
	return &GetMessageByIDHandler{messageRepo: messageRepo}
}

func (h *GetMessageByIDHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	qry := query.(*GetMessageByIDQuery)

	message, err := h.messageRepo.GetByID(ctx, qry.ID)
	if err != nil {
		return nil, err
	}
	if message == nil {
		return nil, errors.New("message not found")
	}

	return &GetMessageResult{Message: message}, nil
}

type ListMessagesHandler struct {
	messageRepo interfaces.MessageRepository
}

func NewListMessagesHandler(messageRepo interfaces.MessageRepository) *ListMessagesHandler {
	return &ListMessagesHandler{messageRepo: messageRepo}
}

func (h *ListMessagesHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	qry := query.(*ListMessagesQuery)

	offset := (qry.Page - 1) * qry.PageSize

	messages, total, err := h.messageRepo.List(ctx, offset, qry.PageSize, qry.StreamID, qry.UserID, qry.IncludeDeleted)
	if err != nil {
		return nil, err
	}

	return &ListMessagesResult{
		Messages: messages,
		Total:    total,
	}, nil
}

type SearchMessagesHandler struct {
	messageRepo interfaces.MessageRepository
}

func NewSearchMessagesHandler(messageRepo interfaces.MessageRepository) *SearchMessagesHandler {
	return &SearchMessagesHandler{messageRepo: messageRepo}
}

func (h *SearchMessagesHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	qry := query.(*SearchMessagesQuery)

	offset := (qry.Page - 1) * qry.PageSize

	messages, total, err := h.messageRepo.Search(ctx, qry.Query, qry.StreamID, qry.UserID, offset, qry.PageSize)
	if err != nil {
		return nil, err
	}

	return &SearchMessagesResult{
		Messages: messages,
		Total:    total,
	}, nil
}

type GetMessageStatsHandler struct {
	messageRepo interfaces.MessageRepository
}

func NewGetMessageStatsHandler(messageRepo interfaces.MessageRepository) *GetMessageStatsHandler {
	return &GetMessageStatsHandler{messageRepo: messageRepo}
}

func (h *GetMessageStatsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	qry := query.(*GetMessageStatsQuery)

	totalMessages, uniqueUsers, mostActiveUser, mostActiveUserCount, err := h.messageRepo.GetStats(ctx, qry.StreamID)
	if err != nil {
		return nil, err
	}

	return &GetMessageStatsResult{
		TotalMessages:       totalMessages,
		UniqueUsers:         uniqueUsers,
		MostActiveUser:      mostActiveUser,
		MostActiveUserCount: mostActiveUserCount,
	}, nil
}
