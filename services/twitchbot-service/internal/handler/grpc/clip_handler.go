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

type ClipHandler struct {
	pb.UnimplementedClipServiceServer
	commandBus *cqrs.CommandBus
	queryBus   *cqrs.QueryBus
}

func NewClipHandler(commandBus *cqrs.CommandBus, queryBus *cqrs.QueryBus) *ClipHandler {
	return &ClipHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
	}
}

func (h *ClipHandler) CreateClip(ctx context.Context, req *pb.CreateClipRequest) (*pb.CreateClipResponse, error) {
	cmd := &command.CreateClipCommand{
		BaseCommand:    cqrs.BaseCommand{},
		StreamID:       req.StreamId,
		TwitchClipID:   req.TwitchClipId,
		Title:          req.Title,
		URL:            req.Url,
		EmbedURL:       req.EmbedUrl,
		ThumbnailURL:   req.ThumbnailUrl,
		CreatorName:    req.CreatorName,
		CreatorID:      req.CreatorId,
		ViewCount:      int(req.ViewCount),
		DurationSeconds: int(req.DurationSeconds),
		CreatedAtTwitch: mapper.ProtoToTime(req.CreatedAtTwitch),
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Fetch the created clip
	qry := &query.GetClipByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        cmd.AggregateID,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	clipResult := result.(*query.GetClipResult)

	return &pb.CreateClipResponse{
		Clip: mapper.ClipToProto(clipResult.Clip),
	}, nil
}

func (h *ClipHandler) GetClip(ctx context.Context, req *pb.IdRequest) (*pb.GetClipResponse, error) {
	qry := &query.GetClipByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "clip not found")
	}

	clipResult := result.(*query.GetClipResult)

	return &pb.GetClipResponse{
		Clip: mapper.ClipToProto(clipResult.Clip),
	}, nil
}

func (h *ClipHandler) ListClips(ctx context.Context, req *pb.ListClipsRequest) (*pb.ListClipsResponse, error) {
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

	qry := &query.ListClipsQuery{
		BaseQuery:      cqrs.BaseQuery{},
		Page:           page,
		PageSize:       pageSize,
		StreamID:       req.StreamId,
		OrderBy:        req.OrderBy,
		IncludeDeleted: includeDeleted,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	listResult := result.(*query.ListClipsResult)

	return &pb.ListClipsResponse{
		Clips: mapper.ClipsToProto(listResult.Clips),
		Total: int32(listResult.Total),
	}, nil
}

func (h *ClipHandler) UpdateClip(ctx context.Context, req *pb.UpdateClipRequest) (*pb.UpdateClipResponse, error) {
	var title *string
	var viewCount *int

	if req.Title != nil {
		title = req.Title
	}
	if req.ViewCount != nil {
		vc := int(*req.ViewCount)
		viewCount = &vc
	}

	cmd := &command.UpdateClipCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
		Title:       title,
		ViewCount:   viewCount,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Fetch the updated clip
	qry := &query.GetClipByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "clip not found")
	}

	clipResult := result.(*query.GetClipResult)

	return &pb.UpdateClipResponse{
		Clip: mapper.ClipToProto(clipResult.Clip),
	}, nil
}

func (h *ClipHandler) DeleteClip(ctx context.Context, req *pb.IdRequest) (*pb.DeleteResponse, error) {
	cmd := &command.DeleteClipCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Clip deleted successfully",
	}, nil
}

func (h *ClipHandler) GetClipByTwitchId(ctx context.Context, req *pb.GetClipByTwitchIdRequest) (*pb.GetClipResponse, error) {
	qry := &query.GetClipByTwitchClipIDQuery{
		BaseQuery:    cqrs.BaseQuery{},
		TwitchClipID: req.TwitchClipId,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "clip not found")
	}

	clipResult := result.(*query.GetClipResult)

	return &pb.GetClipResponse{
		Clip: mapper.ClipToProto(clipResult.Clip),
	}, nil
}
