package usecase

import (
	"context"
	"errors"
	"time"

	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/interfaces"
)

var (
	ErrCommandNotFound      = errors.New("command not found")
	ErrInvalidCommandData   = errors.New("invalid command data")
	ErrCommandOnCooldown    = errors.New("command is on cooldown")
	ErrCommandNotActive     = errors.New("command is not active")
	ErrCommandNotAuthorized = errors.New("user not authorized to use this command")
)

type CommandUseCase interface {
	CreateCommand(ctx context.Context, name, description, response string, isActive, moderatorOnly, subscriberOnly bool, cooldownSeconds int) (*domain.Command, error)
	GetCommandByID(ctx context.Context, id string) (*domain.Command, error)
	GetCommandByName(ctx context.Context, name string) (*domain.Command, error)
	ListCommands(ctx context.Context, page, pageSize int, onlyActive bool, includeDeleted bool) ([]*domain.Command, int64, error)
	UpdateCommand(ctx context.Context, id string, name *string, description *string, response *string, isActive *bool, moderatorOnly *bool, subscriberOnly *bool, cooldownSeconds *int) (*domain.Command, error)
	ExecuteCommand(ctx context.Context, commandName, userID, username string, isModerator, isSubscriber bool) (success bool, response string, err error, cooldownRemaining int)
	DeleteCommand(ctx context.Context, id string) error
}

type commandUseCase struct {
	commandRepo interfaces.CommandRepository
}

func NewCommandUseCase(commandRepo interfaces.CommandRepository) CommandUseCase {
	return &commandUseCase{
		commandRepo: commandRepo,
	}
}

func (uc *commandUseCase) CreateCommand(ctx context.Context, name, description, response string, isActive, moderatorOnly, subscriberOnly bool, cooldownSeconds int) (*domain.Command, error) {
	if name == "" || response == "" {
		return nil, ErrInvalidCommandData
	}

	command := &domain.Command{
		Name:            name,
		Description:     description,
		Response:        response,
		IsActive:        isActive,
		ModeratorOnly:   moderatorOnly,
		SubscriberOnly:  subscriberOnly,
		CooldownSeconds: cooldownSeconds,
		UsageCount:      0,
	}

	if err := uc.commandRepo.Create(ctx, command); err != nil {
		return nil, err
	}

	return command, nil
}

func (uc *commandUseCase) GetCommandByID(ctx context.Context, id string) (*domain.Command, error) {
	command, err := uc.commandRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if command == nil {
		return nil, ErrCommandNotFound
	}

	return command, nil
}

func (uc *commandUseCase) GetCommandByName(ctx context.Context, name string) (*domain.Command, error) {
	command, err := uc.commandRepo.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}

	if command == nil {
		return nil, ErrCommandNotFound
	}

	return command, nil
}

func (uc *commandUseCase) ListCommands(ctx context.Context, page, pageSize int, onlyActive bool, includeDeleted bool) ([]*domain.Command, int64, error) {
	offset := (page - 1) * pageSize

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return uc.commandRepo.List(ctx, offset, pageSize, onlyActive, includeDeleted)
}

func (uc *commandUseCase) UpdateCommand(ctx context.Context, id string, name *string, description *string, response *string, isActive *bool, moderatorOnly *bool, subscriberOnly *bool, cooldownSeconds *int) (*domain.Command, error) {
	command, err := uc.commandRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if command == nil {
		return nil, ErrCommandNotFound
	}

	if name != nil {
		command.Name = *name
	}
	if description != nil {
		command.Description = *description
	}
	if response != nil {
		command.Response = *response
	}
	if isActive != nil {
		command.IsActive = *isActive
	}
	if moderatorOnly != nil {
		command.ModeratorOnly = *moderatorOnly
	}
	if subscriberOnly != nil {
		command.SubscriberOnly = *subscriberOnly
	}
	if cooldownSeconds != nil {
		command.CooldownSeconds = *cooldownSeconds
	}

	if err := uc.commandRepo.Update(ctx, command); err != nil {
		return nil, err
	}

	return command, nil
}

func (uc *commandUseCase) ExecuteCommand(ctx context.Context, commandName, userID, username string, isModerator, isSubscriber bool) (success bool, response string, err error, cooldownRemaining int) {
	command, err := uc.commandRepo.GetByName(ctx, commandName)
	if err != nil {
		return false, "", err, 0
	}

	if command == nil {
		return false, "", ErrCommandNotFound, 0
	}

	// Check if command is active
	if !command.IsActive {
		return false, "", ErrCommandNotActive, 0
	}

	// Check permissions
	if command.ModeratorOnly && !isModerator {
		return false, "", ErrCommandNotAuthorized, 0
	}

	if command.SubscriberOnly && !isSubscriber && !isModerator {
		return false, "", ErrCommandNotAuthorized, 0
	}

	// Check cooldown
	if command.LastUsed != nil && command.CooldownSeconds > 0 {
		timeSinceLastUse := time.Since(*command.LastUsed)
		cooldownDuration := time.Duration(command.CooldownSeconds) * time.Second

		if timeSinceLastUse < cooldownDuration {
			remaining := int(cooldownDuration.Seconds() - timeSinceLastUse.Seconds())
			return false, "", ErrCommandOnCooldown, remaining
		}
	}

	// Increment usage
	if err := uc.commandRepo.IncrementUsage(ctx, command.ID); err != nil {
		return false, "", err, 0
	}

	return true, command.Response, nil, 0
}

func (uc *commandUseCase) DeleteCommand(ctx context.Context, id string) error {
	command, err := uc.commandRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if command == nil {
		return ErrCommandNotFound
	}

	return uc.commandRepo.Delete(ctx, id)
}
