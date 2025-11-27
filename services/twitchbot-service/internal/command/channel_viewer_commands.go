package command

import (
	"context"
	"time"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/interfaces"
)

type AddViewerCommand struct {
	cqrs.BaseCommand
	Channel     string `json:"channel"`
	TwitchID    string `json:"twitch_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	IsModerator bool   `json:"is_moderator"`
	IsVIP       bool   `json:"is_vip"`
}

func (c *AddViewerCommand) CommandName() string { return "add_viewer" }
func (c *AddViewerCommand) Validate() error    { return nil }

type UpdateLastSeenCommand struct {
	cqrs.BaseCommand
	Channel  string `json:"channel"`
	TwitchID string `json:"twitch_id"`
}

func (c *UpdateLastSeenCommand) CommandName() string { return "update_last_seen" }
func (c *UpdateLastSeenCommand) Validate() error    { return nil }

type RemoveViewerCommand struct {
	cqrs.BaseCommand
	Channel  string `json:"channel"`
	TwitchID string `json:"twitch_id"`
}

func (c *RemoveViewerCommand) CommandName() string { return "remove_viewer" }
func (c *RemoveViewerCommand) Validate() error    { return nil }

// Handlers

type AddViewerHandler struct {
	channelViewerRepo interfaces.ChannelViewerRepository
	viewerRepo        interfaces.ViewerRepository
}

func NewAddViewerHandler(
	channelViewerRepo interfaces.ChannelViewerRepository,
	viewerRepo interfaces.ViewerRepository,
) *AddViewerHandler {
	return &AddViewerHandler{
		channelViewerRepo: channelViewerRepo,
		viewerRepo:        viewerRepo,
	}
}

func (h *AddViewerHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	addCmd := cmd.(*AddViewerCommand)
	now := time.Now()

	// Check if viewer already exists in this channel
	existing, err := h.channelViewerRepo.GetByChannelAndTwitchID(ctx, addCmd.Channel, addCmd.TwitchID)
	if err != nil {
		return err
	}

	if existing != nil {
		// Update existing viewer
		existing.Username = addCmd.Username
		existing.DisplayName = addCmd.DisplayName
		existing.LastSeen = now
		existing.IsModerator = addCmd.IsModerator
		existing.IsVIP = addCmd.IsVIP

		return h.channelViewerRepo.Upsert(ctx, existing)
	}

	// Create new channel viewer
	channelViewer := &domain.ChannelViewer{
		Channel:     addCmd.Channel,
		TwitchID:    addCmd.TwitchID,
		Username:    addCmd.Username,
		DisplayName: addCmd.DisplayName,
		FirstSeen:   now,
		LastSeen:    now,
		IsModerator: addCmd.IsModerator,
		IsVIP:       addCmd.IsVIP,
	}

	if err := h.channelViewerRepo.Upsert(ctx, channelViewer); err != nil {
		return err
	}

	// Also create/update global viewer record
	globalViewer, err := h.viewerRepo.GetByTwitchID(ctx, addCmd.TwitchID)
	if err != nil {
		return nil // Ignore global viewer errors
	}

	if globalViewer == nil {
		// Create new global viewer
		globalViewer = &domain.Viewer{
			TwitchID:    addCmd.TwitchID,
			Username:    addCmd.Username,
			DisplayName: addCmd.DisplayName,
			FirstSeen:   now,
			LastSeen:    now,
		}
		h.viewerRepo.Create(ctx, globalViewer)
	} else {
		// Update global viewer's last seen
		globalViewer.Username = addCmd.Username
		globalViewer.DisplayName = addCmd.DisplayName
		globalViewer.LastSeen = now
		h.viewerRepo.Update(ctx, globalViewer)
	}

	return nil
}

type UpdateLastSeenHandler struct {
	channelViewerRepo interfaces.ChannelViewerRepository
}

func NewUpdateLastSeenHandler(channelViewerRepo interfaces.ChannelViewerRepository) *UpdateLastSeenHandler {
	return &UpdateLastSeenHandler{channelViewerRepo: channelViewerRepo}
}

func (h *UpdateLastSeenHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateLastSeenCommand)
	return h.channelViewerRepo.UpdateLastSeen(ctx, updateCmd.Channel, updateCmd.TwitchID)
}

type RemoveViewerHandler struct {
	channelViewerRepo interfaces.ChannelViewerRepository
}

func NewRemoveViewerHandler(channelViewerRepo interfaces.ChannelViewerRepository) *RemoveViewerHandler {
	return &RemoveViewerHandler{channelViewerRepo: channelViewerRepo}
}

func (h *RemoveViewerHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	removeCmd := cmd.(*RemoveViewerCommand)
	return h.channelViewerRepo.Delete(ctx, removeCmd.Channel, removeCmd.TwitchID)
}
