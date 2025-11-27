package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/toxictoast/toxictoastgo/shared/auth"
	"github.com/toxictoast/toxictoastgo/shared/cqrs"

	pb "toxictoast/services/blog-service/api/proto"
	"toxictoast/services/blog-service/internal/command"
	"toxictoast/services/blog-service/internal/domain"
	"toxictoast/services/blog-service/internal/query"
	"toxictoast/services/blog-service/internal/repository"
)

type TagHandler struct {
	pb.UnimplementedBlogServiceServer
	commandBus  *cqrs.CommandBus
	queryBus    *cqrs.QueryBus
	authEnabled bool
}

func NewTagHandler(commandBus *cqrs.CommandBus, queryBus *cqrs.QueryBus, authEnabled bool) *TagHandler {
	return &TagHandler{
		commandBus:  commandBus,
		queryBus:    queryBus,
		authEnabled: authEnabled,
	}
}

func (h *TagHandler) CreateTag(ctx context.Context, req *pb.CreateTagRequest) (*pb.TagResponse, error) {
	// Get user from context (authenticated by middleware)
	_, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	// Create command
	cmd := &command.CreateTagCommand{
		BaseCommand: cqrs.BaseCommand{},
		Name:        req.Name,
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create tag: %v", err)
	}

	// Query created tag
	getQuery := &query.GetTagByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		TagID:     cmd.AggregateID,
	}

	result, err := h.queryBus.Dispatch(ctx, getQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve created tag: %v", err)
	}

	tag := result.(*domain.Tag)

	return &pb.TagResponse{
		Tag: domainTagToProto(tag),
	}, nil
}

func (h *TagHandler) GetTag(ctx context.Context, req *pb.GetTagRequest) (*pb.TagResponse, error) {
	var tag *domain.Tag
	var err error
	var result interface{}

	switch identifier := req.Identifier.(type) {
	case *pb.GetTagRequest_Id:
		getQuery := &query.GetTagByIDQuery{
			BaseQuery: cqrs.BaseQuery{},
			TagID:     identifier.Id,
		}
		result, err = h.queryBus.Dispatch(ctx, getQuery)
		if err == nil {
			tag = result.(*domain.Tag)
		}

	case *pb.GetTagRequest_Slug:
		getQuery := &query.GetTagBySlugQuery{
			BaseQuery: cqrs.BaseQuery{},
			Slug:      identifier.Slug,
		}
		result, err = h.queryBus.Dispatch(ctx, getQuery)
		if err == nil {
			tag = result.(*domain.Tag)
		}

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

	// Create command
	cmd := &command.UpdateTagCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
		Name:        req.Name,
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update tag: %v", err)
	}

	// Query updated tag
	getQuery := &query.GetTagByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		TagID:     req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, getQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve updated tag: %v", err)
	}

	tag := result.(*domain.Tag)

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

	// Create command
	cmd := &command.DeleteTagCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
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

	// Create query
	listQuery := &query.ListTagsQuery{
		BaseQuery: cqrs.BaseQuery{},
		Filters:   filters,
	}

	// Dispatch query
	result, err := h.queryBus.Dispatch(ctx, listQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list tags: %v", err)
	}

	listResult := result.(*query.ListTagsResult)

	// Convert to proto
	protoTags := make([]*pb.Tag, len(listResult.Tags))
	for i, tag := range listResult.Tags {
		protoTags[i] = domainTagToProto(&tag)
	}

	return &pb.ListTagsResponse{
		Tags:  protoTags,
		Total: int32(listResult.Total),
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
