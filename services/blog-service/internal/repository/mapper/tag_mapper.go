package mapper

import (
	"toxictoast/services/blog-service/internal/domain"
	"toxictoast/services/blog-service/internal/repository/entity"

	"gorm.io/gorm"
)

// TagToEntity converts domain model to database entity
func TagToEntity(tag *domain.Tag) *entity.TagEntity {
	if tag == nil {
		return nil
	}

	e := &entity.TagEntity{
		ID:        tag.ID,
		Name:      tag.Name,
		Slug:      tag.Slug,
		CreatedAt: tag.CreatedAt,
		UpdatedAt: tag.UpdatedAt,
	}

	// Convert DeletedAt
	if tag.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *tag.DeletedAt,
			Valid: true,
		}
	}

	return e
}

// TagToDomain converts database entity to domain model
func TagToDomain(e *entity.TagEntity) *domain.Tag {
	if e == nil {
		return nil
	}

	tag := &domain.Tag{
		ID:        e.ID,
		Name:      e.Name,
		Slug:      e.Slug,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		tag.DeletedAt = &deletedAt
	}

	return tag
}

// TagsToDomain converts slice of entities to domain models
func TagsToDomain(entities []entity.TagEntity) []*domain.Tag {
	tags := make([]*domain.Tag, 0, len(entities))
	for _, e := range entities {
		tags = append(tags, TagToDomain(&e))
	}
	return tags
}
