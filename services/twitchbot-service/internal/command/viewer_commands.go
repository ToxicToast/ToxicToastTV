package command

import (
	"context"
	"errors"
	"time"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/interfaces"
)

type CreateViewerCommand struct {
	cqrs.BaseCommand
	TwitchID    string `json:"twitch_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

func (c *CreateViewerCommand) CommandName() string { return "create_viewer" }
func (c *CreateViewerCommand) Validate() error {
	if c.TwitchID == "" || c.Username == "" {
		return errors.New("invalid viewer data")
	}
	return nil
}

type UpdateViewerCommand struct {
	cqrs.BaseCommand
	Username            *string `json:"username"`
	DisplayName         *string `json:"display_name"`
	TotalMessages       *int    `json:"total_messages"`
	TotalStreamsWatched *int    `json:"total_streams_watched"`
}

func (c *UpdateViewerCommand) CommandName() string { return "update_viewer" }
func (c *UpdateViewerCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("viewer ID is required")
	}
	return nil
}

type DeleteViewerCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteViewerCommand) CommandName() string { return "delete_viewer" }
func (c *DeleteViewerCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("viewer ID is required")
	}
	return nil
}

// Handlers

type CreateViewerHandler struct {
	viewerRepo interfaces.ViewerRepository
}

func NewCreateViewerHandler(viewerRepo interfaces.ViewerRepository) *CreateViewerHandler {
	return &CreateViewerHandler{viewerRepo: viewerRepo}
}

func (h *CreateViewerHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	createCmd := cmd.(*CreateViewerCommand)

	viewer := &domain.Viewer{
		TwitchID:            createCmd.TwitchID,
		Username:            createCmd.Username,
		DisplayName:         createCmd.DisplayName,
		TotalMessages:       0,
		TotalStreamsWatched: 0,
		FirstSeen:           time.Now(),
		LastSeen:            time.Now(),
	}

	if err := h.viewerRepo.Create(ctx, viewer); err != nil {
		return err
	}

	createCmd.AggregateID = viewer.ID
	return nil
}

type UpdateViewerHandler struct {
	viewerRepo interfaces.ViewerRepository
}

func NewUpdateViewerHandler(viewerRepo interfaces.ViewerRepository) *UpdateViewerHandler {
	return &UpdateViewerHandler{viewerRepo: viewerRepo}
}

func (h *UpdateViewerHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateViewerCommand)

	viewer, err := h.viewerRepo.GetByID(ctx, updateCmd.AggregateID)
	if err != nil {
		return err
	}
	if viewer == nil {
		return errors.New("viewer not found")
	}

	if updateCmd.Username != nil {
		viewer.Username = *updateCmd.Username
	}
	if updateCmd.DisplayName != nil {
		viewer.DisplayName = *updateCmd.DisplayName
	}
	if updateCmd.TotalMessages != nil {
		viewer.TotalMessages = *updateCmd.TotalMessages
	}
	if updateCmd.TotalStreamsWatched != nil {
		viewer.TotalStreamsWatched = *updateCmd.TotalStreamsWatched
	}

	return h.viewerRepo.Update(ctx, viewer)
}

type DeleteViewerHandler struct {
	viewerRepo interfaces.ViewerRepository
}

func NewDeleteViewerHandler(viewerRepo interfaces.ViewerRepository) *DeleteViewerHandler {
	return &DeleteViewerHandler{viewerRepo: viewerRepo}
}

func (h *DeleteViewerHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteViewerCommand)

	viewer, err := h.viewerRepo.GetByID(ctx, deleteCmd.AggregateID)
	if err != nil {
		return err
	}
	if viewer == nil {
		return errors.New("viewer not found")
	}

	return h.viewerRepo.Delete(ctx, deleteCmd.AggregateID)
}
