package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	pb "toxictoast/services/blog-service/api/proto"
	"google.golang.org/grpc"
)

// BlogHandler handles HTTP-to-gRPC translation for blog service
type BlogHandler struct {
	client pb.BlogServiceClient
}

// NewBlogHandler creates a new blog handler
func NewBlogHandler(conn *grpc.ClientConn) *BlogHandler {
	return &BlogHandler{
		client: pb.NewBlogServiceClient(conn),
	}
}

// RegisterRoutes registers all blog routes
func (h *BlogHandler) RegisterRoutes(router *mux.Router) {
	// Post routes
	router.HandleFunc("/posts", h.ListPosts).Methods("GET")
	router.HandleFunc("/posts", h.CreatePost).Methods("POST")
	router.HandleFunc("/posts/{id}", h.GetPost).Methods("GET")
	router.HandleFunc("/posts/{id}", h.UpdatePost).Methods("PUT")
	router.HandleFunc("/posts/{id}", h.DeletePost).Methods("DELETE")
	router.HandleFunc("/posts/{id}/publish", h.PublishPost).Methods("POST")

	// Category routes
	router.HandleFunc("/categories", h.ListCategories).Methods("GET")
	router.HandleFunc("/categories", h.CreateCategory).Methods("POST")
	router.HandleFunc("/categories/{id}", h.GetCategory).Methods("GET")
	router.HandleFunc("/categories/{id}", h.UpdateCategory).Methods("PUT")
	router.HandleFunc("/categories/{id}", h.DeleteCategory).Methods("DELETE")

	// Tag routes
	router.HandleFunc("/tags", h.ListTags).Methods("GET")
	router.HandleFunc("/tags", h.CreateTag).Methods("POST")
	router.HandleFunc("/tags/{id}", h.GetTag).Methods("GET")
	router.HandleFunc("/tags/{id}", h.UpdateTag).Methods("PUT")
	router.HandleFunc("/tags/{id}", h.DeleteTag).Methods("DELETE")

	// Media routes
	router.HandleFunc("/media", h.ListMedia).Methods("GET")
	router.HandleFunc("/media/{id}", h.GetMedia).Methods("GET")
	router.HandleFunc("/media/{id}", h.DeleteMedia).Methods("DELETE")
	// Upload handled separately due to streaming

	// Comment routes
	router.HandleFunc("/comments", h.ListComments).Methods("GET")
	router.HandleFunc("/comments", h.CreateComment).Methods("POST")
	router.HandleFunc("/comments/{id}", h.GetComment).Methods("GET")
	router.HandleFunc("/comments/{id}", h.UpdateComment).Methods("PUT")
	router.HandleFunc("/comments/{id}", h.DeleteComment).Methods("DELETE")
	router.HandleFunc("/comments/{id}/moderate", h.ModerateComment).Methods("POST")
}

// ListPosts handles GET /posts
func (h *BlogHandler) ListPosts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)

	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}

	req := &pb.ListPostsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		SortBy:   r.URL.Query().Get("sort_by"),
		SortOrder: r.URL.Query().Get("sort_order"),
	}

	// Optional filters
	if categoryID := r.URL.Query().Get("category_id"); categoryID != "" {
		req.CategoryId = &categoryID
	}
	if tagID := r.URL.Query().Get("tag_id"); tagID != "" {
		req.TagId = &tagID
	}
	if authorID := r.URL.Query().Get("author_id"); authorID != "" {
		req.AuthorId = &authorID
	}
	if search := r.URL.Query().Get("search"); search != "" {
		req.Search = &search
	}
	if featured := r.URL.Query().Get("featured"); featured != "" {
		featuredBool := featured == "true"
		req.Featured = &featuredBool
	}
	if status := r.URL.Query().Get("status"); status != "" {
		statusValue := parsePostStatus(status)
		req.Status = &statusValue
	}

	// Call gRPC service
	resp, err := h.client.ListPosts(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to list posts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CreatePost handles POST /posts
func (h *BlogHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var req pb.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.client.CreatePost(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to create post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// GetPost handles GET /posts/{id}
func (h *BlogHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Check if it's a slug or ID
	req := &pb.GetPostRequest{}
	if strings.Contains(id, "-") {
		// Likely a slug
		req.Identifier = &pb.GetPostRequest_Slug{Slug: id}
	} else {
		req.Identifier = &pb.GetPostRequest_Id{Id: id}
	}

	resp, err := h.client.GetPost(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get post: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UpdatePost handles PUT /posts/{id}
func (h *BlogHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req pb.UpdatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.Id = id

	resp, err := h.client.UpdatePost(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to update post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DeletePost handles DELETE /posts/{id}
func (h *BlogHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := &pb.DeletePostRequest{Id: id}
	resp, err := h.client.DeletePost(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to delete post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// PublishPost handles POST /posts/{id}/publish
func (h *BlogHandler) PublishPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := &pb.PublishPostRequest{Id: id}
	resp, err := h.client.PublishPost(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to publish post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ListCategories handles GET /categories
func (h *BlogHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)

	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 50
	}

	req := &pb.ListCategoriesRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	if parentID := r.URL.Query().Get("parent_id"); parentID != "" {
		req.ParentId = &parentID
	}

	resp, err := h.client.ListCategories(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to list categories: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CreateCategory handles POST /categories
func (h *BlogHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.client.CreateCategory(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to create category: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// GetCategory handles GET /categories/{id}
func (h *BlogHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := &pb.GetCategoryRequest{}
	if strings.Contains(id, "-") {
		req.Identifier = &pb.GetCategoryRequest_Slug{Slug: id}
	} else {
		req.Identifier = &pb.GetCategoryRequest_Id{Id: id}
	}

	resp, err := h.client.GetCategory(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get category: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UpdateCategory handles PUT /categories/{id}
func (h *BlogHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req pb.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.Id = id

	resp, err := h.client.UpdateCategory(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to update category: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DeleteCategory handles DELETE /categories/{id}
func (h *BlogHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := &pb.DeleteCategoryRequest{Id: id}
	resp, err := h.client.DeleteCategory(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to delete category: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ListTags handles GET /tags
func (h *BlogHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)

	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 50
	}

	req := &pb.ListTagsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	if search := r.URL.Query().Get("search"); search != "" {
		req.Search = &search
	}

	resp, err := h.client.ListTags(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to list tags: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CreateTag handles POST /tags
func (h *BlogHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.client.CreateTag(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to create tag: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// GetTag handles GET /tags/{id}
func (h *BlogHandler) GetTag(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := &pb.GetTagRequest{}
	if strings.Contains(id, "-") {
		req.Identifier = &pb.GetTagRequest_Slug{Slug: id}
	} else {
		req.Identifier = &pb.GetTagRequest_Id{Id: id}
	}

	resp, err := h.client.GetTag(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get tag: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UpdateTag handles PUT /tags/{id}
func (h *BlogHandler) UpdateTag(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req pb.UpdateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.Id = id

	resp, err := h.client.UpdateTag(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to update tag: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DeleteTag handles DELETE /tags/{id}
func (h *BlogHandler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := &pb.DeleteTagRequest{Id: id}
	resp, err := h.client.DeleteTag(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to delete tag: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ListMedia handles GET /media
func (h *BlogHandler) ListMedia(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)

	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}

	req := &pb.ListMediaRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	if mimeType := r.URL.Query().Get("mime_type"); mimeType != "" {
		req.MimeType = &mimeType
	}

	resp, err := h.client.ListMedia(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to list media: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetMedia handles GET /media/{id}
func (h *BlogHandler) GetMedia(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := &pb.GetMediaRequest{Id: id}
	resp, err := h.client.GetMedia(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get media: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DeleteMedia handles DELETE /media/{id}
func (h *BlogHandler) DeleteMedia(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := &pb.DeleteMediaRequest{Id: id}
	resp, err := h.client.DeleteMedia(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to delete media: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ListComments handles GET /comments
func (h *BlogHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)

	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}

	req := &pb.ListCommentsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	if postID := r.URL.Query().Get("post_id"); postID != "" {
		req.PostId = &postID
	}
	if status := r.URL.Query().Get("status"); status != "" {
		statusValue := parseCommentStatus(status)
		req.Status = &statusValue
	}

	resp, err := h.client.ListComments(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to list comments: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CreateComment handles POST /comments
func (h *BlogHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.client.CreateComment(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to create comment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// GetComment handles GET /comments/{id}
func (h *BlogHandler) GetComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := &pb.GetCommentRequest{Id: id}
	resp, err := h.client.GetComment(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get comment: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UpdateComment handles PUT /comments/{id}
func (h *BlogHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req pb.UpdateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.Id = id

	resp, err := h.client.UpdateComment(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to update comment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DeleteComment handles DELETE /comments/{id}
func (h *BlogHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := &pb.DeleteCommentRequest{Id: id}
	resp, err := h.client.DeleteComment(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to delete comment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ModerateComment handles POST /comments/{id}/moderate
func (h *BlogHandler) ModerateComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var reqBody struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	req := &pb.ModerateCommentRequest{
		Id:     id,
		Status: parseCommentStatus(reqBody.Status),
	}

	resp, err := h.client.ModerateComment(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to moderate comment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Helper functions

func parsePostStatus(status string) pb.PostStatus {
	switch strings.ToLower(status) {
	case "draft":
		return pb.PostStatus_POST_STATUS_DRAFT
	case "published":
		return pb.PostStatus_POST_STATUS_PUBLISHED
	default:
		return pb.PostStatus_POST_STATUS_UNSPECIFIED
	}
}

func parseCommentStatus(status string) pb.CommentStatus {
	switch strings.ToLower(status) {
	case "pending":
		return pb.CommentStatus_COMMENT_STATUS_PENDING
	case "approved":
		return pb.CommentStatus_COMMENT_STATUS_APPROVED
	case "spam":
		return pb.CommentStatus_COMMENT_STATUS_SPAM
	case "trash":
		return pb.CommentStatus_COMMENT_STATUS_TRASH
	default:
		return pb.CommentStatus_COMMENT_STATUS_UNSPECIFIED
	}
}
