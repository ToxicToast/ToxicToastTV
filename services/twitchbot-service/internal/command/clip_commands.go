package command

import (
	"context"
	"errors"
	"time"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/interfaces"
)

type CreateClipCommand struct {
	cqrs.BaseCommand
	StreamID        string    `json:"stream_id"`
	TwitchClipID    string    `json:"twitch_clip_id"`
	Title           string    `json:"title"`
	URL             string    `json:"url"`
	EmbedURL        string    `json:"embed_url"`
	ThumbnailURL    string    `json:"thumbnail_url"`
	CreatorName     string    `json:"creator_name"`
	CreatorID       string    `json:"creator_id"`
	ViewCount       int       `json:"view_count"`
	DurationSeconds int       `json:"duration_seconds"`
	CreatedAtTwitch time.Time `json:"created_at_twitch"`
}

func (c *CreateClipCommand) CommandName() string { return "create_clip" }
func (c *CreateClipCommand) Validate() error {
	if c.StreamID == "" || c.TwitchClipID == "" || c.Title == "" || c.URL == "" {
		return errors.New("invalid clip data")
	}
	return nil
}

type UpdateClipCommand struct {
	cqrs.BaseCommand
	Title     *string `json:"title"`
	ViewCount *int    `json:"view_count"`
}

func (c *UpdateClipCommand) CommandName() string { return "update_clip" }
func (c *UpdateClipCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("clip ID is required")
	}
	return nil
}

type DeleteClipCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteClipCommand) CommandName() string { return "delete_clip" }
func (c *DeleteClipCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("clip ID is required")
	}
	return nil
}

// Handlers

type CreateClipHandler struct {
	clipRepo   interfaces.ClipRepository
	streamRepo interfaces.StreamRepository
}

func NewCreateClipHandler(clipRepo interfaces.ClipRepository, streamRepo interfaces.StreamRepository) *CreateClipHandler {
	return &CreateClipHandler{
		clipRepo:   clipRepo,
		streamRepo: streamRepo,
	}
}

func (h *CreateClipHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	createCmd := cmd.(*CreateClipCommand)

	// Verify stream exists
	stream, err := h.streamRepo.GetByID(ctx, createCmd.StreamID)
	if err != nil {
		return err
	}
	if stream == nil {
		return errors.New("stream not found")
	}

	clip := &domain.Clip{
		StreamID:        createCmd.StreamID,
		TwitchClipID:    createCmd.TwitchClipID,
		Title:           createCmd.Title,
		URL:             createCmd.URL,
		EmbedURL:        createCmd.EmbedURL,
		ThumbnailURL:    createCmd.ThumbnailURL,
		CreatorName:     createCmd.CreatorName,
		CreatorID:       createCmd.CreatorID,
		ViewCount:       createCmd.ViewCount,
		DurationSeconds: createCmd.DurationSeconds,
		CreatedAtTwitch: createCmd.CreatedAtTwitch,
	}

	if err := h.clipRepo.Create(ctx, clip); err != nil {
		return err
	}

	createCmd.AggregateID = clip.ID
	return nil
}

type UpdateClipHandler struct {
	clipRepo interfaces.ClipRepository
}

func NewUpdateClipHandler(clipRepo interfaces.ClipRepository) *UpdateClipHandler {
	return &UpdateClipHandler{clipRepo: clipRepo}
}

func (h *UpdateClipHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateClipCommand)

	clip, err := h.clipRepo.GetByID(ctx, updateCmd.AggregateID)
	if err != nil {
		return err
	}
	if clip == nil {
		return errors.New("clip not found")
	}

	if updateCmd.Title != nil {
		clip.Title = *updateCmd.Title
	}
	if updateCmd.ViewCount != nil {
		clip.ViewCount = *updateCmd.ViewCount
	}

	return h.clipRepo.Update(ctx, clip)
}

type DeleteClipHandler struct {
	clipRepo interfaces.ClipRepository
}

func NewDeleteClipHandler(clipRepo interfaces.ClipRepository) *DeleteClipHandler {
	return &DeleteClipHandler{clipRepo: clipRepo}
}

func (h *DeleteClipHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteClipCommand)

	clip, err := h.clipRepo.GetByID(ctx, deleteCmd.AggregateID)
	if err != nil {
		return err
	}
	if clip == nil {
		return errors.New("clip not found")
	}

	return h.clipRepo.Delete(ctx, deleteCmd.AggregateID)
}
