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

type ViewerHandler struct {
	pb.UnimplementedViewerServiceServer
	commandBus *cqrs.CommandBus
	queryBus   *cqrs.QueryBus
}

func NewViewerHandler(commandBus *cqrs.CommandBus, queryBus *cqrs.QueryBus) *ViewerHandler {
	return &ViewerHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
	}
}

func (h *ViewerHandler) CreateViewer(ctx context.Context, req *pb.CreateViewerRequest) (*pb.CreateViewerResponse, error) {
	cmd := &command.CreateViewerCommand{
		BaseCommand: cqrs.BaseCommand{},
		TwitchID:    req.TwitchId,
		Username:    req.Username,
		DisplayName: req.DisplayName,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Fetch the created viewer
	qry := &query.GetViewerByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        cmd.AggregateID,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	viewerResult := result.(*query.GetViewerResult)

	return &pb.CreateViewerResponse{
		Viewer: mapper.ViewerToProto(viewerResult.Viewer),
	}, nil
}

func (h *ViewerHandler) GetViewer(ctx context.Context, req *pb.IdRequest) (*pb.GetViewerResponse, error) {
	qry := &query.GetViewerByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "viewer not found")
	}

	viewerResult := result.(*query.GetViewerResult)

	return &pb.GetViewerResponse{
		Viewer: mapper.ViewerToProto(viewerResult.Viewer),
	}, nil
}

func (h *ViewerHandler) ListViewers(ctx context.Context, req *pb.ListViewersRequest) (*pb.ListViewersResponse, error) {
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

	qry := &query.ListViewersQuery{
		BaseQuery:      cqrs.BaseQuery{},
		Page:           page,
		PageSize:       pageSize,
		OrderBy:        req.OrderBy,
		IncludeDeleted: includeDeleted,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	listResult := result.(*query.ListViewersResult)

	return &pb.ListViewersResponse{
		Viewers: mapper.ViewersToProto(listResult.Viewers),
		Total:   int32(listResult.Total),
	}, nil
}

func (h *ViewerHandler) UpdateViewer(ctx context.Context, req *pb.UpdateViewerRequest) (*pb.UpdateViewerResponse, error) {
	var username, displayName *string
	var totalMessages, totalStreamsWatched *int

	if req.Username != nil {
		username = req.Username
	}
	if req.DisplayName != nil {
		displayName = req.DisplayName
	}
	if req.TotalMessages != nil {
		tm := int(*req.TotalMessages)
		totalMessages = &tm
	}
	if req.TotalStreamsWatched != nil {
		tsw := int(*req.TotalStreamsWatched)
		totalStreamsWatched = &tsw
	}

	cmd := &command.UpdateViewerCommand{
		BaseCommand:         cqrs.BaseCommand{AggregateID: req.Id},
		Username:            username,
		DisplayName:         displayName,
		TotalMessages:       totalMessages,
		TotalStreamsWatched: totalStreamsWatched,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Fetch the updated viewer
	qry := &query.GetViewerByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "viewer not found")
	}

	viewerResult := result.(*query.GetViewerResult)

	return &pb.UpdateViewerResponse{
		Viewer: mapper.ViewerToProto(viewerResult.Viewer),
	}, nil
}

func (h *ViewerHandler) DeleteViewer(ctx context.Context, req *pb.IdRequest) (*pb.DeleteResponse, error) {
	cmd := &command.DeleteViewerCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Viewer deleted successfully",
	}, nil
}

func (h *ViewerHandler) GetViewerByTwitchId(ctx context.Context, req *pb.GetViewerByTwitchIdRequest) (*pb.GetViewerResponse, error) {
	qry := &query.GetViewerByTwitchIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		TwitchID:  req.TwitchId,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "viewer not found")
	}

	viewerResult := result.(*query.GetViewerResult)

	return &pb.GetViewerResponse{
		Viewer: mapper.ViewerToProto(viewerResult.Viewer),
	}, nil
}

func (h *ViewerHandler) GetViewerStats(ctx context.Context, req *pb.IdRequest) (*pb.GetViewerStatsResponse, error) {
	qry := &query.GetViewerStatsQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "viewer not found")
	}

	statsResult := result.(*query.GetViewerStatsResult)

	return &pb.GetViewerStatsResponse{
		TotalMessages:       int32(statsResult.TotalMessages),
		TotalStreamsWatched: int32(statsResult.TotalStreamsWatched),
		DaysSinceFirstSeen:  int32(statsResult.DaysSinceFirstSeen),
		DaysSinceLastSeen:   int32(statsResult.DaysSinceLastSeen),
	}, nil
}
