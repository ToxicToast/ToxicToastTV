package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/toxictoast/toxictoastgo/shared/auth"

	pb "toxictoast/services/blog-service/api/proto"
	"toxictoast/services/blog-service/internal/domain"
	"toxictoast/services/blog-service/internal/repository"
	"toxictoast/services/blog-service/internal/usecase"
)

type TagHandler struct {
	pb.UnimplementedBlogServiceServer
	tagUseCase  usecase.TagUseCase
	authEnabled bool
}

func NewTagHandler(tagUseCase usecase.TagUseCase, authEnabled bool) *TagHandler {
	return &TagHandler{
		tagUseCase:  tagUseCase,
		authEnabled: authEnabled,
	}
}

func (h *TagHandler) CreateTag(ctx context.Context, req *pb.CreateTagRequest) (*pb.TagResponse, error) {
	// Get user from context (authenticated by middleware)
	_, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	// Convert request to use case input
	input := usecase.CreateTagInput{
		Name: req.Name,
	}

	// Create tag
	tag, err := h.tagUseCase.CreateTag(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create tag: %v", err)
	}

	return &pb.TagResponse{
		Tag: domainTagToProto(tag),
	}, nil
}

func (h *TagHandler) GetTag(ctx context.Context, req *pb.GetTagRequest) (*pb.TagResponse, error) {
	var tag *domain.Tag
	var err error

	switch identifier := req.Identifier.(type) {
	case *pb.GetTagRequest_Id:
		tag, err = h.tagUseCase.GetTag(ctx, identifier.Id)
	case *pb.GetTagRequest_Slug:
		tag, err = h.tagUseCase.GetTagBySlug(ctx, identifier.Slug)
	default:
		return nil, status.Error(codes.InvalidArgument, "must provide either id or slug")
	}

	if err != nil {
		return nil, status.Errorf(codes.NotFound, "tag not found: %v", err)
	}

	return &pb.TagResponse{
		Tag: domainTagToProto(tag),
	}, nil
}

func (h *TagHandler) UpdateTag(ctx context.Context, req *pb.UpdateTagRequest) (*pb.TagResponse, error) {
	// Get user from context (authenticated by middleware)
	_, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	// Convert request to use case input
	input := usecase.UpdateTagInput{
		Name: req.Name,
	}

	// Update tag
	tag, err := h.tagUseCase.UpdateTag(ctx, req.Id, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update tag: %v", err)
	}

	return &pb.TagResponse{
		Tag: domainTagToProto(tag),
	}, nil
}

func (h *TagHandler) DeleteTag(ctx context.Context, req *pb.DeleteTagRequest) (*pb.DeleteResponse, error) {
	// Get user from context (authenticated by middleware)
	_, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	// Delete tag
	if err := h.tagUseCase.DeleteTag(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete tag: %v", err)
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Tag deleted successfully",
	}, nil
}

func (h *TagHandler) ListTags(ctx context.Context, req *pb.ListTagsRequest) (*pb.ListTagsResponse, error) {
	// Convert request to filters
	filters := repository.TagFilters{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
		Search:   stringPtrFromOptional(req.Search),
	}

	// Default pagination
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 {
		filters.PageSize = 100 // Higher default for tags
	}

	// List tags
	tags, total, err := h.tagUseCase.ListTags(ctx, filters)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list tags: %v", err)
	}

	// Convert to proto
	protoTags := make([]*pb.Tag, len(tags))
	for i, tag := range tags {
		protoTags[i] = domainTagToProto(&tag)
	}

	return &pb.ListTagsResponse{
		Tags:  protoTags,
		Total: int32(total),
	}, nil
}

// Helper functions for conversion

func domainTagToProto(tag *domain.Tag) *pb.Tag {
	if tag == nil {
		return nil
	}

	return &pb.Tag{
		Id:        tag.ID,
		Name:      tag.Name,
		Slug:      tag.Slug,
		CreatedAt: timestamppb.New(tag.CreatedAt),
		UpdatedAt: timestamppb.New(tag.UpdatedAt),
	}
}

// requireAuth checks authentication if enabled, returns user context or error
func (h *TagHandler) requireAuth(ctx context.Context) (*auth.UserContext, error) {
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
