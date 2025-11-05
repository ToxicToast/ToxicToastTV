package grpc

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/toxictoast/toxictoastgo/shared/auth"

	pb "toxictoast/services/blog-service/api/proto"
	"toxictoast/services/blog-service/internal/domain"
	"toxictoast/services/blog-service/internal/repository"
	"toxictoast/services/blog-service/internal/usecase"
)

type MediaHandler struct {
	pb.UnimplementedBlogServiceServer
	mediaUseCase usecase.MediaUseCase
	authEnabled  bool
}

func NewMediaHandler(mediaUseCase usecase.MediaUseCase, authEnabled bool) *MediaHandler {
	return &MediaHandler{
		mediaUseCase: mediaUseCase,
		authEnabled:  authEnabled,
	}
}

func (h *MediaHandler) UploadMedia(stream pb.BlogService_UploadMediaServer) error {
	// Get user from context (authenticated by middleware)
	user, err := h.requireAuth(stream.Context())
	if err != nil {
		return err
	}

	var fileData bytes.Buffer
	var metadata *pb.MediaMetadata

	// Receive chunks from stream
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Errorf(codes.Internal, "failed to receive chunk: %v", err)
		}

		switch data := req.Data.(type) {
		case *pb.UploadMediaRequest_Metadata:
			// First message should contain metadata
			if metadata != nil {
				return status.Error(codes.InvalidArgument, "metadata already received")
			}
			metadata = data.Metadata

		case *pb.UploadMediaRequest_Chunk:
			// Subsequent messages contain file chunks
			if metadata == nil {
				return status.Error(codes.InvalidArgument, "metadata must be sent first")
			}
			if _, err := fileData.Write(data.Chunk); err != nil {
				return status.Errorf(codes.Internal, "failed to write chunk: %v", err)
			}
		}
	}

	// Validate that we received metadata
	if metadata == nil {
		return status.Error(codes.InvalidArgument, "no metadata received")
	}

	// Create media via use case
	input := usecase.UploadMediaInput{
		Data:             fileData.Bytes(),
		OriginalFilename: metadata.Filename,
		MimeType:         metadata.MimeType,
		UploadedBy:       user.UserID,
	}

	media, err := h.mediaUseCase.UploadMedia(stream.Context(), input)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to upload media: %v", err)
	}

	// Send response
	response := &pb.MediaResponse{
		Media: domainMediaToProto(media),
	}

	if err := stream.SendAndClose(response); err != nil {
		return status.Errorf(codes.Internal, "failed to send response: %v", err)
	}

	return nil
}

func (h *MediaHandler) GetMedia(ctx context.Context, req *pb.GetMediaRequest) (*pb.MediaResponse, error) {
	media, err := h.mediaUseCase.GetMedia(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "media not found: %v", err)
	}

	return &pb.MediaResponse{
		Media: domainMediaToProto(media),
	}, nil
}

func (h *MediaHandler) DeleteMedia(ctx context.Context, req *pb.DeleteMediaRequest) (*pb.DeleteResponse, error) {
	// Require authentication for deletion
	_, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	// Delete media
	if err := h.mediaUseCase.DeleteMedia(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete media: %v", err)
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Media deleted successfully",
	}, nil
}

func (h *MediaHandler) ListMedia(ctx context.Context, req *pb.ListMediaRequest) (*pb.ListMediaResponse, error) {
	// Convert request to filters
	filters := repository.MediaFilters{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
		MimeType: stringPtrFromOptional(req.MimeType),
	}

	// Default pagination
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 {
		filters.PageSize = 20
	}

	// List media
	mediaList, total, err := h.mediaUseCase.ListMedia(ctx, filters)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list media: %v", err)
	}

	// Convert to proto
	protoMedia := make([]*pb.Media, len(mediaList))
	for i, media := range mediaList {
		protoMedia[i] = domainMediaToProto(&media)
	}

	return &pb.ListMediaResponse{
		Media: protoMedia,
		Total: int32(total),
	}, nil
}

// Helper functions for conversion

func domainMediaToProto(media *domain.Media) *pb.Media {
	if media == nil {
		return nil
	}

	protoMedia := &pb.Media{
		Id:               media.ID,
		Filename:         media.Filename,
		OriginalFilename: media.OriginalFilename,
		MimeType:         media.MimeType,
		Size:             media.Size,
		Url:              media.URL,
		Width:            int32(media.Width),
		Height:           int32(media.Height),
		UploadedBy:       media.UploadedBy,
		CreatedAt:        timestamppb.New(media.CreatedAt),
	}

	// Add thumbnail URL if exists
	if media.ThumbnailURL != nil {
		protoMedia.ThumbnailUrl = stringPtrToOptional(media.ThumbnailURL)
	}

	return protoMedia
}

// requireAuth checks authentication if enabled, returns user context or error
func (h *MediaHandler) requireAuth(ctx context.Context) (*auth.UserContext, error) {
	if !h.authEnabled {
		// Return a dummy user context when auth is disabled
		return &auth.UserContext{
			UserID:   "test-user",
			Username: "test",
			Email:    "test@example.com",
			Roles:    []string{"admin"},
		}, nil
	}

	user, err := auth.GetUserContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}
	return user, nil
}

// Helper to convert string to bytes for chunk size estimation
func estimateChunkCount(fileSize int64, chunkSize int) int {
	chunks := fileSize / int64(chunkSize)
	if fileSize%int64(chunkSize) != 0 {
		chunks++
	}
	return int(chunks)
}

// FormatFileSize formats bytes to human-readable size
func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
