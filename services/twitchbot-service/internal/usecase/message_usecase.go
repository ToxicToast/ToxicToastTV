package usecase

import (
	"context"
	"errors"
	"time"

	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/interfaces"
)

var (
	ErrMessageNotFound    = errors.New("message not found")
	ErrInvalidMessageData = errors.New("invalid message data")
)

type MessageUseCase interface {
	CreateMessage(ctx context.Context, streamID, userID, username, displayName, message string, isModerator, isSubscriber, isVIP, isBroadcaster bool) (*domain.Message, error)
	GetMessageByID(ctx context.Context, id string) (*domain.Message, error)
	ListMessages(ctx context.Context, page, pageSize int, streamID, userID string, includeDeleted bool) ([]*domain.Message, int64, error)
	SearchMessages(ctx context.Context, query, streamID, userID string, page, pageSize int) ([]*domain.Message, int64, error)
	GetMessageStats(ctx context.Context, streamID string) (totalMessages int64, uniqueUsers int64, mostActiveUser string, mostActiveUserCount int64, err error)
	DeleteMessage(ctx context.Context, id string) error
}

type messageUseCase struct {
	messageRepo interfaces.MessageRepository
	streamRepo  interfaces.StreamRepository
	viewerRepo  interfaces.ViewerRepository
}

func NewMessageUseCase(messageRepo interfaces.MessageRepository, streamRepo interfaces.StreamRepository, viewerRepo interfaces.ViewerRepository) MessageUseCase {
	return &messageUseCase{
		messageRepo: messageRepo,
		streamRepo:  streamRepo,
		viewerRepo:  viewerRepo,
	}
}

func (uc *messageUseCase) CreateMessage(ctx context.Context, streamID, userID, username, displayName, message string, isModerator, isSubscriber, isVIP, isBroadcaster bool) (*domain.Message, error) {
	if streamID == "" || userID == "" || username == "" || message == "" {
		return nil, ErrInvalidMessageData
	}

	// Verify stream exists
	stream, err := uc.streamRepo.GetByID(ctx, streamID)
	if err != nil {
		return nil, err
	}
	if stream == nil {
		return nil, errors.New("stream not found")
	}

	// Create or update viewer
	viewer, err := uc.viewerRepo.GetByTwitchID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if viewer == nil {
		// Create new viewer
		viewer = &domain.Viewer{
			TwitchID:            userID,
			Username:            username,
			DisplayName:         displayName,
			TotalMessages:       1,
			TotalStreamsWatched: 1,
			FirstSeen:           time.Now(),
			LastSeen:            time.Now(),
		}
		if err := uc.viewerRepo.Create(ctx, viewer); err != nil {
			return nil, err
		}
	} else {
		// Update existing viewer
		viewer.TotalMessages++
		viewer.LastSeen = time.Now()
		viewer.Username = username
		viewer.DisplayName = displayName
		if err := uc.viewerRepo.Update(ctx, viewer); err != nil {
			return nil, err
		}
	}

	// Create message
	msg := &domain.Message{
		StreamID:      streamID,
		UserID:        userID,
		Username:      username,
		DisplayName:   displayName,
		Message:       message,
		IsModerator:   isModerator,
		IsSubscriber:  isSubscriber,
		IsVIP:         isVIP,
		IsBroadcaster: isBroadcaster,
		SentAt:        time.Now(),
	}

	if err := uc.messageRepo.Create(ctx, msg); err != nil {
		return nil, err
	}

	// Update stream message count
	stream.TotalMessages++
	if err := uc.streamRepo.Update(ctx, stream); err != nil {
		return nil, err
	}

	return msg, nil
}

func (uc *messageUseCase) GetMessageByID(ctx context.Context, id string) (*domain.Message, error) {
	message, err := uc.messageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if message == nil {
		return nil, ErrMessageNotFound
	}

	return message, nil
}

func (uc *messageUseCase) ListMessages(ctx context.Context, page, pageSize int, streamID, userID string, includeDeleted bool) ([]*domain.Message, int64, error) {
	offset := (page - 1) * pageSize

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return uc.messageRepo.List(ctx, offset, pageSize, streamID, userID, includeDeleted)
}

func (uc *messageUseCase) SearchMessages(ctx context.Context, query, streamID, userID string, page, pageSize int) ([]*domain.Message, int64, error) {
	offset := (page - 1) * pageSize

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return uc.messageRepo.Search(ctx, query, streamID, userID, offset, pageSize)
}

func (uc *messageUseCase) GetMessageStats(ctx context.Context, streamID string) (totalMessages int64, uniqueUsers int64, mostActiveUser string, mostActiveUserCount int64, err error) {
	return uc.messageRepo.GetStats(ctx, streamID)
}

func (uc *messageUseCase) DeleteMessage(ctx context.Context, id string) error {
	message, err := uc.messageRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if message == nil {
		return ErrMessageNotFound
	}

	return uc.messageRepo.Delete(ctx, id)
}
