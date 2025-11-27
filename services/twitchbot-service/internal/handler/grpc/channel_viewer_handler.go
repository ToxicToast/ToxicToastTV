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

type ChannelViewerHandler struct {
	pb.UnimplementedChannelViewerServiceServer
	commandBus *cqrs.CommandBus
	queryBus   *cqrs.QueryBus
}

func NewChannelViewerHandler(commandBus *cqrs.CommandBus, queryBus *cqrs.QueryBus) *ChannelViewerHandler {
	return &ChannelViewerHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
	}
}

func (h *ChannelViewerHandler) GetChannelViewer(ctx context.Context, req *pb.GetChannelViewerRequest) (*pb.GetChannelViewerResponse, error) {
	if req.Channel == "" {
		return nil, status.Error(codes.InvalidArgument, "channel is required")
	}
	if req.TwitchId == "" {
		return nil, status.Error(codes.InvalidArgument, "twitch_id is required")
	}

	qry := &query.GetChannelViewerQuery{
		BaseQuery: cqrs.BaseQuery{},
		Channel:   req.Channel,
		TwitchID:  req.TwitchId,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "viewer not found in this channel")
	}

	viewerResult := result.(*query.GetChannelViewerResult)
	if viewerResult.ChannelViewer == nil {
		return nil, status.Error(codes.NotFound, "viewer not found in this channel")
	}

	return &pb.GetChannelViewerResponse{
		Viewer: mapper.ChannelViewerToProto(viewerResult.ChannelViewer),
	}, nil
}

func (h *ChannelViewerHandler) ListChannelViewers(ctx context.Context, req *pb.ListChannelViewersRequest) (*pb.ListChannelViewersResponse, error) {
	if req.Channel == "" {
		return nil, status.Error(codes.InvalidArgument, "channel is required")
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 50 // Default
	}

	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}

	qry := &query.ListChannelViewersQuery{
		BaseQuery: cqrs.BaseQuery{},
		Channel:   req.Channel,
		Limit:     limit,
		Offset:    offset,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	listResult := result.(*query.ListChannelViewersResult)

	return &pb.ListChannelViewersResponse{
		Viewers: mapper.ChannelViewersToProto(listResult.ChannelViewers),
		Total:   int32(listResult.Total),
	}, nil
}

func (h *ChannelViewerHandler) CountChannelViewers(ctx context.Context, req *pb.CountChannelViewersRequest) (*pb.CountChannelViewersResponse, error) {
	if req.Channel == "" {
		return nil, status.Error(codes.InvalidArgument, "channel is required")
	}

	qry := &query.CountChannelViewersQuery{
		BaseQuery: cqrs.BaseQuery{},
		Channel:   req.Channel,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	countResult := result.(*query.CountChannelViewersResult)

	return &pb.CountChannelViewersResponse{
		Count: int32(countResult.Count),
	}, nil
}

func (h *ChannelViewerHandler) RemoveChannelViewer(ctx context.Context, req *pb.RemoveChannelViewerRequest) (*pb.DeleteResponse, error) {
	if req.Channel == "" {
		return nil, status.Error(codes.InvalidArgument, "channel is required")
	}
	if req.TwitchId == "" {
		return nil, status.Error(codes.InvalidArgument, "twitch_id is required")
	}

	cmd := &command.RemoveViewerCommand{
		BaseCommand: cqrs.BaseCommand{},
		Channel:     req.Channel,
		TwitchID:    req.TwitchId,
	}

	err := h.commandBus.Dispatch(ctx, cmd)
	if err != nil {
		return &pb.DeleteResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Viewer removed from channel successfully",
	}, nil
}
