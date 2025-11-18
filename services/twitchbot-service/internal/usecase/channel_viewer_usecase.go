package usecase

import (
	"context"
	"time"

	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/interfaces"
)

type ChannelViewerUseCase interface {
	// Add or update a viewer in a channel
	AddViewer(ctx context.Context, channel, twitchID, username, displayName string, isMod, isVIP bool) (*domain.ChannelViewer, error)

	// Update last seen for a viewer in a channel
	UpdateLastSeen(ctx context.Context, channel, twitchID string) error

	// Get viewer by channel and Twitch ID
	GetViewer(ctx context.Context, channel, twitchID string) (*domain.ChannelViewer, error)

	// List all viewers in a channel
	ListViewers(ctx context.Context, channel string, limit, offset int) ([]*domain.ChannelViewer, int64, error)

	// Remove viewer from channel
	RemoveViewer(ctx context.Context, channel, twitchID string) error

	// Count viewers in a channel
	CountViewers(ctx context.Context, channel string) (int64, error)
}

type channelViewerUseCase struct {
	channelViewerRepo interfaces.ChannelViewerRepository
	viewerRepo        interfaces.ViewerRepository
}

func NewChannelViewerUseCase(
	channelViewerRepo interfaces.ChannelViewerRepository,
	viewerRepo interfaces.ViewerRepository,
) ChannelViewerUseCase {
	return &channelViewerUseCase{
		channelViewerRepo: channelViewerRepo,
		viewerRepo:        viewerRepo,
	}
}

func (uc *channelViewerUseCase) AddViewer(
	ctx context.Context,
	channel, twitchID, username, displayName string,
	isMod, isVIP bool,
) (*domain.ChannelViewer, error) {
	now := time.Now()

	// Check if viewer already exists in this channel
	existing, err := uc.channelViewerRepo.GetByChannelAndTwitchID(ctx, channel, twitchID)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		// Update existing viewer
		existing.Username = username
		existing.DisplayName = displayName
		existing.LastSeen = now
		existing.IsModerator = isMod
		existing.IsVIP = isVIP

		if err := uc.channelViewerRepo.Upsert(ctx, existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	// Create new channel viewer
	channelViewer := &domain.ChannelViewer{
		Channel:     channel,
		TwitchID:    twitchID,
		Username:    username,
		DisplayName: displayName,
		FirstSeen:   now,
		LastSeen:    now,
		IsModerator: isMod,
		IsVIP:       isVIP,
	}

	if err := uc.channelViewerRepo.Upsert(ctx, channelViewer); err != nil {
		return nil, err
	}

	// Also create/update global viewer record
	globalViewer, err := uc.viewerRepo.GetByTwitchID(ctx, twitchID)
	if err != nil {
		return channelViewer, nil // Ignore global viewer errors
	}

	if globalViewer == nil {
		// Create new global viewer
		globalViewer = &domain.Viewer{
			TwitchID:    twitchID,
			Username:    username,
			DisplayName: displayName,
			FirstSeen:   now,
			LastSeen:    now,
		}
		uc.viewerRepo.Create(ctx, globalViewer)
	} else {
		// Update global viewer's last seen
		globalViewer.Username = username
		globalViewer.DisplayName = displayName
		globalViewer.LastSeen = now
		uc.viewerRepo.Update(ctx, globalViewer)
	}

	return channelViewer, nil
}

func (uc *channelViewerUseCase) UpdateLastSeen(ctx context.Context, channel, twitchID string) error {
	return uc.channelViewerRepo.UpdateLastSeen(ctx, channel, twitchID)
}

func (uc *channelViewerUseCase) GetViewer(ctx context.Context, channel, twitchID string) (*domain.ChannelViewer, error) {
	return uc.channelViewerRepo.GetByChannelAndTwitchID(ctx, channel, twitchID)
}

func (uc *channelViewerUseCase) ListViewers(ctx context.Context, channel string, limit, offset int) ([]*domain.ChannelViewer, int64, error) {
	return uc.channelViewerRepo.ListByChannel(ctx, channel, limit, offset)
}

func (uc *channelViewerUseCase) RemoveViewer(ctx context.Context, channel, twitchID string) error {
	return uc.channelViewerRepo.Delete(ctx, channel, twitchID)
}

func (uc *channelViewerUseCase) CountViewers(ctx context.Context, channel string) (int64, error) {
	return uc.channelViewerRepo.CountByChannel(ctx, channel)
}
