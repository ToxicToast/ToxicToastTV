package cqrs

import (
	"context"
	"errors"
)

var (
	// ErrCommandHandlerNotFound is returned when no handler is registered for a command
	ErrCommandHandlerNotFound = errors.New("command handler not found")

	// ErrCommandValidation is returned when command validation fails
	ErrCommandValidation = errors.New("command validation failed")
)

// Command represents a command in the CQRS pattern
// Commands are intentions to change the system state
type Command interface {
	// CommandName returns the name of the command
	CommandName() string

	// Validate validates the command
	Validate() error
}

// CommandHandler handles a specific type of command
type CommandHandler interface {
	// Handle processes the command and returns an error if it fails
	Handle(ctx context.Context, command Command) error
}

// CommandBus dispatches commands to their handlers
type CommandBus struct {
	handlers map[string]CommandHandler
}

// NewCommandBus creates a new command bus
func NewCommandBus() *CommandBus {
	return &CommandBus{
		handlers: make(map[string]CommandHandler),
	}
}

// RegisterHandler registers a command handler
func (b *CommandBus) RegisterHandler(commandName string, handler CommandHandler) {
	b.handlers[commandName] = handler
}

// Dispatch dispatches a command to its handler
func (b *CommandBus) Dispatch(ctx context.Context, command Command) error {
	// Validate command
	if err := command.Validate(); err != nil {
		return ErrCommandValidation
	}

	// Get handler
	handler, ok := b.handlers[command.CommandName()]
	if !ok {
		return ErrCommandHandlerNotFound
	}

	// Execute handler
	return handler.Handle(ctx, command)
}

// BaseCommand provides common functionality for commands
type BaseCommand struct {
	AggregateID string `json:"aggregate_id"`
}

// GetAggregateID returns the aggregate ID
func (c *BaseCommand) GetAggregateID() string {
	return c.AggregateID
}
