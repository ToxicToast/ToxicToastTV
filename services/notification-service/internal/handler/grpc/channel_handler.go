package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	pb "toxictoast/services/notification-service/api/proto"
	"toxictoast/services/notification-service/internal/command"
	"toxictoast/services/notification-service/internal/domain"
	"toxictoast/services/notification-service/internal/handler/mapper"
	"toxictoast/services/notification-service/internal/query"
)

type ChannelHandler struct {
	pb.UnimplementedChannelManagementServiceServer
	commandBus *cqrs.CommandBus
	queryBus   *cqrs.QueryBus
}

func NewChannelHandler(commandBus *cqrs.CommandBus, queryBus *cqrs.QueryBus) *ChannelHandler {
	return &ChannelHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
	}
}

func (h *ChannelHandler) CreateChannel(ctx context.Context, req *pb.CreateChannelRequest) (*pb.ChannelResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.WebhookUrl == "" {
		return nil, status.Error(codes.InvalidArgument, "webhook_url is required")
	}

	cmd := &command.CreateChannelCommand{
		BaseCommand: cqrs.BaseCommand{},
		Name:        req.Name,
		WebhookURL:  req.WebhookUrl,
		EventTypes:  req.EventTypes,
		Color:       int(req.Color),
		Description: req.Description,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create channel: %v", err)
	}

	// Query the created channel
	qry := &query.GetChannelByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        cmd.AggregateID,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "channel not found")
	}

	channel := result.(*domain.DiscordChannel)

	return &pb.ChannelResponse{Channel: mapper.ToProtoChannel(channel)}, nil
}

func (h *ChannelHandler) GetChannel(ctx context.Context, req *pb.GetChannelRequest) (*pb.ChannelResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	qry := &query.GetChannelByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "channel not found: %v", err)
	}

	channel := result.(*domain.DiscordChannel)

	return &pb.ChannelResponse{Channel: mapper.ToProtoChannel(channel)}, nil
}

func (h *ChannelHandler) ListChannels(ctx context.Context, req *pb.ListChannelsRequest) (*pb.ListChannelsResponse, error) {
	limit := int(req.Limit)
	if limit == 0 {
		limit = 50
	}
	if limit > 1000 {
		return nil, status.Error(codes.InvalidArgument, "limit cannot exceed 1000")
	}

	qry := &query.ListChannelsQuery{
		BaseQuery:  cqrs.BaseQuery{},
		Limit:      limit,
		Offset:     int(req.Offset),
		ActiveOnly: req.ActiveOnly,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list channels: %v", err)
	}

	listResult := result.(*query.ListChannelsResult)

	return &pb.ListChannelsResponse{
		Channels: mapper.ToProtoChannels(listResult.Channels),
		Total:    listResult.Total,
		Limit:    int32(limit),
		Offset:   req.Offset,
	}, nil
}

func (h *ChannelHandler) UpdateChannel(ctx context.Context, req *pb.UpdateChannelRequest) (*pb.ChannelResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	cmd := &command.UpdateChannelCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if req.Name != "" {
		cmd.Name = &req.Name
	}
	if req.WebhookUrl != "" {
		cmd.WebhookURL = &req.WebhookUrl
	}
	if len(req.EventTypes) > 0 {
		cmd.EventTypes = &req.EventTypes
	}
	if req.Color > 0 {
		color := int(req.Color)
		cmd.Color = &color
	}
	if req.Description != "" {
		cmd.Description = &req.Description
	}
	cmd.Active = &req.Active

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update channel: %v", err)
	}

	// Query the updated channel
	qry := &query.GetChannelByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "channel not found")
	}

	channel := result.(*domain.DiscordChannel)

	return &pb.ChannelResponse{Channel: mapper.ToProtoChannel(channel)}, nil
}

func (h *ChannelHandler) DeleteChannel(ctx context.Context, req *pb.DeleteChannelRequest) (*pb.DeleteChannelResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	cmd := &command.DeleteChannelCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete channel: %v", err)
	}

	return &pb.DeleteChannelResponse{Success: true, Message: "Channel deleted successfully"}, nil
}

func (h *ChannelHandler) ToggleChannel(ctx context.Context, req *pb.ToggleChannelRequest) (*pb.ChannelResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	cmd := &command.ToggleChannelCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
		Active:      req.Active,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to toggle channel: %v", err)
	}

	// Query the updated channel
	qry := &query.GetChannelByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "channel not found")
	}

	channel := result.(*domain.DiscordChannel)

	return &pb.ChannelResponse{Channel: mapper.ToProtoChannel(channel)}, nil
}

func (h *ChannelHandler) TestChannel(ctx context.Context, req *pb.TestChannelRequest) (*pb.TestChannelResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	cmd := &command.TestChannelCommand{
		BaseCommand: cqrs.BaseCommand{},
		ChannelID:   req.Id,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to test channel: %v", err)
	}

	return &pb.TestChannelResponse{Success: true, Message: "Test notification sent successfully"}, nil
}
