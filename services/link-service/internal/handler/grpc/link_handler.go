package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "toxictoast/services/link-service/api/proto"
	"toxictoast/services/link-service/internal/handler/mapper"
	"toxictoast/services/link-service/internal/repository"
	"toxictoast/services/link-service/internal/usecase"
)

type LinkHandler struct {
	pb.UnimplementedLinkServiceServer
	linkUseCase  usecase.LinkUseCase
	clickUseCase usecase.ClickUseCase
}

// NewLinkHandler creates a new instance of LinkHandler
func NewLinkHandler(linkUseCase usecase.LinkUseCase, clickUseCase usecase.ClickUseCase) *LinkHandler {
	return &LinkHandler{
		linkUseCase:  linkUseCase,
		clickUseCase: clickUseCase,
	}
}

func (h *LinkHandler) CreateLink(ctx context.Context, req *pb.CreateLinkRequest) (*pb.CreateLinkResponse, error) {
	// Convert request to use case input
	input := usecase.CreateLinkInput{
		OriginalURL: req.OriginalUrl,
		CustomAlias: req.CustomAlias,
		Title:       req.Title,
		Description: req.Description,
		ExpiresAt:   mapper.ProtoToTime(req.ExpiresAt),
	}

	// Create link
	link, shortURL, err := h.linkUseCase.CreateLink(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create link: %v", err)
	}

	return &pb.CreateLinkResponse{
		Link:     mapper.LinkToProto(link),
		ShortUrl: shortURL,
	}, nil
}

func (h *LinkHandler) GetLink(ctx context.Context, req *pb.GetLinkRequest) (*pb.GetLinkResponse, error) {
	link, err := h.linkUseCase.GetLink(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "link not found: %v", err)
	}

	return &pb.GetLinkResponse{
		Link: mapper.LinkToProto(link),
	}, nil
}

func (h *LinkHandler) GetLinkByShortCode(ctx context.Context, req *pb.GetLinkByShortCodeRequest) (*pb.GetLinkResponse, error) {
	link, err := h.linkUseCase.GetLinkByShortCode(ctx, req.ShortCode)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "link not found: %v", err)
	}

	return &pb.GetLinkResponse{
		Link: mapper.LinkToProto(link),
	}, nil
}

func (h *LinkHandler) ListLinks(ctx context.Context, req *pb.ListLinksRequest) (*pb.ListLinksResponse, error) {
	// Get default pagination
	page, pageSize := mapper.GetDefaultPagination(req.Page, req.PageSize)

	// Convert request to filters
	filters := repository.LinkFilters{
		Page:           int(page),
		PageSize:       int(pageSize),
		IsActive:       req.IsActive,
		IncludeExpired: mapper.BoolValue(req.IncludeExpired),
		Search:         req.Search,
		SortBy:         "created_at",
		SortOrder:      "DESC",
	}

	// List links
	links, total, err := h.linkUseCase.ListLinks(ctx, filters)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list links: %v", err)
	}

	// Calculate total pages
	totalPages := int32(total) / pageSize
	if int32(total)%pageSize > 0 {
		totalPages++
	}

	return &pb.ListLinksResponse{
		Links:      mapper.LinksToProto(links),
		Total:      int32(total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (h *LinkHandler) UpdateLink(ctx context.Context, req *pb.UpdateLinkRequest) (*pb.UpdateLinkResponse, error) {
	// Convert request to use case input
	input := usecase.UpdateLinkInput{
		OriginalURL: req.OriginalUrl,
		CustomAlias: req.CustomAlias,
		Title:       req.Title,
		Description: req.Description,
		ExpiresAt:   mapper.ProtoToTime(req.ExpiresAt),
		IsActive:    req.IsActive,
	}

	// Update link
	link, err := h.linkUseCase.UpdateLink(ctx, req.Id, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update link: %v", err)
	}

	return &pb.UpdateLinkResponse{
		Link: mapper.LinkToProto(link),
	}, nil
}

func (h *LinkHandler) DeleteLink(ctx context.Context, req *pb.DeleteLinkRequest) (*pb.DeleteResponse, error) {
	if err := h.linkUseCase.DeleteLink(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete link: %v", err)
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Link deleted successfully",
	}, nil
}

func (h *LinkHandler) IncrementClick(ctx context.Context, req *pb.IncrementClickRequest) (*pb.IncrementClickResponse, error) {
	clickCount, err := h.linkUseCase.IncrementClick(ctx, req.ShortCode)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to increment click: %v", err)
	}

	return &pb.IncrementClickResponse{
		ClickCount: int32(clickCount),
	}, nil
}

func (h *LinkHandler) GetLinkStats(ctx context.Context, req *pb.GetLinkStatsRequest) (*pb.GetLinkStatsResponse, error) {
	stats, err := h.clickUseCase.GetLinkAnalytics(ctx, req.LinkId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get link stats: %v", err)
	}

	// Convert map[string]int to map[string]int32
	clicksByCountry := make(map[string]int32)
	for k, v := range stats.ClicksByCountry {
		clicksByCountry[k] = int32(v)
	}

	clicksByDevice := make(map[string]int32)
	for k, v := range stats.ClicksByDevice {
		clicksByDevice[k] = int32(v)
	}

	return &pb.GetLinkStatsResponse{
		LinkId:           stats.LinkID,
		TotalClicks:      int32(stats.TotalClicks),
		UniqueIps:        int32(stats.UniqueIPs),
		ClicksToday:      int32(stats.ClicksToday),
		ClicksThisWeek:   int32(stats.ClicksThisWeek),
		ClicksThisMonth:  int32(stats.ClicksThisMonth),
		ClicksByCountry:  clicksByCountry,
		ClicksByDevice:   clicksByDevice,
		TopReferers:      stats.TopReferers,
	}, nil
}

func (h *LinkHandler) RecordClick(ctx context.Context, req *pb.RecordClickRequest) (*pb.RecordClickResponse, error) {
	// Convert request to use case input
	input := usecase.RecordClickInput{
		LinkID:     req.LinkId,
		IPAddress:  req.IpAddress,
		UserAgent:  req.UserAgent,
		Referer:    req.Referer,
		Country:    req.Country,
		City:       req.City,
		DeviceType: req.DeviceType,
	}

	// Record click
	click, err := h.clickUseCase.RecordClick(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to record click: %v", err)
	}

	return &pb.RecordClickResponse{
		Click: mapper.ClickToProto(click),
	}, nil
}

func (h *LinkHandler) GetLinkClicks(ctx context.Context, req *pb.GetLinkClicksRequest) (*pb.GetLinkClicksResponse, error) {
	// Get default pagination
	page, pageSize := mapper.GetDefaultPagination(req.Page, req.PageSize)

	// Get clicks
	clicks, total, err := h.clickUseCase.GetLinkClicks(
		ctx,
		req.LinkId,
		int(page),
		int(pageSize),
		mapper.ProtoToTime(req.StartDate),
		mapper.ProtoToTime(req.EndDate),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get link clicks: %v", err)
	}

	// Calculate total pages
	totalPages := int32(total) / pageSize
	if int32(total)%pageSize > 0 {
		totalPages++
	}

	return &pb.GetLinkClicksResponse{
		Clicks:     mapper.ClicksToProto(clicks),
		Total:      int32(total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (h *LinkHandler) GetClicksByDate(ctx context.Context, req *pb.GetClicksByDateRequest) (*pb.GetClicksByDateResponse, error) {
	startDate := req.StartDate.AsTime()
	endDate := req.EndDate.AsTime()

	clicksByDate, totalClicks, err := h.clickUseCase.GetClicksByDate(ctx, req.LinkId, startDate, endDate)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get clicks by date: %v", err)
	}

	// Convert to proto
	data := make([]*pb.ClicksByDate, 0, len(clicksByDate))
	for date, clicks := range clicksByDate {
		data = append(data, &pb.ClicksByDate{
			Date:   date,
			Clicks: int32(clicks),
		})
	}

	return &pb.GetClicksByDateResponse{
		Data:        data,
		TotalClicks: int32(totalClicks),
	}, nil
}
