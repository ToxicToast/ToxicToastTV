package mapper

import (
	"toxictoast/services/notification-service/internal/domain"
	"toxictoast/services/notification-service/internal/repository/entity"

	"gorm.io/gorm"
)

// DiscordChannelToEntity converts domain model to database entity
func DiscordChannelToEntity(channel *domain.DiscordChannel) *entity.DiscordChannelEntity {
	if channel == nil {
		return nil
	}

	e := &entity.DiscordChannelEntity{
		ID:                   channel.ID,
		Name:                 channel.Name,
		WebhookURL:           channel.WebhookURL,
		EventTypes:           channel.EventTypes,
		Color:                channel.Color,
		Active:               channel.Active,
		Description:          channel.Description,
		CreatedAt:            channel.CreatedAt,
		UpdatedAt:            channel.UpdatedAt,
		TotalNotifications:   channel.TotalNotifications,
		SuccessNotifications: channel.SuccessNotifications,
		FailedNotifications:  channel.FailedNotifications,
		LastNotificationAt:   channel.LastNotificationAt,
		LastSuccessAt:        channel.LastSuccessAt,
		LastFailureAt:        channel.LastFailureAt,
	}

	// Convert DeletedAt
	if channel.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *channel.DeletedAt,
			Valid: true,
		}
	}

	return e
}

// DiscordChannelToDomain converts database entity to domain model
func DiscordChannelToDomain(e *entity.DiscordChannelEntity) *domain.DiscordChannel {
	if e == nil {
		return nil
	}

	channel := &domain.DiscordChannel{
		ID:                   e.ID,
		Name:                 e.Name,
		WebhookURL:           e.WebhookURL,
		EventTypes:           e.EventTypes,
		Color:                e.Color,
		Active:               e.Active,
		Description:          e.Description,
		CreatedAt:            e.CreatedAt,
		UpdatedAt:            e.UpdatedAt,
		TotalNotifications:   e.TotalNotifications,
		SuccessNotifications: e.SuccessNotifications,
		FailedNotifications:  e.FailedNotifications,
		LastNotificationAt:   e.LastNotificationAt,
		LastSuccessAt:        e.LastSuccessAt,
		LastFailureAt:        e.LastFailureAt,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		channel.DeletedAt = &deletedAt
	}

	return channel
}

// DiscordChannelsToDomain converts slice of entities to domain models
func DiscordChannelsToDomain(entities []entity.DiscordChannelEntity) []*domain.DiscordChannel {
	channels := make([]*domain.DiscordChannel, 0, len(entities))
	for _, e := range entities {
		channels = append(channels, DiscordChannelToDomain(&e))
	}
	return channels
}
