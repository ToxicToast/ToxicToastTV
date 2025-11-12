package mapper

import (
	"toxictoast/services/blog-service/internal/domain"
	"toxictoast/services/blog-service/internal/repository/entity"

	"gorm.io/gorm"
)

// MediaToEntity converts domain model to database entity
func MediaToEntity(media *domain.Media) *entity.MediaEntity {
	if media == nil {
		return nil
	}

	e := &entity.MediaEntity{
		ID:               media.ID,
		Filename:         media.Filename,
		OriginalFilename: media.OriginalFilename,
		MimeType:         media.MimeType,
		Size:             media.Size,
		Path:             media.Path,
		URL:              media.URL,
		ThumbnailURL:     media.ThumbnailURL,
		Width:            media.Width,
		Height:           media.Height,
		UploadedBy:       media.UploadedBy,
		CreatedAt:        media.CreatedAt,
	}

	// Convert DeletedAt
	if media.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *media.DeletedAt,
			Valid: true,
		}
	}

	return e
}

// MediaToDomain converts database entity to domain model
func MediaToDomain(e *entity.MediaEntity) *domain.Media {
	if e == nil {
		return nil
	}

	media := &domain.Media{
		ID:               e.ID,
		Filename:         e.Filename,
		OriginalFilename: e.OriginalFilename,
		MimeType:         e.MimeType,
		Size:             e.Size,
		Path:             e.Path,
		URL:              e.URL,
		ThumbnailURL:     e.ThumbnailURL,
		Width:            e.Width,
		Height:           e.Height,
		UploadedBy:       e.UploadedBy,
		CreatedAt:        e.CreatedAt,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		media.DeletedAt = &deletedAt
	}

	return media
}

// MediaListToDomain converts slice of entities to domain models
func MediaListToDomain(entities []entity.MediaEntity) []*domain.Media {
	mediaList := make([]*domain.Media, 0, len(entities))
	for _, e := range entities {
		mediaList = append(mediaList, MediaToDomain(&e))
	}
	return mediaList
}
