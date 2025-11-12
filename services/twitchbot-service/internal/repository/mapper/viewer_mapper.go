package mapper

import (
	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/entity"

	"gorm.io/gorm"
)

// ViewerToEntity converts domain model to database entity
func ViewerToEntity(viewer *domain.Viewer) *entity.ViewerEntity {
	if viewer == nil {
		return nil
	}

	e := &entity.ViewerEntity{
		ID:                  viewer.ID,
		TwitchID:            viewer.TwitchID,
		Username:            viewer.Username,
		DisplayName:         viewer.DisplayName,
		TotalMessages:       viewer.TotalMessages,
		TotalStreamsWatched: viewer.TotalStreamsWatched,
		FirstSeen:           viewer.FirstSeen,
		LastSeen:            viewer.LastSeen,
		CreatedAt:           viewer.CreatedAt,
		UpdatedAt:           viewer.UpdatedAt,
	}

	// Convert DeletedAt
	if viewer.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *viewer.DeletedAt,
			Valid: true,
		}
	}

	// Note: We don't convert Messages and Clips here to avoid circular dependencies
	// These are loaded by GORM when needed

	return e
}

// ViewerToDomain converts database entity to domain model
func ViewerToDomain(e *entity.ViewerEntity) *domain.Viewer {
	if e == nil {
		return nil
	}

	viewer := &domain.Viewer{
		ID:                  e.ID,
		TwitchID:            e.TwitchID,
		Username:            e.Username,
		DisplayName:         e.DisplayName,
		TotalMessages:       e.TotalMessages,
		TotalStreamsWatched: e.TotalStreamsWatched,
		FirstSeen:           e.FirstSeen,
		LastSeen:            e.LastSeen,
		CreatedAt:           e.CreatedAt,
		UpdatedAt:           e.UpdatedAt,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		viewer.DeletedAt = &deletedAt
	}

	// Note: Messages and Clips conversion handled by their respective mappers
	// to avoid circular dependencies

	return viewer
}

// ViewersToDomain converts slice of entities to domain models
func ViewersToDomain(entities []entity.ViewerEntity) []*domain.Viewer {
	viewers := make([]*domain.Viewer, 0, len(entities))
	for _, e := range entities {
		viewers = append(viewers, ViewerToDomain(&e))
	}
	return viewers
}
