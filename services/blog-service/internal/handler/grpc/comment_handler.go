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

type CommentHandler struct {
	pb.UnimplementedBlogServiceServer
	commentUseCase usecase.CommentUseCase
	authEnabled    bool
}

func NewCommentHandler(commentUseCase usecase.CommentUseCase, authEnabled bool) *CommentHandler {
	return &CommentHandler{
		commentUseCase: commentUseCase,
		authEnabled:    authEnabled,
	}
}

func (h *CommentHandler) CreateComment(ctx context.Context, req *pb.CreateCommentRequest) (*pb.CommentResponse, error) {
	// Comments can be created by anyone (public endpoint)
	// Convert request to use case input
	input := usecase.CreateCommentInput{
		PostID:      req.PostId,
		ParentID:    stringPtrFromOptional(req.ParentId),
		AuthorName:  req.AuthorName,
		AuthorEmail: req.AuthorEmail,
		Content:     req.Content,
	}

	// Create comment
	comment, err := h.commentUseCase.CreateComment(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create comment: %v", err)
	}

	return &pb.CommentResponse{
		Comment: domainCommentToProto(comment),
	}, nil
}

func (h *CommentHandler) GetComment(ctx context.Context, req *pb.GetCommentRequest) (*pb.CommentResponse, error) {
	comment, err := h.commentUseCase.GetComment(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "comment not found: %v", err)
	}

	return &pb.CommentResponse{
		Comment: domainCommentToProto(comment),
	}, nil
}

func (h *CommentHandler) UpdateComment(ctx context.Context, req *pb.UpdateCommentRequest) (*pb.CommentResponse, error) {
	// In a real app, you'd verify that the user owns this comment
	// For now, we allow anyone to update (or require auth)

	// Convert request to use case input
	input := usecase.UpdateCommentInput{
		Content: req.Content,
	}

	// Update comment
	comment, err := h.commentUseCase.UpdateComment(ctx, req.Id, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update comment: %v", err)
	}

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

	// Delete comment
	if err := h.commentUseCase.DeleteComment(ctx, req.Id); err != nil {
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

	// List comments
	comments, total, err := h.commentUseCase.ListComments(ctx, filters)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list comments: %v", err)
	}

	// Convert to proto
	protoComments := make([]*pb.Comment, len(comments))
	for i, comment := range comments {
		protoComments[i] = domainCommentToProto(&comment)
	}

	return &pb.ListCommentsResponse{
		Comments: protoComments,
		Total:    int32(total),
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

	// Moderate comment
	comment, err := h.commentUseCase.ModerateComment(ctx, req.Id, domainStatus)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to moderate comment: %v", err)
	}

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
