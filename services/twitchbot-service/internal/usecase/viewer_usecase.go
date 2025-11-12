package usecase

import (
	"context"
	"errors"
	"time"

	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/interfaces"
)

var (
	ErrViewerNotFound    = errors.New("viewer not found")
	ErrInvalidViewerData = errors.New("invalid viewer data")
)

type ViewerUseCase interface {
	CreateViewer(ctx context.Context, twitchID, username, displayName string) (*domain.Viewer, error)
	GetViewerByID(ctx context.Context, id string) (*domain.Viewer, error)
	GetViewerByTwitchID(ctx context.Context, twitchID string) (*domain.Viewer, error)
	ListViewers(ctx context.Context, page, pageSize int, orderBy string, includeDeleted bool) ([]*domain.Viewer, int64, error)
	UpdateViewer(ctx context.Context, id string, username *string, displayName *string, totalMessages *int, totalStreamsWatched *int) (*domain.Viewer, error)
	GetViewerStats(ctx context.Context, id string) (totalMessages int, totalStreamsWatched int, daysSinceFirstSeen int, daysSinceLastSeen int, err error)
	DeleteViewer(ctx context.Context, id string) error
}

type viewerUseCase struct {
	viewerRepo interfaces.ViewerRepository
}

func NewViewerUseCase(viewerRepo interfaces.ViewerRepository) ViewerUseCase {
	return &viewerUseCase{
		viewerRepo: viewerRepo,
	}
}

func (uc *viewerUseCase) CreateViewer(ctx context.Context, twitchID, username, displayName string) (*domain.Viewer, error) {
	if twitchID == "" || username == "" {
		return nil, ErrInvalidViewerData
	}

	viewer := &domain.Viewer{
		TwitchID:            twitchID,
		Username:            username,
		DisplayName:         displayName,
		TotalMessages:       0,
		TotalStreamsWatched: 0,
		FirstSeen:           time.Now(),
		LastSeen:            time.Now(),
	}

	if err := uc.viewerRepo.Create(ctx, viewer); err != nil {
		return nil, err
	}

	return viewer, nil
}

func (uc *viewerUseCase) GetViewerByID(ctx context.Context, id string) (*domain.Viewer, error) {
	viewer, err := uc.viewerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if viewer == nil {
		return nil, ErrViewerNotFound
	}

	return viewer, nil
}

func (uc *viewerUseCase) GetViewerByTwitchID(ctx context.Context, twitchID string) (*domain.Viewer, error) {
	viewer, err := uc.viewerRepo.GetByTwitchID(ctx, twitchID)
	if err != nil {
		return nil, err
	}

	if viewer == nil {
		return nil, ErrViewerNotFound
	}

	return viewer, nil
}

func (uc *viewerUseCase) ListViewers(ctx context.Context, page, pageSize int, orderBy string, includeDeleted bool) ([]*domain.Viewer, int64, error) {
	offset := (page - 1) * pageSize

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return uc.viewerRepo.List(ctx, offset, pageSize, orderBy, includeDeleted)
}

func (uc *viewerUseCase) UpdateViewer(ctx context.Context, id string, username *string, displayName *string, totalMessages *int, totalStreamsWatched *int) (*domain.Viewer, error) {
	viewer, err := uc.viewerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if viewer == nil {
		return nil, ErrViewerNotFound
	}

	if username != nil {
		viewer.Username = *username
	}
	if displayName != nil {
		viewer.DisplayName = *displayName
	}
	if totalMessages != nil {
		viewer.TotalMessages = *totalMessages
	}
	if totalStreamsWatched != nil {
		viewer.TotalStreamsWatched = *totalStreamsWatched
	}

	if err := uc.viewerRepo.Update(ctx, viewer); err != nil {
		return nil, err
	}

	return viewer, nil
}

func (uc *viewerUseCase) GetViewerStats(ctx context.Context, id string) (totalMessages int, totalStreamsWatched int, daysSinceFirstSeen int, daysSinceLastSeen int, err error) {
	viewer, err := uc.viewerRepo.GetByID(ctx, id)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	if viewer == nil {
		return 0, 0, 0, 0, ErrViewerNotFound
	}

	daysSinceFirstSeen = int(time.Since(viewer.FirstSeen).Hours() / 24)
	daysSinceLastSeen = int(time.Since(viewer.LastSeen).Hours() / 24)

	return viewer.TotalMessages, viewer.TotalStreamsWatched, daysSinceFirstSeen, daysSinceLastSeen, nil
}

func (uc *viewerUseCase) DeleteViewer(ctx context.Context, id string) error {
	viewer, err := uc.viewerRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if viewer == nil {
		return ErrViewerNotFound
	}

	return uc.viewerRepo.Delete(ctx, id)
}
