package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"toxictoast/services/twitchbot-service/internal/handler/mapper"
	"toxictoast/services/twitchbot-service/internal/usecase"
	pb "toxictoast/services/twitchbot-service/api/proto"
)

type CommandHandler struct {
	pb.UnimplementedCommandServiceServer
	commandUC usecase.CommandUseCase
}

func NewCommandHandler(commandUC usecase.CommandUseCase) *CommandHandler {
	return &CommandHandler{
		commandUC: commandUC,
	}
}

func (h *CommandHandler) CreateCommand(ctx context.Context, req *pb.CreateCommandRequest) (*pb.CreateCommandResponse, error) {
	command, err := h.commandUC.CreateCommand(
		ctx,
		req.Name,
		req.Description,
		req.Response,
		req.IsActive,
		req.ModeratorOnly,
		req.SubscriberOnly,
		int(req.CooldownSeconds),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateCommandResponse{
		Command: mapper.CommandToProto(command),
	}, nil
}

func (h *CommandHandler) GetCommand(ctx context.Context, req *pb.IdRequest) (*pb.GetCommandResponse, error) {
	command, err := h.commandUC.GetCommandByID(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrCommandNotFound {
			return nil, status.Error(codes.NotFound, "command not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetCommandResponse{
		Command: mapper.CommandToProto(command),
	}, nil
}

func (h *CommandHandler) ListCommands(ctx context.Context, req *pb.ListCommandsRequest) (*pb.ListCommandsResponse, error) {
	page := req.Offset
	if page <= 0 {
		page = 1
	}
	pageSize := req.Limit
	if pageSize <= 0 {
		pageSize = 10
	}

	includeDeleted := false
	if req.DeletedFilter != nil {
		includeDeleted = req.DeletedFilter.IncludeDeleted
	}

	commands, total, err := h.commandUC.ListCommands(ctx, int(page), int(pageSize), req.OnlyActive, includeDeleted)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ListCommandsResponse{
		Commands: mapper.CommandsToProto(commands),
		Total:    int32(total),
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

	command, err := h.commandUC.UpdateCommand(ctx, req.Id, name, description, response, isActive, moderatorOnly, subscriberOnly, cooldownSeconds)
	if err != nil {
		if err == usecase.ErrCommandNotFound {
			return nil, status.Error(codes.NotFound, "command not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateCommandResponse{
		Command: mapper.CommandToProto(command),
	}, nil
}

func (h *CommandHandler) DeleteCommand(ctx context.Context, req *pb.IdRequest) (*pb.DeleteResponse, error) {
	if err := h.commandUC.DeleteCommand(ctx, req.Id); err != nil {
		if err == usecase.ErrCommandNotFound {
			return nil, status.Error(codes.NotFound, "command not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Command deleted successfully",
	}, nil
}

func (h *CommandHandler) GetCommandByName(ctx context.Context, req *pb.GetCommandByNameRequest) (*pb.GetCommandResponse, error) {
	command, err := h.commandUC.GetCommandByName(ctx, req.Name)
	if err != nil {
		if err == usecase.ErrCommandNotFound {
			return nil, status.Error(codes.NotFound, "command not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetCommandResponse{
		Command: mapper.CommandToProto(command),
	}, nil
}

func (h *CommandHandler) ExecuteCommand(ctx context.Context, req *pb.ExecuteCommandRequest) (*pb.ExecuteCommandResponse, error) {
	success, response, err, cooldownRemaining := h.commandUC.ExecuteCommand(
		ctx,
		req.CommandName,
		req.UserId,
		req.Username,
		req.IsModerator,
		req.IsSubscriber,
	)

	if err != nil {
		if err == usecase.ErrCommandNotFound {
			return &pb.ExecuteCommandResponse{
				Success:           false,
				Response:          "",
				Error:             "command not found",
				CooldownRemaining: 0,
			}, nil
		}
		if err == usecase.ErrCommandNotActive {
			return &pb.ExecuteCommandResponse{
				Success:           false,
				Response:          "",
				Error:             "command is not active",
				CooldownRemaining: 0,
			}, nil
		}
		if err == usecase.ErrCommandNotAuthorized {
			return &pb.ExecuteCommandResponse{
				Success:           false,
				Response:          "",
				Error:             "not authorized to use this command",
				CooldownRemaining: 0,
			}, nil
		}
		if err == usecase.ErrCommandOnCooldown {
			return &pb.ExecuteCommandResponse{
				Success:           false,
				Response:          "",
				Error:             "command is on cooldown",
				CooldownRemaining: int32(cooldownRemaining),
			}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ExecuteCommandResponse{
		Success:           success,
		Response:          response,
		Error:             "",
		CooldownRemaining: 0,
	}, nil
}
