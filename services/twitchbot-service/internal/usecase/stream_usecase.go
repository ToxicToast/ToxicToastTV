package usecase

import (
	"context"
	"errors"
	"time"

	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/interfaces"
)

var (
	ErrStreamNotFound     = errors.New("stream not found")
	ErrInvalidStreamData  = errors.New("invalid stream data")
	ErrStreamAlreadyEnded = errors.New("stream already ended")
	ErrNoActiveStream     = errors.New("no active stream found")
)

type StreamUseCase interface {
	CreateStream(ctx context.Context, title, gameName, gameID string) (*domain.Stream, error)
	GetStreamByID(ctx context.Context, id string) (*domain.Stream, error)
	ListStreams(ctx context.Context, page, pageSize int, onlyActive bool, gameName string, includeDeleted bool) ([]*domain.Stream, int64, error)
	GetActiveStream(ctx context.Context) (*domain.Stream, error)
	UpdateStream(ctx context.Context, id string, title *string, gameName *string, gameID *string, peakViewers *int, averageViewers *int) (*domain.Stream, error)
	EndStream(ctx context.Context, id string) (*domain.Stream, error)
	DeleteStream(ctx context.Context, id string) error
	GetStreamStats(ctx context.Context, id string) (peakViewers int, averageViewers int, totalMessages int, uniqueViewers int64, durationSeconds int64, err error)
}

type streamUseCase struct {
	streamRepo  interfaces.StreamRepository
	messageRepo interfaces.MessageRepository
}

func NewStreamUseCase(streamRepo interfaces.StreamRepository, messageRepo interfaces.MessageRepository) StreamUseCase {
	return &streamUseCase{
		streamRepo:  streamRepo,
		messageRepo: messageRepo,
	}
}

func (uc *streamUseCase) CreateStream(ctx context.Context, title, gameName, gameID string) (*domain.Stream, error) {
	if title == "" {
		return nil, ErrInvalidStreamData
	}

	stream := &domain.Stream{
		Title:          title,
		GameName:       gameName,
		GameID:         gameID,
		StartedAt:      time.Now(),
		PeakViewers:    0,
		AverageViewers: 0,
		TotalMessages:  0,
		IsActive:       true,
	}

	if err := uc.streamRepo.Create(ctx, stream); err != nil {
		return nil, err
	}

	return stream, nil
}

func (uc *streamUseCase) GetStreamByID(ctx context.Context, id string) (*domain.Stream, error) {
	stream, err := uc.streamRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if stream == nil {
		return nil, ErrStreamNotFound
	}

	return stream, nil
}

func (uc *streamUseCase) ListStreams(ctx context.Context, page, pageSize int, onlyActive bool, gameName string, includeDeleted bool) ([]*domain.Stream, int64, error) {
	offset := (page - 1) * pageSize

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return uc.streamRepo.List(ctx, offset, pageSize, onlyActive, gameName, includeDeleted)
}

func (uc *streamUseCase) GetActiveStream(ctx context.Context) (*domain.Stream, error) {
	stream, err := uc.streamRepo.GetActive(ctx)
	if err != nil {
		return nil, err
	}

	if stream == nil {
		return nil, ErrNoActiveStream
	}

	return stream, nil
}

func (uc *streamUseCase) UpdateStream(ctx context.Context, id string, title *string, gameName *string, gameID *string, peakViewers *int, averageViewers *int) (*domain.Stream, error) {
	stream, err := uc.streamRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if stream == nil {
		return nil, ErrStreamNotFound
	}

	if title != nil {
		stream.Title = *title
	}
	if gameName != nil {
		stream.GameName = *gameName
	}
	if gameID != nil {
		stream.GameID = *gameID
	}
	if peakViewers != nil {
		stream.PeakViewers = *peakViewers
	}
	if averageViewers != nil {
		stream.AverageViewers = *averageViewers
	}

	if err := uc.streamRepo.Update(ctx, stream); err != nil {
		return nil, err
	}

	return stream, nil
}

func (uc *streamUseCase) EndStream(ctx context.Context, id string) (*domain.Stream, error) {
	stream, err := uc.streamRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if stream == nil {
		return nil, ErrStreamNotFound
	}

	if !stream.IsActive {
		return nil, ErrStreamAlreadyEnded
	}

	if err := uc.streamRepo.EndStream(ctx, id); err != nil {
		return nil, err
	}

	// Refresh stream data
	return uc.streamRepo.GetByID(ctx, id)
}

func (uc *streamUseCase) DeleteStream(ctx context.Context, id string) error {
	stream, err := uc.streamRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if stream == nil {
		return ErrStreamNotFound
	}

	return uc.streamRepo.Delete(ctx, id)
}

func (uc *streamUseCase) GetStreamStats(ctx context.Context, id string) (peakViewers int, averageViewers int, totalMessages int, uniqueViewers int64, durationSeconds int64, err error) {
	stream, err := uc.streamRepo.GetByID(ctx, id)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	if stream == nil {
		return 0, 0, 0, 0, 0, ErrStreamNotFound
	}

	// Get message stats
	totalMsg, uniqueUsers, _, _, err := uc.messageRepo.GetStats(ctx, id)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	// Calculate duration
	var duration int64
	if stream.EndedAt != nil {
		duration = int64(stream.EndedAt.Sub(stream.StartedAt).Seconds())
	} else {
		duration = int64(time.Since(stream.StartedAt).Seconds())
	}

	return stream.PeakViewers, stream.AverageViewers, int(totalMsg), uniqueUsers, duration, nil
}
