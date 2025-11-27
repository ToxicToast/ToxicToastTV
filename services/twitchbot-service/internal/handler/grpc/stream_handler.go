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

type StreamHandler struct {
	pb.UnimplementedStreamServiceServer
	commandBus *cqrs.CommandBus
	queryBus   *cqrs.QueryBus
}

func NewStreamHandler(commandBus *cqrs.CommandBus, queryBus *cqrs.QueryBus) *StreamHandler {
	return &StreamHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
	}
}

func (h *StreamHandler) CreateStream(ctx context.Context, req *pb.CreateStreamRequest) (*pb.CreateStreamResponse, error) {
	cmd := &command.CreateStreamCommand{
		BaseCommand: cqrs.BaseCommand{},
		Title:       req.Title,
		GameName:    req.GameName,
		GameID:      req.GameId,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Fetch the created stream
	qry := &query.GetStreamByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        cmd.AggregateID,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	streamResult := result.(*query.GetStreamResult)

	return &pb.CreateStreamResponse{
		Stream: mapper.StreamToProto(streamResult.Stream),
	}, nil
}

func (h *StreamHandler) GetStream(ctx context.Context, req *pb.IdRequest) (*pb.GetStreamResponse, error) {
	qry := &query.GetStreamByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "stream not found")
	}

	streamResult := result.(*query.GetStreamResult)

	return &pb.GetStreamResponse{
		Stream: mapper.StreamToProto(streamResult.Stream),
	}, nil
}

func (h *StreamHandler) ListStreams(ctx context.Context, req *pb.ListStreamsRequest) (*pb.ListStreamsResponse, error) {
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

	qry := &query.ListStreamsQuery{
		BaseQuery:      cqrs.BaseQuery{},
		Page:           page,
		PageSize:       pageSize,
		OnlyActive:     req.OnlyActive,
		GameName:       req.GameName,
		IncludeDeleted: includeDeleted,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	listResult := result.(*query.ListStreamsResult)

	return &pb.ListStreamsResponse{
		Streams: mapper.StreamsToProto(listResult.Streams),
		Total:   int32(listResult.Total),
	}, nil
}

func (h *StreamHandler) UpdateStream(ctx context.Context, req *pb.UpdateStreamRequest) (*pb.UpdateStreamResponse, error) {
	var title, gameName, gameID *string
	var peakViewers, averageViewers *int

	if req.Title != nil {
		title = req.Title
	}
	if req.GameName != nil {
		gameName = req.GameName
	}
	if req.GameId != nil {
		gameID = req.GameId
	}
	if req.PeakViewers != nil {
		pv := int(*req.PeakViewers)
		peakViewers = &pv
	}
	if req.AverageViewers != nil {
		av := int(*req.AverageViewers)
		averageViewers = &av
	}

	cmd := &command.UpdateStreamCommand{
		BaseCommand:    cqrs.BaseCommand{AggregateID: req.Id},
		Title:          title,
		GameName:       gameName,
		GameID:         gameID,
		PeakViewers:    peakViewers,
		AverageViewers: averageViewers,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Fetch the updated stream
	qry := &query.GetStreamByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "stream not found")
	}

	streamResult := result.(*query.GetStreamResult)

	return &pb.UpdateStreamResponse{
		Stream: mapper.StreamToProto(streamResult.Stream),
	}, nil
}

func (h *StreamHandler) DeleteStream(ctx context.Context, req *pb.IdRequest) (*pb.DeleteResponse, error) {
	cmd := &command.DeleteStreamCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Stream deleted successfully",
	}, nil
}

func (h *StreamHandler) EndStream(ctx context.Context, req *pb.EndStreamRequest) (*pb.EndStreamResponse, error) {
	cmd := &command.EndStreamCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Fetch the ended stream
	qry := &query.GetStreamByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "stream not found")
	}

	streamResult := result.(*query.GetStreamResult)

	return &pb.EndStreamResponse{
		Stream: mapper.StreamToProto(streamResult.Stream),
	}, nil
}

func (h *StreamHandler) GetActiveStream(ctx context.Context, req *pb.GetActiveStreamRequest) (*pb.GetActiveStreamResponse, error) {
	qry := &query.GetActiveStreamQuery{
		BaseQuery: cqrs.BaseQuery{},
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "no active stream")
	}

	streamResult := result.(*query.GetActiveStreamResult)

	return &pb.GetActiveStreamResponse{
		Stream: mapper.StreamToProto(streamResult.Stream),
	}, nil
}

func (h *StreamHandler) GetStreamStats(ctx context.Context, req *pb.IdRequest) (*pb.GetStreamStatsResponse, error) {
	qry := &query.GetStreamStatsQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "stream not found")
	}

	statsResult := result.(*query.GetStreamStatsResult)

	return &pb.GetStreamStatsResponse{
		PeakViewers:     int32(statsResult.PeakViewers),
		AverageViewers:  int32(statsResult.AverageViewers),
		TotalMessages:   int32(statsResult.TotalMessages),
		UniqueViewers:   int32(statsResult.UniqueViewers),
		DurationSeconds: statsResult.DurationSeconds,
	}, nil
}
