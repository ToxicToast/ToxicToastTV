package command

import (
	"context"
	"errors"
	"time"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/interfaces"
)

type CreateStreamCommand struct {
	cqrs.BaseCommand
	Title    string `json:"title"`
	GameName string `json:"game_name"`
	GameID   string `json:"game_id"`
}

func (c *CreateStreamCommand) CommandName() string { return "create_stream" }
func (c *CreateStreamCommand) Validate() error {
	if c.Title == "" {
		return errors.New("stream title is required")
	}
	return nil
}

type UpdateStreamCommand struct {
	cqrs.BaseCommand
	Title          *string `json:"title"`
	GameName       *string `json:"game_name"`
	GameID         *string `json:"game_id"`
	PeakViewers    *int    `json:"peak_viewers"`
	AverageViewers *int    `json:"average_viewers"`
}

func (c *UpdateStreamCommand) CommandName() string { return "update_stream" }
func (c *UpdateStreamCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("stream ID is required")
	}
	return nil
}

type EndStreamCommand struct {
	cqrs.BaseCommand
}

func (c *EndStreamCommand) CommandName() string { return "end_stream" }
func (c *EndStreamCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("stream ID is required")
	}
	return nil
}

type DeleteStreamCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteStreamCommand) CommandName() string { return "delete_stream" }
func (c *DeleteStreamCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("stream ID is required")
	}
	return nil
}

// Handlers

type CreateStreamHandler struct {
	streamRepo interfaces.StreamRepository
}

func NewCreateStreamHandler(streamRepo interfaces.StreamRepository) *CreateStreamHandler {
	return &CreateStreamHandler{streamRepo: streamRepo}
}

func (h *CreateStreamHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	createCmd := cmd.(*CreateStreamCommand)

	stream := &domain.Stream{
		Title:          createCmd.Title,
		GameName:       createCmd.GameName,
		GameID:         createCmd.GameID,
		StartedAt:      time.Now(),
		PeakViewers:    0,
		AverageViewers: 0,
		TotalMessages:  0,
		IsActive:       true,
	}

	if err := h.streamRepo.Create(ctx, stream); err != nil {
		return err
	}

	createCmd.AggregateID = stream.ID
	return nil
}

type UpdateStreamHandler struct {
	streamRepo interfaces.StreamRepository
}

func NewUpdateStreamHandler(streamRepo interfaces.StreamRepository) *UpdateStreamHandler {
	return &UpdateStreamHandler{streamRepo: streamRepo}
}

func (h *UpdateStreamHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateStreamCommand)

	stream, err := h.streamRepo.GetByID(ctx, updateCmd.AggregateID)
	if err != nil {
		return err
	}
	if stream == nil {
		return errors.New("stream not found")
	}

	if updateCmd.Title != nil {
		stream.Title = *updateCmd.Title
	}
	if updateCmd.GameName != nil {
		stream.GameName = *updateCmd.GameName
	}
	if updateCmd.GameID != nil {
		stream.GameID = *updateCmd.GameID
	}
	if updateCmd.PeakViewers != nil {
		stream.PeakViewers = *updateCmd.PeakViewers
	}
	if updateCmd.AverageViewers != nil {
		stream.AverageViewers = *updateCmd.AverageViewers
	}

	return h.streamRepo.Update(ctx, stream)
}

type EndStreamHandler struct {
	streamRepo interfaces.StreamRepository
}

func NewEndStreamHandler(streamRepo interfaces.StreamRepository) *EndStreamHandler {
	return &EndStreamHandler{streamRepo: streamRepo}
}

func (h *EndStreamHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	endCmd := cmd.(*EndStreamCommand)

	stream, err := h.streamRepo.GetByID(ctx, endCmd.AggregateID)
	if err != nil {
		return err
	}
	if stream == nil {
		return errors.New("stream not found")
	}
	if !stream.IsActive {
		return errors.New("stream already ended")
	}

	return h.streamRepo.EndStream(ctx, endCmd.AggregateID)
}

type DeleteStreamHandler struct {
	streamRepo interfaces.StreamRepository
}

func NewDeleteStreamHandler(streamRepo interfaces.StreamRepository) *DeleteStreamHandler {
	return &DeleteStreamHandler{streamRepo: streamRepo}
}

func (h *DeleteStreamHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteStreamCommand)

	stream, err := h.streamRepo.GetByID(ctx, deleteCmd.AggregateID)
	if err != nil {
		return err
	}
	if stream == nil {
		return errors.New("stream not found")
	}

	return h.streamRepo.Delete(ctx, deleteCmd.AggregateID)
}
