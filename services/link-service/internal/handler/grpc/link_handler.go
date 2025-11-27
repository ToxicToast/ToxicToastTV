package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"

	pb "toxictoast/services/link-service/api/proto"
	"toxictoast/services/link-service/internal/command"
	"toxictoast/services/link-service/internal/domain"
	"toxictoast/services/link-service/internal/handler/mapper"
	"toxictoast/services/link-service/internal/query"
	"toxictoast/services/link-service/internal/repository"
	"toxictoast/services/link-service/pkg/config"
)

type LinkHandler struct {
	pb.UnimplementedLinkServiceServer
	commandBus *cqrs.CommandBus
	queryBus   *cqrs.QueryBus
	config     *config.Config
}

// NewLinkHandler creates a new instance of LinkHandler
func NewLinkHandler(commandBus *cqrs.CommandBus, queryBus *cqrs.QueryBus, cfg *config.Config) *LinkHandler {
	return &LinkHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
		config:     cfg,
	}
}

func (h *LinkHandler) CreateLink(ctx context.Context, req *pb.CreateLinkRequest) (*pb.CreateLinkResponse, error) {
	// Create command
	cmd := &command.CreateLinkCommand{
		BaseCommand: cqrs.BaseCommand{},
		OriginalURL: req.OriginalUrl,
		CustomAlias: req.CustomAlias,
		Title:       req.Title,
		Description: req.Description,
		ExpiresAt:   mapper.ProtoToTime(req.ExpiresAt),
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create link: %v", err)
	}

	// Query created link
	getQuery := &query.GetLinkByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		LinkID:    cmd.AggregateID,
	}

	result, err := h.queryBus.Dispatch(ctx, getQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve created link: %v", err)
	}

	link := result.(*domain.Link)

	return &pb.CreateLinkResponse{
		Link:     mapper.LinkToProto(link),
		ShortUrl: cmd.ShortURL,
	}, nil
}

func (h *LinkHandler) GetLink(ctx context.Context, req *pb.GetLinkRequest) (*pb.GetLinkResponse, error) {
	// Query link
	getQuery := &query.GetLinkByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		LinkID:    req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, getQuery)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "link not found: %v", err)
	}

	link := result.(*domain.Link)

	return &pb.GetLinkResponse{
		Link: mapper.LinkToProto(link),
	}, nil
}

func (h *LinkHandler) GetLinkByShortCode(ctx context.Context, req *pb.GetLinkByShortCodeRequest) (*pb.GetLinkResponse, error) {
	// Query link by short code
	getQuery := &query.GetLinkByShortCodeQuery{
		BaseQuery: cqrs.BaseQuery{},
		ShortCode: req.ShortCode,
	}

	result, err := h.queryBus.Dispatch(ctx, getQuery)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "link not found: %v", err)
	}

	link := result.(*domain.Link)

	return &pb.GetLinkResponse{
		Link: mapper.LinkToProto(link),
	}, nil
}

func (h *LinkHandler) ListLinks(ctx context.Context, req *pb.ListLinksRequest) (*pb.ListLinksResponse, error) {
	// Convert request to filters
	filters := repository.LinkFilters{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
		IsActive: req.IsActive,
	}

	// Default pagination
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 {
		filters.PageSize = 10
	}

	// Create query
	listQuery := &query.ListLinksQuery{
		BaseQuery: cqrs.BaseQuery{},
		Filters:   filters,
	}

	// Dispatch query
	result, err := h.queryBus.Dispatch(ctx, listQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list links: %v", err)
	}

	listResult := result.(*query.ListLinksResult)

	// Convert to proto
	protoLinks := make([]*pb.Link, len(listResult.Links))
	for i, link := range listResult.Links {
		protoLinks[i] = mapper.LinkToProto(&link)
	}

	return &pb.ListLinksResponse{
		Links: protoLinks,
		Total: int32(listResult.Total),
	}, nil
}

func (h *LinkHandler) UpdateLink(ctx context.Context, req *pb.UpdateLinkRequest) (*pb.UpdateLinkResponse, error) {
	// Create command
	cmd := &command.UpdateLinkCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
		OriginalURL: req.OriginalUrl,
		CustomAlias: req.CustomAlias,
		Title:       req.Title,
		Description: req.Description,
		ExpiresAt:   mapper.ProtoToTime(req.ExpiresAt),
		IsActive:    req.IsActive,
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update link: %v", err)
	}

	// Query updated link
	getQuery := &query.GetLinkByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		LinkID:    req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, getQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve updated link: %v", err)
	}

	link := result.(*domain.Link)

	return &pb.UpdateLinkResponse{
		Link: mapper.LinkToProto(link),
	}, nil
}

func (h *LinkHandler) DeleteLink(ctx context.Context, req *pb.DeleteLinkRequest) (*pb.DeleteResponse, error) {
	// Create command
	cmd := &command.DeleteLinkCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete link: %v", err)
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Link deleted successfully",
	}, nil
}

func (h *LinkHandler) IncrementClick(ctx context.Context, req *pb.IncrementClickRequest) (*pb.IncrementClickResponse, error) {
	// Create command
	cmd := &command.IncrementClickCommand{
		BaseCommand: cqrs.BaseCommand{},
		ShortCode:   req.ShortCode,
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to increment click: %v", err)
	}

	return &pb.IncrementClickResponse{
		ClickCount: int32(cmd.NewCount),
	}, nil
}

func (h *LinkHandler) GetLinkStats(ctx context.Context, req *pb.GetLinkStatsRequest) (*pb.GetLinkStatsResponse, error) {
	// Query link stats
	statsQuery := &query.GetLinkStatsQuery{
		BaseQuery: cqrs.BaseQuery{},
		LinkID:    req.LinkId,
	}

	result, err := h.queryBus.Dispatch(ctx, statsQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get link stats: %v", err)
	}

	stats := result.(*repository.LinkStats)

	// Get analytics for detailed stats (optional, if needed)
	analyticsQuery := &query.GetLinkAnalyticsQuery{
		BaseQuery: cqrs.BaseQuery{},
		LinkID:    req.LinkId,
	}

	analyticsResult, _ := h.queryBus.Dispatch(ctx, analyticsQuery)

	var clicksByCountry map[string]int32
	var clicksByDevice map[string]int32
	var topReferers []string

	if analyticsResult != nil {
		clickStats := analyticsResult.(*repository.ClickStats)
		clicksByCountry = make(map[string]int32)
		for k, v := range clickStats.ClicksByCountry {
			clicksByCountry[k] = int32(v)
		}
		clicksByDevice = make(map[string]int32)
		for k, v := range clickStats.ClicksByDevice {
			clicksByDevice[k] = int32(v)
		}
		topReferers = clickStats.TopReferers
	}

	return &pb.GetLinkStatsResponse{
		LinkId:            stats.LinkID,
		TotalClicks:       int32(stats.TotalClicks),
		UniqueIps:         int32(stats.UniqueIPs),
		ClicksToday:       int32(stats.ClicksToday),
		ClicksThisWeek:    int32(stats.ClicksWeek),
		ClicksThisMonth:   int32(stats.ClicksMonth),
		ClicksByCountry:   clicksByCountry,
		ClicksByDevice:    clicksByDevice,
		TopReferers:       topReferers,
	}, nil
}

func (h *LinkHandler) RecordClick(ctx context.Context, req *pb.RecordClickRequest) (*pb.RecordClickResponse, error) {
	// Create command
	cmd := &command.RecordClickCommand{
		BaseCommand: cqrs.BaseCommand{},
		LinkID:      req.LinkId,
		IPAddress:   req.IpAddress,
		UserAgent:   req.UserAgent,
		Referer:     req.Referer,
		Country:     req.Country,
		City:        req.City,
		DeviceType:  req.DeviceType,
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to record click: %v", err)
	}

	// Create a simple Click response (we don't need to query it back)
	click := &domain.Click{
		ID:         cmd.AggregateID,
		LinkID:     req.LinkId,
		IPAddress:  req.IpAddress,
		UserAgent:  req.UserAgent,
		Referer:    req.Referer,
		Country:    req.Country,
		City:       req.City,
		DeviceType: req.DeviceType,
	}

	return &pb.RecordClickResponse{
		Click: mapper.ClickToProto(click),
	}, nil
}

func (h *LinkHandler) GetLinkClicks(ctx context.Context, req *pb.GetLinkClicksRequest) (*pb.GetLinkClicksResponse, error) {
	// Create query
	clicksQuery := &query.GetLinkClicksQuery{
		BaseQuery: cqrs.BaseQuery{},
		LinkID:    req.LinkId,
		Page:      int(req.Page),
		PageSize:  int(req.PageSize),
		StartDate: mapper.ProtoToTime(req.StartDate),
		EndDate:   mapper.ProtoToTime(req.EndDate),
	}

	// Dispatch query
	result, err := h.queryBus.Dispatch(ctx, clicksQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get link clicks: %v", err)
	}

	clicksResult := result.(*query.GetLinkClicksResult)

	// Convert to proto
	protoClicks := make([]*pb.Click, len(clicksResult.Clicks))
	for i, click := range clicksResult.Clicks {
		protoClicks[i] = mapper.ClickToProto(&click)
	}

	return &pb.GetLinkClicksResponse{
		Clicks: protoClicks,
		Total:  int32(clicksResult.Total),
	}, nil
}

func (h *LinkHandler) GetClicksByDate(ctx context.Context, req *pb.GetClicksByDateRequest) (*pb.GetClicksByDateResponse, error) {
	// Validate dates
	startDate := mapper.ProtoToTime(req.StartDate)
	endDate := mapper.ProtoToTime(req.EndDate)

	if startDate == nil || endDate == nil {
		return nil, status.Error(codes.InvalidArgument, "start_date and end_date are required")
	}

	// Create query
	clicksByDateQuery := &query.GetClicksByDateQuery{
		BaseQuery: cqrs.BaseQuery{},
		LinkID:    req.LinkId,
		StartDate: *startDate,
		EndDate:   *endDate,
	}

	// Dispatch query
	result, err := h.queryBus.Dispatch(ctx, clicksByDateQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get clicks by date: %v", err)
	}

	clicksByDateResult := result.(*query.GetClicksByDateResult)

	// Convert to proto format (repeated ClicksByDate)
	data := make([]*pb.ClicksByDate, 0, len(clicksByDateResult.ClicksByDate))
	for date, count := range clicksByDateResult.ClicksByDate {
		data = append(data, &pb.ClicksByDate{
			Date:   date,
			Clicks: int32(count),
		})
	}

	return &pb.GetClicksByDateResponse{
		Data:        data,
		TotalClicks: int32(clicksByDateResult.TotalClicks),
	}, nil
}
