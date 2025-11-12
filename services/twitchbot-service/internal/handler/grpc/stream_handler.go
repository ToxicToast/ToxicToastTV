package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"toxictoast/services/twitchbot-service/internal/handler/mapper"
	"toxictoast/services/twitchbot-service/internal/usecase"
	pb "toxictoast/services/twitchbot-service/api/proto"
)

type StreamHandler struct {
	pb.UnimplementedStreamServiceServer
	streamUC usecase.StreamUseCase
}

func NewStreamHandler(streamUC usecase.StreamUseCase) *StreamHandler {
	return &StreamHandler{
		streamUC: streamUC,
	}
}

func (h *StreamHandler) CreateStream(ctx context.Context, req *pb.CreateStreamRequest) (*pb.CreateStreamResponse, error) {
	stream, err := h.streamUC.CreateStream(ctx, req.Title, req.GameName, req.GameId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateStreamResponse{
		Stream: mapper.StreamToProto(stream),
	}, nil
}

func (h *StreamHandler) GetStream(ctx context.Context, req *pb.IdRequest) (*pb.GetStreamResponse, error) {
	stream, err := h.streamUC.GetStreamByID(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrStreamNotFound {
			return nil, status.Error(codes.NotFound, "stream not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetStreamResponse{
		Stream: mapper.StreamToProto(stream),
	}, nil
}

func (h *StreamHandler) ListStreams(ctx context.Context, req *pb.ListStreamsRequest) (*pb.ListStreamsResponse, error) {
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

	streams, total, err := h.streamUC.ListStreams(ctx, int(page), int(pageSize), req.OnlyActive, req.GameName, includeDeleted)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ListStreamsResponse{
		Streams: mapper.StreamsToProto(streams),
		Total:   int32(total),
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

	stream, err := h.streamUC.UpdateStream(ctx, req.Id, title, gameName, gameID, peakViewers, averageViewers)
	if err != nil {
		if err == usecase.ErrStreamNotFound {
			return nil, status.Error(codes.NotFound, "stream not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateStreamResponse{
		Stream: mapper.StreamToProto(stream),
	}, nil
}

func (h *StreamHandler) DeleteStream(ctx context.Context, req *pb.IdRequest) (*pb.DeleteResponse, error) {
	if err := h.streamUC.DeleteStream(ctx, req.Id); err != nil {
		if err == usecase.ErrStreamNotFound {
			return nil, status.Error(codes.NotFound, "stream not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Stream deleted successfully",
	}, nil
}

func (h *StreamHandler) EndStream(ctx context.Context, req *pb.EndStreamRequest) (*pb.EndStreamResponse, error) {
	stream, err := h.streamUC.EndStream(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrStreamNotFound {
			return nil, status.Error(codes.NotFound, "stream not found")
		}
		if err == usecase.ErrStreamAlreadyEnded {
			return nil, status.Error(codes.FailedPrecondition, "stream already ended")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.EndStreamResponse{
		Stream: mapper.StreamToProto(stream),
	}, nil
}

func (h *StreamHandler) GetActiveStream(ctx context.Context, req *pb.GetActiveStreamRequest) (*pb.GetActiveStreamResponse, error) {
	stream, err := h.streamUC.GetActiveStream(ctx)
	if err != nil {
		if err == usecase.ErrNoActiveStream {
			return nil, status.Error(codes.NotFound, "no active stream")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetActiveStreamResponse{
		Stream: mapper.StreamToProto(stream),
	}, nil
}

func (h *StreamHandler) GetStreamStats(ctx context.Context, req *pb.IdRequest) (*pb.GetStreamStatsResponse, error) {
	peakViewers, averageViewers, totalMessages, uniqueViewers, durationSeconds, err := h.streamUC.GetStreamStats(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrStreamNotFound {
			return nil, status.Error(codes.NotFound, "stream not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetStreamStatsResponse{
		PeakViewers:     int32(peakViewers),
		AverageViewers:  int32(averageViewers),
		TotalMessages:   int32(totalMessages),
		UniqueViewers:   int32(uniqueViewers),
		DurationSeconds: durationSeconds,
	}, nil
}
