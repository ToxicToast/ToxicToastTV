package query

import (
	"context"
	"errors"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"toxictoast/services/twitchbot-service/internal/domain"
	"toxictoast/services/twitchbot-service/internal/repository/interfaces"
)

type GetCommandByIDQuery struct {
	cqrs.BaseQuery
	ID string `json:"id"`
}

func (q *GetCommandByIDQuery) QueryName() string { return "get_command_by_id" }
func (q *GetCommandByIDQuery) Validate() error {
	if q.ID == "" {
		return errors.New("command ID is required")
	}
	return nil
}

type GetCommandByNameQuery struct {
	cqrs.BaseQuery
	Name string `json:"name"`
}

func (q *GetCommandByNameQuery) QueryName() string { return "get_command_by_name" }
func (q *GetCommandByNameQuery) Validate() error {
	if q.Name == "" {
		return errors.New("command name is required")
	}
	return nil
}

type ListCommandsQuery struct {
	cqrs.BaseQuery
	Page           int  `json:"page"`
	PageSize       int  `json:"page_size"`
	OnlyActive     bool `json:"only_active"`
	IncludeDeleted bool `json:"include_deleted"`
}

func (q *ListCommandsQuery) QueryName() string { return "list_commands" }
func (q *ListCommandsQuery) Validate() error {
	if q.PageSize <= 0 || q.PageSize > 100 {
		q.PageSize = 20
	}
	return nil
}

// Results

type GetCommandResult struct {
	Command *domain.Command
}

type ListCommandsResult struct {
	Commands []*domain.Command
	Total    int64
}

// Handlers

type GetCommandByIDHandler struct {
	commandRepo interfaces.CommandRepository
}

func NewGetCommandByIDHandler(commandRepo interfaces.CommandRepository) *GetCommandByIDHandler {
	return &GetCommandByIDHandler{commandRepo: commandRepo}
}

func (h *GetCommandByIDHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	qry := query.(*GetCommandByIDQuery)

	cmd, err := h.commandRepo.GetByID(ctx, qry.ID)
	if err != nil {
		return nil, err
	}
	if cmd == nil {
		return nil, errors.New("command not found")
	}

	return &GetCommandResult{Command: cmd}, nil
}

type GetCommandByNameHandler struct {
	commandRepo interfaces.CommandRepository
}

func NewGetCommandByNameHandler(commandRepo interfaces.CommandRepository) *GetCommandByNameHandler {
	return &GetCommandByNameHandler{commandRepo: commandRepo}
}

func (h *GetCommandByNameHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	qry := query.(*GetCommandByNameQuery)

	cmd, err := h.commandRepo.GetByName(ctx, qry.Name)
	if err != nil {
		return nil, err
	}
	if cmd == nil {
		return nil, errors.New("command not found")
	}

	return &GetCommandResult{Command: cmd}, nil
}

type ListCommandsHandler struct {
	commandRepo interfaces.CommandRepository
}

func NewListCommandsHandler(commandRepo interfaces.CommandRepository) *ListCommandsHandler {
	return &ListCommandsHandler{commandRepo: commandRepo}
}

func (h *ListCommandsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	qry := query.(*ListCommandsQuery)

	offset := (qry.Page - 1) * qry.PageSize

	commands, total, err := h.commandRepo.List(ctx, offset, qry.PageSize, qry.OnlyActive, qry.IncludeDeleted)
	if err != nil {
		return nil, err
	}

	return &ListCommandsResult{
		Commands: commands,
		Total:    total,
	}, nil
}
