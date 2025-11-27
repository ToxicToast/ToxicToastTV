package query

import (
	"context"
	"errors"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/interfaces"
)

type GetClipByIDQuery struct {
	cqrs.BaseQuery
	ID string `json:"id"`
}

func (q *GetClipByIDQuery) QueryName() string { return "get_clip_by_id" }
func (q *GetClipByIDQuery) Validate() error {
	if q.ID == "" {
		return errors.New("clip ID is required")
	}
	return nil
}

type GetClipByTwitchClipIDQuery struct {
	cqrs.BaseQuery
	TwitchClipID string `json:"twitch_clip_id"`
}

func (q *GetClipByTwitchClipIDQuery) QueryName() string { return "get_clip_by_twitch_clip_id" }
func (q *GetClipByTwitchClipIDQuery) Validate() error {
	if q.TwitchClipID == "" {
		return errors.New("twitch clip ID is required")
	}
	return nil
}

type ListClipsQuery struct {
	cqrs.BaseQuery
	Page           int    `json:"page"`
	PageSize       int    `json:"page_size"`
	StreamID       string `json:"stream_id"`
	OrderBy        string `json:"order_by"`
	IncludeDeleted bool   `json:"include_deleted"`
}

func (q *ListClipsQuery) QueryName() string { return "list_clips" }
func (q *ListClipsQuery) Validate() error {
	if q.PageSize <= 0 || q.PageSize > 100 {
		q.PageSize = 20
	}
	return nil
}

// Results

type GetClipResult struct {
	Clip *domain.Clip
}

type ListClipsResult struct {
	Clips []*domain.Clip
	Total int64
}

// Handlers

type GetClipByIDHandler struct {
	clipRepo interfaces.ClipRepository
}

func NewGetClipByIDHandler(clipRepo interfaces.ClipRepository) *GetClipByIDHandler {
	return &GetClipByIDHandler{clipRepo: clipRepo}
}

func (h *GetClipByIDHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	qry := query.(*GetClipByIDQuery)

	clip, err := h.clipRepo.GetByID(ctx, qry.ID)
	if err != nil {
		return nil, err
	}
	if clip == nil {
		return nil, errors.New("clip not found")
	}

	return &GetClipResult{Clip: clip}, nil
}

type GetClipByTwitchClipIDHandler struct {
	clipRepo interfaces.ClipRepository
}

func NewGetClipByTwitchClipIDHandler(clipRepo interfaces.ClipRepository) *GetClipByTwitchClipIDHandler {
	return &GetClipByTwitchClipIDHandler{clipRepo: clipRepo}
}

func (h *GetClipByTwitchClipIDHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	qry := query.(*GetClipByTwitchClipIDQuery)

	clip, err := h.clipRepo.GetByTwitchClipID(ctx, qry.TwitchClipID)
	if err != nil {
		return nil, err
	}
	if clip == nil {
		return nil, errors.New("clip not found")
	}

	return &GetClipResult{Clip: clip}, nil
}

type ListClipsHandler struct {
	clipRepo interfaces.ClipRepository
}

func NewListClipsHandler(clipRepo interfaces.ClipRepository) *ListClipsHandler {
	return &ListClipsHandler{clipRepo: clipRepo}
}

func (h *ListClipsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	qry := query.(*ListClipsQuery)

	offset := (qry.Page - 1) * qry.PageSize

	clips, total, err := h.clipRepo.List(ctx, offset, qry.PageSize, qry.StreamID, qry.OrderBy, qry.IncludeDeleted)
	if err != nil {
		return nil, err
	}

	return &ListClipsResult{
		Clips: clips,
		Total: total,
	}, nil
}
