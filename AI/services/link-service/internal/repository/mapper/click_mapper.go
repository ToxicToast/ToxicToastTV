package mapper

import (
	"toxictoast/services/link-service/internal/domain"
	"toxictoast/services/link-service/internal/repository/entity"
)

// ClickToEntity converts domain.Click to entity.ClickEntity
func ClickToEntity(click *domain.Click) *entity.ClickEntity {
	if click == nil {
		return nil
	}

	return &entity.ClickEntity{
		ID:         click.ID,
		LinkID:     click.LinkID,
		IPAddress:  click.IPAddress,
		UserAgent:  click.UserAgent,
		Referer:    click.Referer,
		Country:    click.Country,
		City:       click.City,
		DeviceType: click.DeviceType,
		ClickedAt:  click.ClickedAt,
		CreatedAt:  click.CreatedAt,
		// Note: We don't convert the Link relation to avoid circular dependencies
	}
}

// ClickToDomain converts entity.ClickEntity to domain.Click
func ClickToDomain(e *entity.ClickEntity) *domain.Click {
	if e == nil {
		return nil
	}

	click := &domain.Click{
		ID:         e.ID,
		LinkID:     e.LinkID,
		IPAddress:  e.IPAddress,
		UserAgent:  e.UserAgent,
		Referer:    e.Referer,
		Country:    e.Country,
		City:       e.City,
		DeviceType: e.DeviceType,
		ClickedAt:  e.ClickedAt,
		CreatedAt:  e.CreatedAt,
	}

	// Convert Link relation if preloaded
	if e.Link.ID != "" {
		click.Link = LinkToDomain(&e.Link)
	}

	return click
}

// ClicksToDomain converts a slice of entity.ClickEntity to domain.Click
func ClicksToDomain(entities []*entity.ClickEntity) []*domain.Click {
	clicks := make([]*domain.Click, len(entities))
	for i, e := range entities {
		clicks[i] = ClickToDomain(e)
	}
	return clicks
}
