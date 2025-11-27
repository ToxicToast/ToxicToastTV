package command

import (
	"context"
	"errors"
	"time"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/interfaces"
)

type CreateMessageCommand struct {
	cqrs.BaseCommand
	StreamID       string `json:"stream_id"`
	UserID         string `json:"user_id"`
	Username       string `json:"username"`
	DisplayName    string `json:"display_name"`
	Message        string `json:"message"`
	IsModerator    bool   `json:"is_moderator"`
	IsSubscriber   bool   `json:"is_subscriber"`
	IsVIP          bool   `json:"is_vip"`
	IsBroadcaster  bool   `json:"is_broadcaster"`
}

func (c *CreateMessageCommand) CommandName() string { return "create_message" }
func (c *CreateMessageCommand) Validate() error {
	if c.StreamID == "" || c.UserID == "" || c.Username == "" || c.Message == "" {
		return errors.New("invalid message data")
	}
	return nil
}

type DeleteMessageCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteMessageCommand) CommandName() string { return "delete_message" }
func (c *DeleteMessageCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("message ID is required")
	}
	return nil
}

// Handlers

type CreateMessageHandler struct {
	messageRepo interfaces.MessageRepository
	streamRepo  interfaces.StreamRepository
	viewerRepo  interfaces.ViewerRepository
}

func NewCreateMessageHandler(
	messageRepo interfaces.MessageRepository,
	streamRepo interfaces.StreamRepository,
	viewerRepo interfaces.ViewerRepository,
) *CreateMessageHandler {
	return &CreateMessageHandler{
		messageRepo: messageRepo,
		streamRepo:  streamRepo,
		viewerRepo:  viewerRepo,
	}
}

func (h *CreateMessageHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	createCmd := cmd.(*CreateMessageCommand)

	// Verify stream exists
	stream, err := h.streamRepo.GetByID(ctx, createCmd.StreamID)
	if err != nil {
		return err
	}
	if stream == nil {
		return errors.New("stream not found")
	}

	// Create or update viewer
	viewer, err := h.viewerRepo.GetByTwitchID(ctx, createCmd.UserID)
	if err != nil {
		return err
	}

	if viewer == nil {
		// Create new viewer
		viewer = &domain.Viewer{
			TwitchID:            createCmd.UserID,
			Username:            createCmd.Username,
			DisplayName:         createCmd.DisplayName,
			TotalMessages:       1,
			TotalStreamsWatched: 1,
			FirstSeen:           time.Now(),
			LastSeen:            time.Now(),
		}
		if err := h.viewerRepo.Create(ctx, viewer); err != nil {
			return err
		}
	} else {
		// Update existing viewer
		viewer.TotalMessages++
		viewer.LastSeen = time.Now()
		viewer.Username = createCmd.Username
		viewer.DisplayName = createCmd.DisplayName
		if err := h.viewerRepo.Update(ctx, viewer); err != nil {
			return err
		}
	}

	// Create message
	msg := &domain.Message{
		StreamID:      createCmd.StreamID,
		UserID:        createCmd.UserID,
		Username:      createCmd.Username,
		DisplayName:   createCmd.DisplayName,
		Message:       createCmd.Message,
		IsModerator:   createCmd.IsModerator,
		IsSubscriber:  createCmd.IsSubscriber,
		IsVIP:         createCmd.IsVIP,
		IsBroadcaster: createCmd.IsBroadcaster,
		SentAt:        time.Now(),
	}

	if err := h.messageRepo.Create(ctx, msg); err != nil {
		return err
	}

	// Update stream message count
	stream.TotalMessages++
	if err := h.streamRepo.Update(ctx, stream); err != nil {
		return err
	}

	createCmd.AggregateID = msg.ID
	return nil
}

type DeleteMessageHandler struct {
	messageRepo interfaces.MessageRepository
}

func NewDeleteMessageHandler(messageRepo interfaces.MessageRepository) *DeleteMessageHandler {
	return &DeleteMessageHandler{messageRepo: messageRepo}
}

func (h *DeleteMessageHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteMessageCommand)

	message, err := h.messageRepo.GetByID(ctx, deleteCmd.AggregateID)
	if err != nil {
		return err
	}
	if message == nil {
		return errors.New("message not found")
	}

	return h.messageRepo.Delete(ctx, deleteCmd.AggregateID)
}
