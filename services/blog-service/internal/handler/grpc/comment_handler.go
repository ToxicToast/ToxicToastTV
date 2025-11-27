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

type CommentHandler struct {
	pb.UnimplementedBlogServiceServer
	commandBus  *cqrs.CommandBus
	queryBus    *cqrs.QueryBus
	authEnabled bool
}

func NewCommentHandler(commandBus *cqrs.CommandBus, queryBus *cqrs.QueryBus, authEnabled bool) *CommentHandler {
	return &CommentHandler{
		commandBus:  commandBus,
		queryBus:    queryBus,
		authEnabled: authEnabled,
	}
}

func (h *CommentHandler) CreateComment(ctx context.Context, req *pb.CreateCommentRequest) (*pb.CommentResponse, error) {
	// Comments can be created by anyone (public endpoint)

	// Create command
	cmd := &command.CreateCommentCommand{
		BaseCommand: cqrs.BaseCommand{},
		PostID:      req.PostId,
		ParentID:    stringPtrFromOptional(req.ParentId),
		AuthorName:  req.AuthorName,
		AuthorEmail: req.AuthorEmail,
		Content:     req.Content,
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create comment: %v", err)
	}

	// Query created comment
	getQuery := &query.GetCommentByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		CommentID: cmd.AggregateID,
	}

	result, err := h.queryBus.Dispatch(ctx, getQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve created comment: %v", err)
	}

	comment := result.(*domain.Comment)

	return &pb.CommentResponse{
		Comment: domainCommentToProto(comment),
	}, nil
}

func (h *CommentHandler) GetComment(ctx context.Context, req *pb.GetCommentRequest) (*pb.CommentResponse, error) {
	// Query comment
	getQuery := &query.GetCommentByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		CommentID: req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, getQuery)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "comment not found: %v", err)
	}

	comment := result.(*domain.Comment)

	return &pb.CommentResponse{
		Comment: domainCommentToProto(comment),
	}, nil
}

func (h *CommentHandler) UpdateComment(ctx context.Context, req *pb.UpdateCommentRequest) (*pb.CommentResponse, error) {
	// In a real app, you'd verify that the user owns this comment
	// For now, we allow anyone to update (or require auth)

	// Create command
	cmd := &command.UpdateCommentCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
		Content:     req.Content,
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update comment: %v", err)
	}

	// Query updated comment
	getQuery := &query.GetCommentByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		CommentID: req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, getQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve updated comment: %v", err)
	}

	comment := result.(*domain.Comment)

	return &pb.CommentResponse{
		Comment: domainCommentToProto(comment),
	}, nil
}

func (h *CommentHandler) DeleteComment(ctx context.Context, req *pb.DeleteCommentRequest) (*pb.DeleteResponse, error) {
	// Require authentication for deletion
	_, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	// Create command
	cmd := &command.DeleteCommentCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete comment: %v", err)
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Comment deleted successfully",
	}, nil
}

func (h *CommentHandler) ListComments(ctx context.Context, req *pb.ListCommentsRequest) (*pb.ListCommentsResponse, error) {
	// Convert request to filters
	filters := repository.CommentFilters{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
		PostID:   stringPtrFromOptional(req.PostId),
		Status:   commentStatusFromProto(req.Status),
	}

	// Default pagination
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 {
		filters.PageSize = 20
	}

	// Create query
	listQuery := &query.ListCommentsQuery{
		BaseQuery: cqrs.BaseQuery{},
		Filters:   filters,
	}

	// Dispatch query
	result, err := h.queryBus.Dispatch(ctx, listQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list comments: %v", err)
	}

	listResult := result.(*query.ListCommentsResult)

	// Convert to proto
	protoComments := make([]*pb.Comment, len(listResult.Comments))
	for i, comment := range listResult.Comments {
		protoComments[i] = domainCommentToProto(&comment)
	}

	return &pb.ListCommentsResponse{
		Comments: protoComments,
		Total:    int32(listResult.Total),
	}, nil
}

func (h *CommentHandler) ModerateComment(ctx context.Context, req *pb.ModerateCommentRequest) (*pb.CommentResponse, error) {
	// Require authentication for moderation (admin only)
	_, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	// Convert proto status to domain status
	domainStatus := protoCommentStatusToDomain(req.Status)

	// Create command
	cmd := &command.ModerateCommentCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
		Status:      domainStatus,
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to moderate comment: %v", err)
	}

	// Query moderated comment
	getQuery := &query.GetCommentByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		CommentID: req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, getQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve moderated comment: %v", err)
	}

	comment := result.(*domain.Comment)

	return &pb.CommentResponse{
		Comment: domainCommentToProto(comment),
	}, nil
}

// Helper functions for conversion

func domainCommentToProto(comment *domain.Comment) *pb.Comment {
	if comment == nil {
		return nil
	}

	protoComment := &pb.Comment{
		Id:          comment.ID,
		PostId:      comment.PostID,
		AuthorName:  comment.AuthorName,
		AuthorEmail: comment.AuthorEmail,
		Content:     comment.Content,
		Status:      domainCommentStatusToProto(comment.Status),
		CreatedAt:   timestamppb.New(comment.CreatedAt),
		UpdatedAt:   timestamppb.New(comment.UpdatedAt),
	}

	// Add parent ID if exists
	if comment.ParentID != nil {
		protoComment.ParentId = stringPtrToOptional(comment.ParentID)
	}

	return protoComment
}

func domainCommentStatusToProto(status domain.CommentStatus) pb.CommentStatus {
	switch status {
	case domain.CommentStatusPending:
		return pb.CommentStatus_COMMENT_STATUS_PENDING
	case domain.CommentStatusApproved:
		return pb.CommentStatus_COMMENT_STATUS_APPROVED
	case domain.CommentStatusSpam:
		return pb.CommentStatus_COMMENT_STATUS_SPAM
	case domain.CommentStatusTrash:
		return pb.CommentStatus_COMMENT_STATUS_TRASH
	default:
		return pb.CommentStatus_COMMENT_STATUS_UNSPECIFIED
	}
}

func protoCommentStatusToDomain(status pb.CommentStatus) domain.CommentStatus {
	switch status {
	case pb.CommentStatus_COMMENT_STATUS_PENDING:
		return domain.CommentStatusPending
	case pb.CommentStatus_COMMENT_STATUS_APPROVED:
		return domain.CommentStatusApproved
	case pb.CommentStatus_COMMENT_STATUS_SPAM:
		return domain.CommentStatusSpam
	case pb.CommentStatus_COMMENT_STATUS_TRASH:
		return domain.CommentStatusTrash
	default:
		return domain.CommentStatusPending
	}
}

func commentStatusFromProto(status *pb.CommentStatus) *domain.CommentStatus {
	if status == nil || *status == pb.CommentStatus_COMMENT_STATUS_UNSPECIFIED {
		return nil
	}
	domainStatus := protoCommentStatusToDomain(*status)
	return &domainStatus
}

// requireAuth checks authentication if enabled, returns user context or error
func (h *CommentHandler) requireAuth(ctx context.Context) (*auth.UserContext, error) {
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
