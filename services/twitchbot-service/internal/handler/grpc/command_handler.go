package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"toxictoast/services/twitchbot-service/internal/command"
	"toxictoast/services/twitchbot-service/internal/handler/mapper"
	"toxictoast/services/twitchbot-service/internal/query"
	pb "toxictoast/services/twitchbot-service/api/proto"
)

type CommandHandler struct {
	pb.UnimplementedCommandServiceServer
	commandBus *cqrs.CommandBus
	queryBus   *cqrs.QueryBus
}

func NewCommandHandler(commandBus *cqrs.CommandBus, queryBus *cqrs.QueryBus) *CommandHandler {
	return &CommandHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
	}
}

func (h *CommandHandler) CreateCommand(ctx context.Context, req *pb.CreateCommandRequest) (*pb.CreateCommandResponse, error) {
	cmd := &command.CreateCommandCommand{
		BaseCommand:     cqrs.BaseCommand{},
		Name:            req.Name,
		Description:     req.Description,
		Response:        req.Response,
		IsActive:        req.IsActive,
		ModeratorOnly:   req.ModeratorOnly,
		SubscriberOnly:  req.SubscriberOnly,
		CooldownSeconds: int(req.CooldownSeconds),
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Fetch the created command
	qry := &query.GetCommandByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        cmd.AggregateID,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	commandResult := result.(*query.GetCommandResult)

	return &pb.CreateCommandResponse{
		Command: mapper.CommandToProto(commandResult.Command),
	}, nil
}

func (h *CommandHandler) GetCommand(ctx context.Context, req *pb.IdRequest) (*pb.GetCommandResponse, error) {
	qry := &query.GetCommandByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "command not found")
	}

	commandResult := result.(*query.GetCommandResult)

	return &pb.GetCommandResponse{
		Command: mapper.CommandToProto(commandResult.Command),
	}, nil
}

func (h *CommandHandler) ListCommands(ctx context.Context, req *pb.ListCommandsRequest) (*pb.ListCommandsResponse, error) {
	page := int(req.Offset)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.Limit)
	if pageSize <= 0 {
		pageSize = 10
	}

	includeDeleted := false
	if req.DeletedFilter != nil {
		includeDeleted = req.DeletedFilter.IncludeDeleted
	}

	qry := &query.ListCommandsQuery{
		BaseQuery:      cqrs.BaseQuery{},
		Page:           page,
		PageSize:       pageSize,
		OnlyActive:     req.OnlyActive,
		IncludeDeleted: includeDeleted,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	listResult := result.(*query.ListCommandsResult)

	return &pb.ListCommandsResponse{
		Commands: mapper.CommandsToProto(listResult.Commands),
		Total:    int32(listResult.Total),
	}, nil
}

func (h *CommandHandler) UpdateCommand(ctx context.Context, req *pb.UpdateCommandRequest) (*pb.UpdateCommandResponse, error) {
	var name, description, response *string
	var isActive, moderatorOnly, subscriberOnly *bool
	var cooldownSeconds *int

	if req.Name != nil {
		name = req.Name
	}
	if req.Description != nil {
		description = req.Description
	}
	if req.Response != nil {
		response = req.Response
	}
	if req.IsActive != nil {
		isActive = req.IsActive
	}
	if req.ModeratorOnly != nil {
		moderatorOnly = req.ModeratorOnly
	}
	if req.SubscriberOnly != nil {
		subscriberOnly = req.SubscriberOnly
	}
	if req.CooldownSeconds != nil {
		cs := int(*req.CooldownSeconds)
		cooldownSeconds = &cs
	}

	cmd := &command.UpdateCommandCommand{
		BaseCommand:     cqrs.BaseCommand{AggregateID: req.Id},
		Name:            name,
		Description:     description,
		Response:        response,
		IsActive:        isActive,
		ModeratorOnly:   moderatorOnly,
		SubscriberOnly:  subscriberOnly,
		CooldownSeconds: cooldownSeconds,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Fetch the updated command
	qry := &query.GetCommandByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "command not found")
	}

	commandResult := result.(*query.GetCommandResult)

	return &pb.UpdateCommandResponse{
		Command: mapper.CommandToProto(commandResult.Command),
	}, nil
}

func (h *CommandHandler) DeleteCommand(ctx context.Context, req *pb.IdRequest) (*pb.DeleteResponse, error) {
	cmd := &command.DeleteCommandCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Command deleted successfully",
	}, nil
}

func (h *CommandHandler) GetCommandByName(ctx context.Context, req *pb.GetCommandByNameRequest) (*pb.GetCommandResponse, error) {
	qry := &query.GetCommandByNameQuery{
		BaseQuery: cqrs.BaseQuery{},
		Name:      req.Name,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "command not found")
	}

	commandResult := result.(*query.GetCommandResult)

	return &pb.GetCommandResponse{
		Command: mapper.CommandToProto(commandResult.Command),
	}, nil
}

func (h *CommandHandler) ExecuteCommand(ctx context.Context, req *pb.ExecuteCommandRequest) (*pb.ExecuteCommandResponse, error) {
	cmd := &command.ExecuteCommandCommand{
		BaseCommand:  cqrs.BaseCommand{},
		Name:         req.CommandName,
		UserID:       req.UserId,
		Username:     req.Username,
		IsModerator:  req.IsModerator,
		IsSubscriber: req.IsSubscriber,
	}

	err := h.commandBus.Dispatch(ctx, cmd)

	// Handle specific errors
	if err != nil {
		if err.Error() == "command not found" {
			return &pb.ExecuteCommandResponse{
				Success:           false,
				Response:          "",
				Error:             "command not found",
				CooldownRemaining: 0,
			}, nil
		}
		if err.Error() == "command is not active" {
			return &pb.ExecuteCommandResponse{
				Success:           false,
				Response:          "",
				Error:             "command is not active",
				CooldownRemaining: 0,
			}, nil
		}
		if err.Error() == "user not authorized to use this command" {
			return &pb.ExecuteCommandResponse{
				Success:           false,
				Response:          "",
				Error:             "not authorized to use this command",
				CooldownRemaining: 0,
			}, nil
		}
		if err.Error() == "command is on cooldown" {
			return &pb.ExecuteCommandResponse{
				Success:           false,
				Response:          "",
				Error:             "command is on cooldown",
				CooldownRemaining: 0, // TODO: Calculate remaining cooldown
			}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Get the command to return its response
	qry := &query.GetCommandByNameQuery{
		BaseQuery: cqrs.BaseQuery{},
		Name:      req.CommandName,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	commandResult := result.(*query.GetCommandResult)

	return &pb.ExecuteCommandResponse{
		Success:           true,
		Response:          commandResult.Command.Response,
		Error:             "",
		CooldownRemaining: 0,
	}, nil
}
