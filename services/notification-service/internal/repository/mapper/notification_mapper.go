package mapper

import (
	"toxictoast/services/notification-service/internal/domain"
	"toxictoast/services/notification-service/internal/repository/entity"

	"gorm.io/gorm"
)

// NotificationToEntity converts domain model to database entity
func NotificationToEntity(notification *domain.Notification) *entity.NotificationEntity {
	if notification == nil {
		return nil
	}

	e := &entity.NotificationEntity{
		ID:               notification.ID,
		ChannelID:        notification.ChannelID,
		EventID:          notification.EventID,
		EventType:        notification.EventType,
		EventPayload:     notification.EventPayload,
		DiscordMessageID: notification.DiscordMessageID,
		Status:           string(notification.Status),
		AttemptCount:     notification.AttemptCount,
		LastError:        notification.LastError,
		SentAt:           notification.SentAt,
		CreatedAt:        notification.CreatedAt,
		UpdatedAt:        notification.UpdatedAt,
	}

	// Convert DeletedAt
	if notification.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *notification.DeletedAt,
			Valid: true,
		}
	}

	// Convert Channel relation
	if notification.Channel != nil {
		e.Channel = DiscordChannelToEntity(notification.Channel)
	}

	// Convert Attempts relation
	if len(notification.Attempts) > 0 {
		e.Attempts = make([]entity.NotificationAttemptEntity, 0, len(notification.Attempts))
		for _, attempt := range notification.Attempts {
			e.Attempts = append(e.Attempts, *NotificationAttemptToEntity(&attempt))
		}
	}

	return e
}

// NotificationToDomain converts database entity to domain model
func NotificationToDomain(e *entity.NotificationEntity) *domain.Notification {
	if e == nil {
		return nil
	}

	notification := &domain.Notification{
		ID:               e.ID,
		ChannelID:        e.ChannelID,
		EventID:          e.EventID,
		EventType:        e.EventType,
		EventPayload:     e.EventPayload,
		DiscordMessageID: e.DiscordMessageID,
		Status:           domain.NotificationStatus(e.Status),
		AttemptCount:     e.AttemptCount,
		LastError:        e.LastError,
		SentAt:           e.SentAt,
		CreatedAt:        e.CreatedAt,
		UpdatedAt:        e.UpdatedAt,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		notification.DeletedAt = &deletedAt
	}

	// Convert Channel relation
	if e.Channel != nil {
		notification.Channel = DiscordChannelToDomain(e.Channel)
	}

	// Convert Attempts relation
	if len(e.Attempts) > 0 {
		notification.Attempts = make([]domain.NotificationAttempt, 0, len(e.Attempts))
		for _, attempt := range e.Attempts {
			notification.Attempts = append(notification.Attempts, *NotificationAttemptToDomain(&attempt))
		}
	}

	return notification
}

// NotificationsToDomain converts slice of entities to domain models
func NotificationsToDomain(entities []entity.NotificationEntity) []*domain.Notification {
	notifications := make([]*domain.Notification, 0, len(entities))
	for _, e := range entities {
		notifications = append(notifications, NotificationToDomain(&e))
	}
	return notifications
}

// NotificationAttemptToEntity converts domain model to database entity
func NotificationAttemptToEntity(attempt *domain.NotificationAttempt) *entity.NotificationAttemptEntity {
	if attempt == nil {
		return nil
	}

	e := &entity.NotificationAttemptEntity{
		ID:               attempt.ID,
		NotificationID:   attempt.NotificationID,
		AttemptNumber:    attempt.AttemptNumber,
		ResponseStatus:   attempt.ResponseStatus,
		ResponseBody:     attempt.ResponseBody,
		DiscordMessageID: attempt.DiscordMessageID,
		Success:          attempt.Success,
		Error:            attempt.Error,
		DurationMs:       attempt.DurationMs,
		CreatedAt:        attempt.CreatedAt,
	}

	// Convert DeletedAt
	if attempt.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *attempt.DeletedAt,
			Valid: true,
		}
	}

	return e
}

// NotificationAttemptToDomain converts database entity to domain model
func NotificationAttemptToDomain(e *entity.NotificationAttemptEntity) *domain.NotificationAttempt {
	if e == nil {
		return nil
	}

	attempt := &domain.NotificationAttempt{
		ID:               e.ID,
		NotificationID:   e.NotificationID,
		AttemptNumber:    e.AttemptNumber,
		ResponseStatus:   e.ResponseStatus,
		ResponseBody:     e.ResponseBody,
		DiscordMessageID: e.DiscordMessageID,
		Success:          e.Success,
		Error:            e.Error,
		DurationMs:       e.DurationMs,
		CreatedAt:        e.CreatedAt,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		attempt.DeletedAt = &deletedAt
	}

	return attempt
}

// NotificationAttemptsToDomain converts slice of entities to domain models
func NotificationAttemptsToDomain(entities []entity.NotificationAttemptEntity) []*domain.NotificationAttempt {
	attempts := make([]*domain.NotificationAttempt, 0, len(entities))
	for _, e := range entities {
		attempts = append(attempts, NotificationAttemptToDomain(&e))
	}
	return attempts
}
