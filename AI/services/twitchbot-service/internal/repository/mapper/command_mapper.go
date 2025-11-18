package mapper

import (
	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/entity"

	"gorm.io/gorm"
)

// CommandToEntity converts domain model to database entity
func CommandToEntity(command *domain.Command) *entity.CommandEntity {
	if command == nil {
		return nil
	}

	e := &entity.CommandEntity{
		ID:              command.ID,
		Name:            command.Name,
		Description:     command.Description,
		Response:        command.Response,
		IsActive:        command.IsActive,
		ModeratorOnly:   command.ModeratorOnly,
		SubscriberOnly:  command.SubscriberOnly,
		CooldownSeconds: command.CooldownSeconds,
		UsageCount:      command.UsageCount,
		LastUsed:        command.LastUsed,
		CreatedAt:       command.CreatedAt,
		UpdatedAt:       command.UpdatedAt,
	}

	// Convert DeletedAt
	if command.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *command.DeletedAt,
			Valid: true,
		}
	}

	return e
}

// CommandToDomain converts database entity to domain model
func CommandToDomain(e *entity.CommandEntity) *domain.Command {
	if e == nil {
		return nil
	}

	command := &domain.Command{
		ID:              e.ID,
		Name:            e.Name,
		Description:     e.Description,
		Response:        e.Response,
		IsActive:        e.IsActive,
		ModeratorOnly:   e.ModeratorOnly,
		SubscriberOnly:  e.SubscriberOnly,
		CooldownSeconds: e.CooldownSeconds,
		UsageCount:      e.UsageCount,
		LastUsed:        e.LastUsed,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		command.DeletedAt = &deletedAt
	}

	return command
}

// CommandsToDomain converts slice of entities to domain models
func CommandsToDomain(entities []entity.CommandEntity) []*domain.Command {
	commands := make([]*domain.Command, 0, len(entities))
	for _, e := range entities {
		commands = append(commands, CommandToDomain(&e))
	}
	return commands
}
