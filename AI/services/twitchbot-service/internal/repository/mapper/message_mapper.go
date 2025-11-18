package mapper

import (
	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/entity"

	"gorm.io/gorm"
)

// MessageToEntity converts domain model to database entity
func MessageToEntity(message *domain.Message) *entity.MessageEntity {
	if message == nil {
		return nil
	}

	e := &entity.MessageEntity{
		ID:            message.ID,
		StreamID:      message.StreamID,
		UserID:        message.UserID,
		Username:      message.Username,
		DisplayName:   message.DisplayName,
		Message:       message.Message,
		IsModerator:   message.IsModerator,
		IsSubscriber:  message.IsSubscriber,
		IsVIP:         message.IsVIP,
		IsBroadcaster: message.IsBroadcaster,
		SentAt:        message.SentAt,
		CreatedAt:     message.CreatedAt,
	}

	// Convert DeletedAt
	if message.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *message.DeletedAt,
			Valid: true,
		}
	}

	// Convert Stream relation
	if message.Stream != nil {
		e.Stream = StreamToEntity(message.Stream)
	}

	// Convert Viewer relation
	if message.Viewer != nil {
		e.Viewer = ViewerToEntity(message.Viewer)
	}

	return e
}

// MessageToDomain converts database entity to domain model
func MessageToDomain(e *entity.MessageEntity) *domain.Message {
	if e == nil {
		return nil
	}

	message := &domain.Message{
		ID:            e.ID,
		StreamID:      e.StreamID,
		UserID:        e.UserID,
		Username:      e.Username,
		DisplayName:   e.DisplayName,
		Message:       e.Message,
		IsModerator:   e.IsModerator,
		IsSubscriber:  e.IsSubscriber,
		IsVIP:         e.IsVIP,
		IsBroadcaster: e.IsBroadcaster,
		SentAt:        e.SentAt,
		CreatedAt:     e.CreatedAt,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		message.DeletedAt = &deletedAt
	}

	// Convert Stream relation
	if e.Stream != nil {
		message.Stream = StreamToDomain(e.Stream)
	}

	// Convert Viewer relation
	if e.Viewer != nil {
		message.Viewer = ViewerToDomain(e.Viewer)
	}

	return message
}

// MessagesToDomain converts slice of entities to domain models
func MessagesToDomain(entities []entity.MessageEntity) []*domain.Message {
	messages := make([]*domain.Message, 0, len(entities))
	for _, e := range entities {
		messages = append(messages, MessageToDomain(&e))
	}
	return messages
}
