package grpc

import (
	"context"

	pb "toxictoast/services/blog-service/api/proto"
)

// BlogHandler composes all individual handlers into a single gRPC service implementation
type BlogHandler struct {
	pb.UnimplementedBlogServiceServer
	postHandler     *PostHandler
	categoryHandler *CategoryHandler
	tagHandler      *TagHandler
	commentHandler  *CommentHandler
	mediaHandler    *MediaHandler
}

// NewBlogHandler creates a new composed blog handler
func NewBlogHandler(
	postHandler *PostHandler,
	categoryHandler *CategoryHandler,
	tagHandler *TagHandler,
	commentHandler *CommentHandler,
	mediaHandler *MediaHandler,
) *BlogHandler {
	return &BlogHandler{
		postHandler:     postHandler,
		categoryHandler: categoryHandler,
		tagHandler:      tagHandler,
		commentHandler:  commentHandler,
		mediaHandler:    mediaHandler,
	}
}

// Post operations - delegate to PostHandler

func (h *BlogHandler) CreatePost(ctx context.Context, req *pb.CreatePostRequest) (*pb.PostResponse, error) {
	return h.postHandler.CreatePost(ctx, req)
}

func (h *BlogHandler) GetPost(ctx context.Context, req *pb.GetPostRequest) (*pb.PostResponse, error) {
	return h.postHandler.GetPost(ctx, req)
}

func (h *BlogHandler) UpdatePost(ctx context.Context, req *pb.UpdatePostRequest) (*pb.PostResponse, error) {
	return h.postHandler.UpdatePost(ctx, req)
}

func (h *BlogHandler) DeletePost(ctx context.Context, req *pb.DeletePostRequest) (*pb.DeleteResponse, error) {
	return h.postHandler.DeletePost(ctx, req)
}

func (h *BlogHandler) ListPosts(ctx context.Context, req *pb.ListPostsRequest) (*pb.ListPostsResponse, error) {
	return h.postHandler.ListPosts(ctx, req)
}

func (h *BlogHandler) PublishPost(ctx context.Context, req *pb.PublishPostRequest) (*pb.PostResponse, error) {
	return h.postHandler.PublishPost(ctx, req)
}

// Category operations - delegate to CategoryHandler

func (h *BlogHandler) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.CategoryResponse, error) {
	return h.categoryHandler.CreateCategory(ctx, req)
}

func (h *BlogHandler) GetCategory(ctx context.Context, req *pb.GetCategoryRequest) (*pb.CategoryResponse, error) {
	return h.categoryHandler.GetCategory(ctx, req)
}

func (h *BlogHandler) UpdateCategory(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.CategoryResponse, error) {
	return h.categoryHandler.UpdateCategory(ctx, req)
}

func (h *BlogHandler) DeleteCategory(ctx context.Context, req *pb.DeleteCategoryRequest) (*pb.DeleteResponse, error) {
	return h.categoryHandler.DeleteCategory(ctx, req)
}

func (h *BlogHandler) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	return h.categoryHandler.ListCategories(ctx, req)
}

// Tag operations - delegate to TagHandler

func (h *BlogHandler) CreateTag(ctx context.Context, req *pb.CreateTagRequest) (*pb.TagResponse, error) {
	return h.tagHandler.CreateTag(ctx, req)
}

func (h *BlogHandler) GetTag(ctx context.Context, req *pb.GetTagRequest) (*pb.TagResponse, error) {
	return h.tagHandler.GetTag(ctx, req)
}

func (h *BlogHandler) UpdateTag(ctx context.Context, req *pb.UpdateTagRequest) (*pb.TagResponse, error) {
	return h.tagHandler.UpdateTag(ctx, req)
}

func (h *BlogHandler) DeleteTag(ctx context.Context, req *pb.DeleteTagRequest) (*pb.DeleteResponse, error) {
	return h.tagHandler.DeleteTag(ctx, req)
}

func (h *BlogHandler) ListTags(ctx context.Context, req *pb.ListTagsRequest) (*pb.ListTagsResponse, error) {
	return h.tagHandler.ListTags(ctx, req)
}

// Comment operations - delegate to CommentHandler

func (h *BlogHandler) CreateComment(ctx context.Context, req *pb.CreateCommentRequest) (*pb.CommentResponse, error) {
	return h.commentHandler.CreateComment(ctx, req)
}

func (h *BlogHandler) GetComment(ctx context.Context, req *pb.GetCommentRequest) (*pb.CommentResponse, error) {
	return h.commentHandler.GetComment(ctx, req)
}

func (h *BlogHandler) UpdateComment(ctx context.Context, req *pb.UpdateCommentRequest) (*pb.CommentResponse, error) {
	return h.commentHandler.UpdateComment(ctx, req)
}

func (h *BlogHandler) DeleteComment(ctx context.Context, req *pb.DeleteCommentRequest) (*pb.DeleteResponse, error) {
	return h.commentHandler.DeleteComment(ctx, req)
}

func (h *BlogHandler) ListComments(ctx context.Context, req *pb.ListCommentsRequest) (*pb.ListCommentsResponse, error) {
	return h.commentHandler.ListComments(ctx, req)
}

func (h *BlogHandler) ModerateComment(ctx context.Context, req *pb.ModerateCommentRequest) (*pb.CommentResponse, error) {
	return h.commentHandler.ModerateComment(ctx, req)
}

// Media operations - delegate to MediaHandler

func (h *BlogHandler) UploadMedia(stream pb.BlogService_UploadMediaServer) error {
	return h.mediaHandler.UploadMedia(stream)
}

func (h *BlogHandler) GetMedia(ctx context.Context, req *pb.GetMediaRequest) (*pb.MediaResponse, error) {
	return h.mediaHandler.GetMedia(ctx, req)
}

func (h *BlogHandler) DeleteMedia(ctx context.Context, req *pb.DeleteMediaRequest) (*pb.DeleteResponse, error) {
	return h.mediaHandler.DeleteMedia(ctx, req)
}

func (h *BlogHandler) ListMedia(ctx context.Context, req *pb.ListMediaRequest) (*pb.ListMediaResponse, error) {
	return h.mediaHandler.ListMedia(ctx, req)
}
