package mapper

import (
	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/entity"

	"gorm.io/gorm"
)

// ChannelViewerToEntity converts domain model to database entity
func ChannelViewerToEntity(cv *domain.ChannelViewer) *entity.ChannelViewerEntity {
	if cv == nil {
		return nil
	}

	e := &entity.ChannelViewerEntity{
		ID:          cv.ID,
		Channel:     cv.Channel,
		TwitchID:    cv.TwitchID,
		Username:    cv.Username,
		DisplayName: cv.DisplayName,
		FirstSeen:   cv.FirstSeen,
		LastSeen:    cv.LastSeen,
		IsModerator: cv.IsModerator,
		IsVIP:       cv.IsVIP,
		CreatedAt:   cv.CreatedAt,
		UpdatedAt:   cv.UpdatedAt,
	}

	// Convert DeletedAt
	if cv.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *cv.DeletedAt,
			Valid: true,
		}
	}

	// Convert Viewer relation
	if cv.Viewer != nil {
		e.Viewer = ViewerToEntity(cv.Viewer)
	}

	return e
}

// ChannelViewerToDomain converts database entity to domain model
func ChannelViewerToDomain(e *entity.ChannelViewerEntity) *domain.ChannelViewer {
	if e == nil {
		return nil
	}

	cv := &domain.ChannelViewer{
		ID:          e.ID,
		Channel:     e.Channel,
		TwitchID:    e.TwitchID,
		Username:    e.Username,
		DisplayName: e.DisplayName,
		FirstSeen:   e.FirstSeen,
		LastSeen:    e.LastSeen,
		IsModerator: e.IsModerator,
		IsVIP:       e.IsVIP,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		cv.DeletedAt = &deletedAt
	}

	// Convert Viewer relation
	if e.Viewer != nil {
		cv.Viewer = ViewerToDomain(e.Viewer)
	}

	return cv
}

// ChannelViewersToDomain converts slice of entities to domain models
func ChannelViewersToDomain(entities []entity.ChannelViewerEntity) []*domain.ChannelViewer {
	cvs := make([]*domain.ChannelViewer, 0, len(entities))
	for _, e := range entities {
		cvs = append(cvs, ChannelViewerToDomain(&e))
	}
	return cvs
}
