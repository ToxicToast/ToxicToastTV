package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	sharedmiddleware "github.com/toxictoast/toxictoastgo/shared/middleware"
	authpb "toxictoast/services/auth-service/api/proto"
	userpb "toxictoast/services/user-service/api/proto"
	"google.golang.org/grpc"
)

// AuthHandler handles HTTP-to-gRPC translation for auth service
type AuthHandler struct {
	authClient     authpb.AuthServiceClient
	userClient     userpb.UserServiceClient
	authMiddleware *sharedmiddleware.AuthMiddleware
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authConn, userConn *grpc.ClientConn, authMiddleware *sharedmiddleware.AuthMiddleware) *AuthHandler {
	return &AuthHandler{
		authClient:     authpb.NewAuthServiceClient(authConn),
		userClient:     userpb.NewUserServiceClient(userConn),
		authMiddleware: authMiddleware,
	}
}

// RegisterRoutes registers all auth routes
func (h *AuthHandler) RegisterRoutes(router *mux.Router, rateLimiter *sharedmiddleware.RateLimiter) {
	// Authentication routes with rate limiting
	if rateLimiter != nil {
		router.Handle("/register", rateLimiter.Limit(http.HandlerFunc(h.Register))).Methods("POST")
		router.Handle("/login", rateLimiter.Limit(http.HandlerFunc(h.Login))).Methods("POST")
		router.Handle("/refresh", rateLimiter.Limit(http.HandlerFunc(h.RefreshToken))).Methods("POST")
	} else {
		router.HandleFunc("/register", h.Register).Methods("POST")
		router.HandleFunc("/login", h.Login).Methods("POST")
		router.HandleFunc("/refresh", h.RefreshToken).Methods("POST")
	}

	// Other auth routes (no rate limiting)
	router.HandleFunc("/logout", h.Logout).Methods("POST")
	router.HandleFunc("/validate", h.ValidateToken).Methods("POST")

	// Role management routes
	router.HandleFunc("/roles", h.ListRoles).Methods("GET")
	router.HandleFunc("/roles", h.CreateRole).Methods("POST")
	router.HandleFunc("/roles/{id}", h.GetRole).Methods("GET")
	router.HandleFunc("/roles/{id}", h.UpdateRole).Methods("PUT")
	router.HandleFunc("/roles/{id}", h.DeleteRole).Methods("DELETE")

	// Permission management routes
	router.HandleFunc("/permissions", h.ListPermissions).Methods("GET")
	router.HandleFunc("/permissions", h.CreatePermission).Methods("POST")
	router.HandleFunc("/permissions/{id}", h.GetPermission).Methods("GET")
	router.HandleFunc("/permissions/{id}", h.UpdatePermission).Methods("PUT")
	router.HandleFunc("/permissions/{id}", h.DeletePermission).Methods("DELETE")

	// RBAC routes
	router.HandleFunc("/users/{user_id}/roles", h.AssignRole).Methods("POST")
	router.HandleFunc("/users/{user_id}/roles/{role_id}", h.RevokeRole).Methods("DELETE")
	router.HandleFunc("/users/{user_id}/roles", h.ListUserRoles).Methods("GET")
	router.HandleFunc("/roles/{role_id}/permissions", h.AssignPermission).Methods("POST")
	router.HandleFunc("/roles/{role_id}/permissions/{permission_id}", h.RevokePermission).Methods("DELETE")
	router.HandleFunc("/roles/{role_id}/permissions", h.ListRolePermissions).Methods("GET")
	router.HandleFunc("/users/{user_id}/permissions", h.ListUserPermissions).Methods("GET")
	router.HandleFunc("/users/{user_id}/check-permission", h.CheckPermission).Methods("POST")

	// User management routes
	router.HandleFunc("/users", h.ListUsers).Methods("GET")
	router.HandleFunc("/users/{id}", h.GetUser).Methods("GET")
	router.HandleFunc("/users/{id}", h.UpdateUser).Methods("PUT")
	router.HandleFunc("/users/{id}", h.DeleteUser).Methods("DELETE")
	router.HandleFunc("/users/{id}/password", h.UpdatePassword).Methods("PUT")
	router.HandleFunc("/users/{id}/activate", h.ActivateUser).Methods("POST")
	router.HandleFunc("/users/{id}/deactivate", h.DeactivateUser).Methods("POST")
}

// Register handles POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email     string `json:"email"`
		Username  string `json:"username"`
		Password  string `json:"password"`
		FirstName string `json:"first_name,omitempty"`
		LastName  string `json:"last_name,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	pbReq := &authpb.RegisterRequest{
		Email:     req.Email,
		Username:  req.Username,
		Password:  req.Password,
		FirstName: &req.FirstName,
		LastName:  &req.LastName,
	}

	resp, err := h.authClient.Register(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Registration failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	pbReq := &authpb.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	resp, err := h.authClient.Login(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Login failed: "+err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Logout handles POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Check for Bearer token format
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid authorization header format. Expected: Bearer <token>", http.StatusUnauthorized)
		return
	}

	tokenString := parts[1]

	// Validate token to ensure it's valid before blacklisting
	_, err := h.authClient.ValidateToken(context.Background(), &authpb.ValidateTokenRequest{
		Token: tokenString,
	})
	if err != nil {
		// Token is already invalid, no need to blacklist
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	// Add token to blacklist with standard access token duration (15 minutes from now)
	// The token will remain blacklisted until it would have expired naturally
	expiresAt := time.Now().Add(15 * time.Minute)
	h.authMiddleware.GetTokenBlacklist().Revoke(tokenString, expiresAt)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Logged out successfully",
	})
}

// ValidateToken handles POST /auth/validate
func (h *AuthHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	pbReq := &authpb.ValidateTokenRequest{
		Token: req.Token,
	}

	resp, err := h.authClient.ValidateToken(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Token validation failed: "+err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// RefreshToken handles POST /auth/refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	pbReq := &authpb.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	}

	resp, err := h.authClient.RefreshToken(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Token refresh failed: "+err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CreateRole handles POST /auth/roles
func (h *AuthHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	pbReq := &authpb.CreateRoleRequest{
		Name:        req.Name,
		Description: req.Description,
	}

	resp, err := h.authClient.CreateRole(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Failed to create role: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// ListRoles handles GET /auth/roles
func (h *AuthHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	pbReq := &authpb.ListRolesRequest{}

	resp, err := h.authClient.ListRoles(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Failed to list roles: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetRole handles GET /auth/roles/{id}
func (h *AuthHandler) GetRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleID := vars["id"]

	pbReq := &authpb.GetRoleRequest{
		Id: roleID,
	}

	resp, err := h.authClient.GetRole(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Failed to get role: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UpdateRole handles PUT /auth/roles/{id}
func (h *AuthHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleID := vars["id"]

	var req struct {
		Name        string `json:"name,omitempty"`
		Description string `json:"description,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	pbReq := &authpb.UpdateRoleRequest{
		Id: roleID,
	}

	if req.Name != "" {
		pbReq.Name = &req.Name
	}
	if req.Description != "" {
		pbReq.Description = &req.Description
	}

	resp, err := h.authClient.UpdateRole(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Failed to update role: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DeleteRole handles DELETE /auth/roles/{id}
func (h *AuthHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleID := vars["id"]

	pbReq := &authpb.DeleteRoleRequest{
		Id: roleID,
	}

	resp, err := h.authClient.DeleteRole(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Failed to delete role: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CreatePermission handles POST /auth/permissions
func (h *AuthHandler) CreatePermission(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Resource    string `json:"resource"`
		Action      string `json:"action"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	pbReq := &authpb.CreatePermissionRequest{
		Resource:    req.Resource,
		Action:      req.Action,
		Description: req.Description,
	}

	resp, err := h.authClient.CreatePermission(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Failed to create permission: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// ListPermissions handles GET /auth/permissions
func (h *AuthHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	pbReq := &authpb.ListPermissionsRequest{}

	resp, err := h.authClient.ListPermissions(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Failed to list permissions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetPermission handles GET /auth/permissions/{id}
func (h *AuthHandler) GetPermission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	permissionID := vars["id"]

	pbReq := &authpb.GetPermissionRequest{
		Id: permissionID,
	}

	resp, err := h.authClient.GetPermission(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Failed to get permission: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UpdatePermission handles PUT /auth/permissions/{id}
func (h *AuthHandler) UpdatePermission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	permissionID := vars["id"]

	var req struct {
		Resource    string `json:"resource,omitempty"`
		Action      string `json:"action,omitempty"`
		Description string `json:"description,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	pbReq := &authpb.UpdatePermissionRequest{
		Id: permissionID,
	}

	if req.Resource != "" {
		pbReq.Resource = &req.Resource
	}
	if req.Action != "" {
		pbReq.Action = &req.Action
	}
	if req.Description != "" {
		pbReq.Description = &req.Description
	}

	resp, err := h.authClient.UpdatePermission(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Failed to update permission: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DeletePermission handles DELETE /auth/permissions/{id}
func (h *AuthHandler) DeletePermission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	permissionID := vars["id"]

	pbReq := &authpb.DeletePermissionRequest{
		Id: permissionID,
	}

	resp, err := h.authClient.DeletePermission(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Failed to delete permission: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// AssignRole handles POST /auth/users/{user_id}/roles
func (h *AuthHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	var req struct {
		RoleID string `json:"role_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	pbReq := &authpb.AssignRoleRequest{
		UserId: userID,
		RoleId: req.RoleID,
	}

	resp, err := h.authClient.AssignRole(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Failed to assign role: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// RevokeRole handles DELETE /auth/users/{user_id}/roles/{role_id}
func (h *AuthHandler) RevokeRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]
	roleID := vars["role_id"]

	pbReq := &authpb.RevokeRoleRequest{
		UserId: userID,
		RoleId: roleID,
	}

	resp, err := h.authClient.RevokeRole(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Failed to revoke role: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ListUserRoles handles GET /auth/users/{user_id}/roles
func (h *AuthHandler) ListUserRoles(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	pbReq := &authpb.ListUserRolesRequest{
		UserId: userID,
	}

	resp, err := h.authClient.ListUserRoles(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Failed to get user roles: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// AssignPermission handles POST /auth/roles/{role_id}/permissions
func (h *AuthHandler) AssignPermission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleID := vars["role_id"]

	var req struct {
		PermissionID string `json:"permission_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	pbReq := &authpb.AssignPermissionRequest{
		RoleId:       roleID,
		PermissionId: req.PermissionID,
	}

	resp, err := h.authClient.AssignPermission(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Failed to assign permission: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// RevokePermission handles DELETE /auth/roles/{role_id}/permissions/{permission_id}
func (h *AuthHandler) RevokePermission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleID := vars["role_id"]
	permissionID := vars["permission_id"]

	pbReq := &authpb.RevokePermissionRequest{
		RoleId:       roleID,
		PermissionId: permissionID,
	}

	resp, err := h.authClient.RevokePermission(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Failed to revoke permission: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ListRolePermissions handles GET /auth/roles/{role_id}/permissions
func (h *AuthHandler) ListRolePermissions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleID := vars["role_id"]

	pbReq := &authpb.ListRolePermissionsRequest{
		RoleId: roleID,
	}

	resp, err := h.authClient.ListRolePermissions(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Failed to get role permissions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ListUserPermissions handles GET /auth/users/{user_id}/permissions
func (h *AuthHandler) ListUserPermissions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	pbReq := &authpb.ListUserPermissionsRequest{
		UserId: userID,
	}

	resp, err := h.authClient.ListUserPermissions(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Failed to get user permissions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CheckPermission handles POST /auth/users/{user_id}/check-permission
func (h *AuthHandler) CheckPermission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	var req struct {
		Resource string `json:"resource"`
		Action   string `json:"action"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	pbReq := &authpb.CheckPermissionRequest{
		UserId:   userID,
		Resource: req.Resource,
		Action:   req.Action,
	}

	resp, err := h.authClient.CheckPermission(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Failed to check permission: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ListUsers handles GET /auth/users
func (h *AuthHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	pbReq := &userpb.ListUsersRequest{}

	resp, err := h.userClient.ListUsers(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Failed to list users: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetUser handles GET /auth/users/{id}
func (h *AuthHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	pbReq := &userpb.GetUserRequest{
		Id: userID,
	}

	resp, err := h.userClient.GetUser(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Failed to get user: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UpdateUser handles PUT /auth/users/{id}
func (h *AuthHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	var req struct {
		Email     string `json:"email,omitempty"`
		Username  string `json:"username,omitempty"`
		FirstName string `json:"first_name,omitempty"`
		LastName  string `json:"last_name,omitempty"`
		AvatarURL string `json:"avatar_url,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	pbReq := &userpb.UpdateUserRequest{
		Id: userID,
	}

	if req.Email != "" {
		pbReq.Email = &req.Email
	}
	if req.Username != "" {
		pbReq.Username = &req.Username
	}
	if req.FirstName != "" {
		pbReq.FirstName = &req.FirstName
	}
	if req.LastName != "" {
		pbReq.LastName = &req.LastName
	}
	if req.AvatarURL != "" {
		pbReq.AvatarUrl = &req.AvatarURL
	}

	resp, err := h.userClient.UpdateUser(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Failed to update user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DeleteUser handles DELETE /auth/users/{id}
func (h *AuthHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	pbReq := &userpb.DeleteUserRequest{
		Id: userID,
	}

	resp, err := h.userClient.DeleteUser(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Failed to delete user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UpdatePassword handles PUT /auth/users/{id}/password
func (h *AuthHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	var req struct {
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	pbReq := &userpb.UpdatePasswordRequest{
		UserId:       userID,
		PasswordHash: req.NewPassword,
	}

	resp, err := h.userClient.UpdatePassword(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Failed to update password: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ActivateUser handles POST /auth/users/{id}/activate
func (h *AuthHandler) ActivateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	pbReq := &userpb.ActivateUserRequest{
		UserId: userID,
	}

	resp, err := h.userClient.ActivateUser(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Failed to activate user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DeactivateUser handles POST /auth/users/{id}/deactivate
func (h *AuthHandler) DeactivateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	pbReq := &userpb.DeactivateUserRequest{
		UserId: userID,
	}

	resp, err := h.userClient.DeactivateUser(context.Background(), pbReq)
	if err != nil {
		http.Error(w, "Failed to deactivate user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
