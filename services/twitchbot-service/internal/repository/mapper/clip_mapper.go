package mapper

import (
	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/entity"

	"gorm.io/gorm"
)

// ClipToEntity converts domain model to database entity
func ClipToEntity(clip *domain.Clip) *entity.ClipEntity {
	if clip == nil {
		return nil
	}

	e := &entity.ClipEntity{
		ID:              clip.ID,
		StreamID:        clip.StreamID,
		TwitchClipID:    clip.TwitchClipID,
		Title:           clip.Title,
		URL:             clip.URL,
		EmbedURL:        clip.EmbedURL,
		ThumbnailURL:    clip.ThumbnailURL,
		CreatorName:     clip.CreatorName,
		CreatorID:       clip.CreatorID,
		ViewCount:       clip.ViewCount,
		DurationSeconds: clip.DurationSeconds,
		CreatedAtTwitch: clip.CreatedAtTwitch,
		CreatedAt:       clip.CreatedAt,
		UpdatedAt:       clip.UpdatedAt,
	}

	// Convert DeletedAt
	if clip.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *clip.DeletedAt,
			Valid: true,
		}
	}

	// Convert Stream relation
	if clip.Stream != nil {
		e.Stream = StreamToEntity(clip.Stream)
	}

	// Convert Creator relation
	if clip.Creator != nil {
		e.Creator = ViewerToEntity(clip.Creator)
	}

	return e
}

// ClipToDomain converts database entity to domain model
func ClipToDomain(e *entity.ClipEntity) *domain.Clip {
	if e == nil {
		return nil
	}

	clip := &domain.Clip{
		ID:              e.ID,
		StreamID:        e.StreamID,
		TwitchClipID:    e.TwitchClipID,
		Title:           e.Title,
		URL:             e.URL,
		EmbedURL:        e.EmbedURL,
		ThumbnailURL:    e.ThumbnailURL,
		CreatorName:     e.CreatorName,
		CreatorID:       e.CreatorID,
		ViewCount:       e.ViewCount,
		DurationSeconds: e.DurationSeconds,
		CreatedAtTwitch: e.CreatedAtTwitch,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		clip.DeletedAt = &deletedAt
	}

	// Convert Stream relation
	if e.Stream != nil {
		clip.Stream = StreamToDomain(e.Stream)
	}

	// Convert Creator relation
	if e.Creator != nil {
		clip.Creator = ViewerToDomain(e.Creator)
	}

	return clip
}

// ClipsToDomain converts slice of entities to domain models
func ClipsToDomain(entities []entity.ClipEntity) []*domain.Clip {
	clips := make([]*domain.Clip, 0, len(entities))
	for _, e := range entities {
		clips = append(clips, ClipToDomain(&e))
	}
	return clips
}
