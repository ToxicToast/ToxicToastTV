package mapper

import (
	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/entity"

	"gorm.io/gorm"
)

// StreamToEntity converts domain model to database entity
func StreamToEntity(stream *domain.Stream) *entity.StreamEntity {
	if stream == nil {
		return nil
	}

	e := &entity.StreamEntity{
		ID:             stream.ID,
		Title:          stream.Title,
		GameName:       stream.GameName,
		GameID:         stream.GameID,
		StartedAt:      stream.StartedAt,
		EndedAt:        stream.EndedAt,
		PeakViewers:    stream.PeakViewers,
		AverageViewers: stream.AverageViewers,
		TotalMessages:  stream.TotalMessages,
		IsActive:       stream.IsActive,
		CreatedAt:      stream.CreatedAt,
		UpdatedAt:      stream.UpdatedAt,
	}

	// Convert DeletedAt
	if stream.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *stream.DeletedAt,
			Valid: true,
		}
	}

	// Note: We don't convert Messages and Clips here to avoid circular dependencies
	// These are loaded by GORM when needed

	return e
}

// StreamToDomain converts database entity to domain model
func StreamToDomain(e *entity.StreamEntity) *domain.Stream {
	if e == nil {
		return nil
	}

	stream := &domain.Stream{
		ID:             e.ID,
		Title:          e.Title,
		GameName:       e.GameName,
		GameID:         e.GameID,
		StartedAt:      e.StartedAt,
		EndedAt:        e.EndedAt,
		PeakViewers:    e.PeakViewers,
		AverageViewers: e.AverageViewers,
		TotalMessages:  e.TotalMessages,
		IsActive:       e.IsActive,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		stream.DeletedAt = &deletedAt
	}

	// Note: Messages and Clips conversion handled by their respective mappers
	// to avoid circular dependencies

	return stream
}

// StreamsToDomain converts slice of entities to domain models
func StreamsToDomain(entities []entity.StreamEntity) []*domain.Stream {
	streams := make([]*domain.Stream, 0, len(entities))
	for _, e := range entities {
		streams = append(streams, StreamToDomain(&e))
	}
	return streams
}
