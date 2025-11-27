package command

import (
	"context"
	"errors"
	"time"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/interfaces"
)

type CreateCommandCommand struct {
	cqrs.BaseCommand
	Name            string `json:"name"`
	Description     string `json:"description"`
	Response        string `json:"response"`
	IsActive        bool   `json:"is_active"`
	ModeratorOnly   bool   `json:"moderator_only"`
	SubscriberOnly  bool   `json:"subscriber_only"`
	CooldownSeconds int    `json:"cooldown_seconds"`
}

func (c *CreateCommandCommand) CommandName() string { return "create_command" }
func (c *CreateCommandCommand) Validate() error {
	if c.Name == "" || c.Response == "" {
		return errors.New("invalid command data")
	}
	return nil
}

type UpdateCommandCommand struct {
	cqrs.BaseCommand
	Name            *string `json:"name"`
	Description     *string `json:"description"`
	Response        *string `json:"response"`
	IsActive        *bool   `json:"is_active"`
	ModeratorOnly   *bool   `json:"moderator_only"`
	SubscriberOnly  *bool   `json:"subscriber_only"`
	CooldownSeconds *int    `json:"cooldown_seconds"`
}

func (c *UpdateCommandCommand) CommandName() string { return "update_command" }
func (c *UpdateCommandCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("command ID is required")
	}
	return nil
}

type ExecuteCommandCommand struct {
	cqrs.BaseCommand
	Name         string `json:"name"`
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	IsModerator  bool   `json:"is_moderator"`
	IsSubscriber bool   `json:"is_subscriber"`
}

func (c *ExecuteCommandCommand) CommandName() string { return "execute_command" }
func (c *ExecuteCommandCommand) Validate() error {
	if c.Name == "" {
		return errors.New("command name is required")
	}
	return nil
}

type DeleteCommandCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteCommandCommand) CommandName() string { return "delete_command" }
func (c *DeleteCommandCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("command ID is required")
	}
	return nil
}

// Handlers

type CreateCommandHandler struct {
	commandRepo interfaces.CommandRepository
}

func NewCreateCommandHandler(commandRepo interfaces.CommandRepository) *CreateCommandHandler {
	return &CreateCommandHandler{commandRepo: commandRepo}
}

func (h *CreateCommandHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	createCmd := cmd.(*CreateCommandCommand)

	cmdEntity := &domain.Command{
		Name:            createCmd.Name,
		Description:     createCmd.Description,
		Response:        createCmd.Response,
		IsActive:        createCmd.IsActive,
		ModeratorOnly:   createCmd.ModeratorOnly,
		SubscriberOnly:  createCmd.SubscriberOnly,
		CooldownSeconds: createCmd.CooldownSeconds,
		UsageCount:      0,
	}

	if err := h.commandRepo.Create(ctx, cmdEntity); err != nil {
		return err
	}

	createCmd.AggregateID = cmdEntity.ID
	return nil
}

type UpdateCommandHandler struct {
	commandRepo interfaces.CommandRepository
}

func NewUpdateCommandHandler(commandRepo interfaces.CommandRepository) *UpdateCommandHandler {
	return &UpdateCommandHandler{commandRepo: commandRepo}
}

func (h *UpdateCommandHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateCommandCommand)

	cmdEntity, err := h.commandRepo.GetByID(ctx, updateCmd.AggregateID)
	if err != nil {
		return err
	}
	if cmdEntity == nil {
		return errors.New("command not found")
	}

	if updateCmd.Name != nil {
		cmdEntity.Name = *updateCmd.Name
	}
	if updateCmd.Description != nil {
		cmdEntity.Description = *updateCmd.Description
	}
	if updateCmd.Response != nil {
		cmdEntity.Response = *updateCmd.Response
	}
	if updateCmd.IsActive != nil {
		cmdEntity.IsActive = *updateCmd.IsActive
	}
	if updateCmd.ModeratorOnly != nil {
		cmdEntity.ModeratorOnly = *updateCmd.ModeratorOnly
	}
	if updateCmd.SubscriberOnly != nil {
		cmdEntity.SubscriberOnly = *updateCmd.SubscriberOnly
	}
	if updateCmd.CooldownSeconds != nil {
		cmdEntity.CooldownSeconds = *updateCmd.CooldownSeconds
	}

	return h.commandRepo.Update(ctx, cmdEntity)
}

type ExecuteCommandHandler struct {
	commandRepo interfaces.CommandRepository
}

func NewExecuteCommandHandler(commandRepo interfaces.CommandRepository) *ExecuteCommandHandler {
	return &ExecuteCommandHandler{commandRepo: commandRepo}
}

func (h *ExecuteCommandHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	execCmd := cmd.(*ExecuteCommandCommand)

	cmdEntity, err := h.commandRepo.GetByName(ctx, execCmd.Name)
	if err != nil {
		return err
	}
	if cmdEntity == nil {
		return errors.New("command not found")
	}

	// Check if command is active
	if !cmdEntity.IsActive {
		return errors.New("command is not active")
	}

	// Check permissions
	if cmdEntity.ModeratorOnly && !execCmd.IsModerator {
		return errors.New("user not authorized to use this command")
	}

	if cmdEntity.SubscriberOnly && !execCmd.IsSubscriber && !execCmd.IsModerator {
		return errors.New("user not authorized to use this command")
	}

	// Check cooldown
	if cmdEntity.LastUsed != nil && cmdEntity.CooldownSeconds > 0 {
		timeSinceLastUse := time.Since(*cmdEntity.LastUsed)
		cooldownDuration := time.Duration(cmdEntity.CooldownSeconds) * time.Second

		if timeSinceLastUse < cooldownDuration {
			return errors.New("command is on cooldown")
		}
	}

	// Increment usage
	return h.commandRepo.IncrementUsage(ctx, cmdEntity.ID)
}

type DeleteCommandHandler struct {
	commandRepo interfaces.CommandRepository
}

func NewDeleteCommandHandler(commandRepo interfaces.CommandRepository) *DeleteCommandHandler {
	return &DeleteCommandHandler{commandRepo: commandRepo}
}

func (h *DeleteCommandHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteCommandCommand)

	cmdEntity, err := h.commandRepo.GetByID(ctx, deleteCmd.AggregateID)
	if err != nil {
		return err
	}
	if cmdEntity == nil {
		return errors.New("command not found")
	}

	return h.commandRepo.Delete(ctx, deleteCmd.AggregateID)
}
