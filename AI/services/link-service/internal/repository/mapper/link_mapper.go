package mapper

import (
	"gorm.io/gorm"

	"toxictoast/services/link-service/internal/domain"
	"toxictoast/services/link-service/internal/repository/entity"
)

// LinkToEntity converts domain.Link to entity.LinkEntity
func LinkToEntity(link *domain.Link) *entity.LinkEntity {
	if link == nil {
		return nil
	}

	e := &entity.LinkEntity{
		ID:          link.ID,
		OriginalURL: link.OriginalURL,
		ShortCode:   link.ShortCode,
		CustomAlias: link.CustomAlias,
		Title:       link.Title,
		Description: link.Description,
		ExpiresAt:   link.ExpiresAt,
		IsActive:    link.IsActive,
		ClickCount:  link.ClickCount,
		CreatedAt:   link.CreatedAt,
		UpdatedAt:   link.UpdatedAt,
	}

	// Convert *time.Time to gorm.DeletedAt
	if link.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *link.DeletedAt,
			Valid: true,
		}
	}

	return e
}

// LinkToDomain converts entity.LinkEntity to domain.Link
func LinkToDomain(e *entity.LinkEntity) *domain.Link {
	if e == nil {
		return nil
	}

	link := &domain.Link{
		ID:          e.ID,
		OriginalURL: e.OriginalURL,
		ShortCode:   e.ShortCode,
		CustomAlias: e.CustomAlias,
		Title:       e.Title,
		Description: e.Description,
		ExpiresAt:   e.ExpiresAt,
		IsActive:    e.IsActive,
		ClickCount:  e.ClickCount,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}

	// Convert gorm.DeletedAt to *time.Time
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		link.DeletedAt = &deletedAt
	}

	return link
}

// LinksToDomain converts a slice of entity.LinkEntity to domain.Link
func LinksToDomain(entities []*entity.LinkEntity) []*domain.Link {
	links := make([]*domain.Link, len(entities))
	for i, e := range entities {
		links[i] = LinkToDomain(e)
	}
	return links
}
