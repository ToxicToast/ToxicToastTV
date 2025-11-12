package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"toxictoast/services/twitchbot-service/internal/handler/mapper"
	"toxictoast/services/twitchbot-service/internal/usecase"
	pb "toxictoast/services/twitchbot-service/api/proto"
)

type ClipHandler struct {
	pb.UnimplementedClipServiceServer
	clipUC usecase.ClipUseCase
}

func NewClipHandler(clipUC usecase.ClipUseCase) *ClipHandler {
	return &ClipHandler{
		clipUC: clipUC,
	}
}

func (h *ClipHandler) CreateClip(ctx context.Context, req *pb.CreateClipRequest) (*pb.CreateClipResponse, error) {
	clip, err := h.clipUC.CreateClip(
		ctx,
		req.StreamId,
		req.TwitchClipId,
		req.Title,
		req.Url,
		req.EmbedUrl,
		req.ThumbnailUrl,
		req.CreatorName,
		req.CreatorId,
		int(req.ViewCount),
		int(req.DurationSeconds),
		mapper.ProtoToTime(req.CreatedAtTwitch),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateClipResponse{
		Clip: mapper.ClipToProto(clip),
	}, nil
}

func (h *ClipHandler) GetClip(ctx context.Context, req *pb.IdRequest) (*pb.GetClipResponse, error) {
	clip, err := h.clipUC.GetClipByID(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrClipNotFound {
			return nil, status.Error(codes.NotFound, "clip not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetClipResponse{
		Clip: mapper.ClipToProto(clip),
	}, nil
}

func (h *ClipHandler) ListClips(ctx context.Context, req *pb.ListClipsRequest) (*pb.ListClipsResponse, error) {
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

	clips, total, err := h.clipUC.ListClips(ctx, int(page), int(pageSize), req.StreamId, req.OrderBy, includeDeleted)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ListClipsResponse{
		Clips: mapper.ClipsToProto(clips),
		Total: int32(total),
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

	clip, err := h.clipUC.UpdateClip(ctx, req.Id, title, viewCount)
	if err != nil {
		if err == usecase.ErrClipNotFound {
			return nil, status.Error(codes.NotFound, "clip not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateClipResponse{
		Clip: mapper.ClipToProto(clip),
	}, nil
}

func (h *ClipHandler) DeleteClip(ctx context.Context, req *pb.IdRequest) (*pb.DeleteResponse, error) {
	if err := h.clipUC.DeleteClip(ctx, req.Id); err != nil {
		if err == usecase.ErrClipNotFound {
			return nil, status.Error(codes.NotFound, "clip not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Clip deleted successfully",
	}, nil
}

func (h *ClipHandler) GetClipByTwitchId(ctx context.Context, req *pb.GetClipByTwitchIdRequest) (*pb.GetClipResponse, error) {
	clip, err := h.clipUC.GetClipByTwitchClipID(ctx, req.TwitchClipId)
	if err != nil {
		if err == usecase.ErrClipNotFound {
			return nil, status.Error(codes.NotFound, "clip not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetClipResponse{
		Clip: mapper.ClipToProto(clip),
	}, nil
}
