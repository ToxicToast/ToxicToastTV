package usecase

import (
	"context"
	"errors"
	"time"

	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/interfaces"
)

var (
	ErrClipNotFound    = errors.New("clip not found")
	ErrInvalidClipData = errors.New("invalid clip data")
)

type ClipUseCase interface {
	CreateClip(ctx context.Context, streamID, twitchClipID, title, url, embedURL, thumbnailURL, creatorName, creatorID string, viewCount, durationSeconds int, createdAtTwitch time.Time) (*domain.Clip, error)
	GetClipByID(ctx context.Context, id string) (*domain.Clip, error)
	GetClipByTwitchClipID(ctx context.Context, twitchClipID string) (*domain.Clip, error)
	ListClips(ctx context.Context, page, pageSize int, streamID, orderBy string, includeDeleted bool) ([]*domain.Clip, int64, error)
	UpdateClip(ctx context.Context, id string, title *string, viewCount *int) (*domain.Clip, error)
	DeleteClip(ctx context.Context, id string) error
}

type clipUseCase struct {
	clipRepo   interfaces.ClipRepository
	streamRepo interfaces.StreamRepository
}

func NewClipUseCase(clipRepo interfaces.ClipRepository, streamRepo interfaces.StreamRepository) ClipUseCase {
	return &clipUseCase{
		clipRepo:   clipRepo,
		streamRepo: streamRepo,
	}
}

func (uc *clipUseCase) CreateClip(ctx context.Context, streamID, twitchClipID, title, url, embedURL, thumbnailURL, creatorName, creatorID string, viewCount, durationSeconds int, createdAtTwitch time.Time) (*domain.Clip, error) {
	if streamID == "" || twitchClipID == "" || title == "" || url == "" {
		return nil, ErrInvalidClipData
	}

	// Verify stream exists
	stream, err := uc.streamRepo.GetByID(ctx, streamID)
	if err != nil {
		return nil, err
	}
	if stream == nil {
		return nil, errors.New("stream not found")
	}

	clip := &domain.Clip{
		StreamID:        streamID,
		TwitchClipID:    twitchClipID,
		Title:           title,
		URL:             url,
		EmbedURL:        embedURL,
		ThumbnailURL:    thumbnailURL,
		CreatorName:     creatorName,
		CreatorID:       creatorID,
		ViewCount:       viewCount,
		DurationSeconds: durationSeconds,
		CreatedAtTwitch: createdAtTwitch,
	}

	if err := uc.clipRepo.Create(ctx, clip); err != nil {
		return nil, err
	}

	return clip, nil
}

func (uc *clipUseCase) GetClipByID(ctx context.Context, id string) (*domain.Clip, error) {
	clip, err := uc.clipRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if clip == nil {
		return nil, ErrClipNotFound
	}

	return clip, nil
}

func (uc *clipUseCase) GetClipByTwitchClipID(ctx context.Context, twitchClipID string) (*domain.Clip, error) {
	clip, err := uc.clipRepo.GetByTwitchClipID(ctx, twitchClipID)
	if err != nil {
		return nil, err
	}

	if clip == nil {
		return nil, ErrClipNotFound
	}

	return clip, nil
}

func (uc *clipUseCase) ListClips(ctx context.Context, page, pageSize int, streamID, orderBy string, includeDeleted bool) ([]*domain.Clip, int64, error) {
	offset := (page - 1) * pageSize

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return uc.clipRepo.List(ctx, offset, pageSize, streamID, orderBy, includeDeleted)
}

func (uc *clipUseCase) UpdateClip(ctx context.Context, id string, title *string, viewCount *int) (*domain.Clip, error) {
	clip, err := uc.clipRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if clip == nil {
		return nil, ErrClipNotFound
	}

	if title != nil {
		clip.Title = *title
	}
	if viewCount != nil {
		clip.ViewCount = *viewCount
	}

	if err := uc.clipRepo.Update(ctx, clip); err != nil {
		return nil, err
	}

	return clip, nil
}

func (uc *clipUseCase) DeleteClip(ctx context.Context, id string) error {
	clip, err := uc.clipRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if clip == nil {
		return ErrClipNotFound
	}

	return uc.clipRepo.Delete(ctx, id)
}
