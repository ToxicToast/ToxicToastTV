package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "toxictoast/services/blog-service/api/proto"
	"github.com/toxictoast/toxictoastgo/shared/auth"

	"toxictoast/services/blog-service/internal/domain"
	"toxictoast/services/blog-service/internal/repository"
	"toxictoast/services/blog-service/internal/usecase"
)

type PostHandler struct {
	pb.UnimplementedBlogServiceServer
	postUseCase usecase.PostUseCase
	authEnabled bool
}

func NewPostHandler(postUseCase usecase.PostUseCase, authEnabled bool) *PostHandler {
	return &PostHandler{
		postUseCase: postUseCase,
		authEnabled: authEnabled,
	}
}

func (h *PostHandler) CreatePost(ctx context.Context, req *pb.CreatePostRequest) (*pb.PostResponse, error) {
	// Get user from context (authenticated by middleware)
	user, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	// Convert request to use case input
	input := usecase.CreatePostInput{
		Title:           req.Title,
		Content:         req.Content,
		Excerpt:         req.Excerpt,
		CategoryIDs:     req.CategoryIds,
		TagIDs:          req.TagIds,
		FeaturedImageID: stringPtrFromString(req.FeaturedImageId),
		Featured:        req.Featured,
		AuthorID:        user.UserID,
	}

	// Convert SEO metadata
	if req.Seo != nil {
		input.SEO = usecase.SEOInput{
			MetaTitle:       req.Seo.MetaTitle,
			MetaDescription: req.Seo.MetaDescription,
			MetaKeywords:    req.Seo.MetaKeywords,
			OGTitle:         req.Seo.OgTitle,
			OGDescription:   req.Seo.OgDescription,
			OGImage:         req.Seo.OgImage,
			CanonicalURL:    req.Seo.CanonicalUrl,
		}
	}

	// Create post
	post, err := h.postUseCase.CreatePost(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create post: %v", err)
	}

	return &pb.PostResponse{
		Post: domainPostToProto(post),
	}, nil
}

func (h *PostHandler) GetPost(ctx context.Context, req *pb.GetPostRequest) (*pb.PostResponse, error) {
	var post *domain.Post
	var err error

	switch identifier := req.Identifier.(type) {
	case *pb.GetPostRequest_Id:
		post, err = h.postUseCase.GetPost(ctx, identifier.Id)
	case *pb.GetPostRequest_Slug:
		post, err = h.postUseCase.GetPostBySlug(ctx, identifier.Slug)
		// Increment view count for public access via slug
		if err == nil && post != nil {
			_ = h.postUseCase.IncrementViewCount(ctx, post.ID)
		}
	default:
		return nil, status.Error(codes.InvalidArgument, "must provide either id or slug")
	}

	if err != nil {
		return nil, status.Errorf(codes.NotFound, "post not found: %v", err)
	}

	return &pb.PostResponse{
		Post: domainPostToProto(post),
	}, nil
}

func (h *PostHandler) UpdatePost(ctx context.Context, req *pb.UpdatePostRequest) (*pb.PostResponse, error) {
	// Get user from context (authenticated by middleware)
	_, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	// Convert request to use case input
	input := usecase.UpdatePostInput{
		Title:           stringPtrFromOptional(req.Title),
		Content:         stringPtrFromOptional(req.Content),
		Excerpt:         stringPtrFromOptional(req.Excerpt),
		CategoryIDs:     req.CategoryIds,
		TagIDs:          req.TagIds,
		FeaturedImageID: stringPtrFromOptional(req.FeaturedImageId),
		Featured:        boolPtrFromOptional(req.Featured),
	}

	// Convert SEO metadata if provided
	if req.Seo != nil {
		input.SEO = &usecase.SEOInput{
			MetaTitle:       req.Seo.MetaTitle,
			MetaDescription: req.Seo.MetaDescription,
			MetaKeywords:    req.Seo.MetaKeywords,
			OGTitle:         req.Seo.OgTitle,
			OGDescription:   req.Seo.OgDescription,
			OGImage:         req.Seo.OgImage,
			CanonicalURL:    req.Seo.CanonicalUrl,
		}
	}

	// Update post
	post, err := h.postUseCase.UpdatePost(ctx, req.Id, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update post: %v", err)
	}

	return &pb.PostResponse{
		Post: domainPostToProto(post),
	}, nil
}

func (h *PostHandler) DeletePost(ctx context.Context, req *pb.DeletePostRequest) (*pb.DeleteResponse, error) {
	// Get user from context (authenticated by middleware)
	_, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	// Delete post
	if err := h.postUseCase.DeletePost(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete post: %v", err)
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Post deleted successfully",
	}, nil
}

func (h *PostHandler) ListPosts(ctx context.Context, req *pb.ListPostsRequest) (*pb.ListPostsResponse, error) {
	// Convert request to filters
	filters := repository.PostFilters{
		Page:       int(req.Page),
		PageSize:   int(req.PageSize),
		CategoryID: stringPtrFromOptional(req.CategoryId),
		TagID:      stringPtrFromOptional(req.TagId),
		AuthorID:   stringPtrFromOptional(req.AuthorId),
		Status:     postStatusFromProto(req.Status),
		Featured:   boolPtrFromOptional(req.Featured),
		Search:     stringPtrFromOptional(req.Search),
		SortBy:     req.SortBy,
		SortOrder:  req.SortOrder,
	}

	// Default pagination
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 {
		filters.PageSize = 10
	}

	// List posts
	posts, total, err := h.postUseCase.ListPosts(ctx, filters)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list posts: %v", err)
	}

	// Convert to proto
	protoPosts := make([]*pb.Post, len(posts))
	for i, post := range posts {
		protoPosts[i] = domainPostToProto(&post)
	}

	// Calculate total pages
	totalPages := int32(total) / req.PageSize
	if int32(total)%req.PageSize > 0 {
		totalPages++
	}

	return &pb.ListPostsResponse{
		Posts:      protoPosts,
		Total:      int32(total),
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (h *PostHandler) PublishPost(ctx context.Context, req *pb.PublishPostRequest) (*pb.PostResponse, error) {
	// Get user from context (authenticated by middleware)
	_, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	// Publish post
	post, err := h.postUseCase.PublishPost(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish post: %v", err)
	}

	return &pb.PostResponse{
		Post: domainPostToProto(post),
	}, nil
}

// Helper functions for conversion

func domainPostToProto(post *domain.Post) *pb.Post {
	if post == nil {
		return nil
	}

	protoPost := &pb.Post{
		Id:             post.ID,
		Title:          post.Title,
		Slug:           post.Slug,
		Content:        post.Content,
		Excerpt:        post.Excerpt,
		Markdown:       post.Markdown,
		Html:           post.HTML,
		Status:         domainPostStatusToProto(post.Status),
		Featured:       post.Featured,
		AuthorId:       post.AuthorID,
		ReadingTime:    int32(post.ReadingTime),
		ViewCount:      int32(post.ViewCount),
		CreatedAt:      timestamppb.New(post.CreatedAt),
		UpdatedAt:      timestamppb.New(post.UpdatedAt),
	}

	// Add published date if available
	if post.PublishedAt != nil {
		protoPost.PublishedAt = timestamppb.New(*post.PublishedAt)
	}

	// Add featured image if available
	if post.FeaturedImageID != nil {
		protoPost.FeaturedImageId = *post.FeaturedImageID
	}

	// Add categories
	categoryIDs := make([]string, len(post.Categories))
	for i, cat := range post.Categories {
		categoryIDs[i] = cat.ID
	}
	protoPost.CategoryIds = categoryIDs

	// Add tags
	tagIDs := make([]string, len(post.Tags))
	for i, tag := range post.Tags {
		tagIDs[i] = tag.ID
	}
	protoPost.TagIds = tagIDs

	// Add SEO metadata
	protoPost.Seo = &pb.SEOMetadata{
		MetaTitle:       post.MetaTitle,
		MetaDescription: post.MetaDescription,
		MetaKeywords:    splitString(post.MetaKeywords, ","),
		OgTitle:         post.OGTitle,
		OgDescription:   post.OGDescription,
		OgImage:         post.OGImage,
		CanonicalUrl:    post.CanonicalURL,
	}

	return protoPost
}

func domainPostStatusToProto(status domain.PostStatus) pb.PostStatus {
	switch status {
	case domain.PostStatusDraft:
		return pb.PostStatus_POST_STATUS_DRAFT
	case domain.PostStatusPublished:
		return pb.PostStatus_POST_STATUS_PUBLISHED
	default:
		return pb.PostStatus_POST_STATUS_UNSPECIFIED
	}
}

func postStatusFromProto(status *pb.PostStatus) *domain.PostStatus {
	if status == nil {
		return nil
	}

	var domainStatus domain.PostStatus
	switch *status {
	case pb.PostStatus_POST_STATUS_DRAFT:
		domainStatus = domain.PostStatusDraft
	case pb.PostStatus_POST_STATUS_PUBLISHED:
		domainStatus = domain.PostStatusPublished
	default:
		return nil
	}

	return &domainStatus
}

func stringPtrFromString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func stringPtrFromOptional(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	return s
}

func boolPtrFromOptional(b *bool) *bool {
	return b
}

func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	result := []string{}
	current := ""
	for i := 0; i < len(s); i++ {
		if string(s[i]) == sep {
			if current != "" {
				result = append(result, current)
			}
			current = ""
		} else {
			current += string(s[i])
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

// requireAuth checks authentication if enabled, returns user context or error
func (h *PostHandler) requireAuth(ctx context.Context) (*auth.UserContext, error) {
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
