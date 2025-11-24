package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/toxictoast/toxictoastgo/shared/middleware"
)

// ProtectedHandler demonstrates protected endpoints using JWT middleware
type ProtectedHandler struct{}

// NewProtectedHandler creates a new protected handler
func NewProtectedHandler() *ProtectedHandler {
	return &ProtectedHandler{}
}

// RegisterRoutes registers all protected test routes
func (h *ProtectedHandler) RegisterRoutes(router *mux.Router, authMiddleware *middleware.AuthMiddleware) {
	// Public endpoint - no authentication required
	router.HandleFunc("/public", h.Public).Methods("GET")

	// Protected endpoint - requires valid JWT token
	protectedRoute := router.PathPrefix("/protected").Subrouter()
	protectedRoute.Use(authMiddleware.Authenticate)
	protectedRoute.HandleFunc("", h.Protected).Methods("GET")

	// Admin only endpoint - requires 'admin' role
	adminRoute := router.PathPrefix("/admin").Subrouter()
	adminRoute.Use(authMiddleware.Authenticate)
	adminRoute.Use(authMiddleware.RequireRole("admin"))
	adminRoute.HandleFunc("", h.AdminOnly).Methods("GET")

	// Editor endpoint - requires 'editor' OR 'admin' role
	editorRoute := router.PathPrefix("/editor").Subrouter()
	editorRoute.Use(authMiddleware.Authenticate)
	editorRoute.Use(authMiddleware.RequireAnyRole("editor", "admin"))
	editorRoute.HandleFunc("", h.EditorOnly).Methods("GET")

	// Write permission endpoint - requires 'posts.write' permission
	writeRoute := router.PathPrefix("/write").Subrouter()
	writeRoute.Use(authMiddleware.Authenticate)
	writeRoute.Use(authMiddleware.RequirePermission("posts.write"))
	writeRoute.HandleFunc("", h.WriteOnly).Methods("GET")
}

// Public endpoint accessible without authentication
func (h *ProtectedHandler) Public(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"message": "This is a public endpoint",
		"authenticated": false,
	}

	// Check if user is authenticated (optional authentication)
	claims := middleware.GetClaims(r.Context())
	if claims != nil {
		response["authenticated"] = true
		response["user_id"] = claims.UserID
		response["username"] = claims.Username
	}

	json.NewEncoder(w).Encode(response)
}

// Protected endpoint requiring valid JWT token
func (h *ProtectedHandler) Protected(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Unauthorized",
			"message": "Valid JWT token required",
		})
		return
	}

	response := map[string]interface{}{
		"message": "Welcome to the protected endpoint",
		"user": map[string]interface{}{
			"user_id":     claims.UserID,
			"email":       claims.Email,
			"username":    claims.Username,
			"roles":       claims.Roles,
			"permissions": claims.Permissions,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// AdminOnly endpoint requiring 'admin' role
func (h *ProtectedHandler) AdminOnly(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Unauthorized",
			"message": "Valid JWT token required",
		})
		return
	}

	response := map[string]interface{}{
		"message": "Welcome to the admin-only endpoint",
		"user": map[string]interface{}{
			"user_id":  claims.UserID,
			"username": claims.Username,
			"roles":    claims.Roles,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// EditorOnly endpoint requiring 'editor' or 'admin' role
func (h *ProtectedHandler) EditorOnly(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Unauthorized",
			"message": "Valid JWT token required",
		})
		return
	}

	response := map[string]interface{}{
		"message": "Welcome to the editor endpoint",
		"user": map[string]interface{}{
			"user_id":  claims.UserID,
			"username": claims.Username,
			"roles":    claims.Roles,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// WriteOnly endpoint requiring 'posts.write' permission
func (h *ProtectedHandler) WriteOnly(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Unauthorized",
			"message": "Valid JWT token required",
		})
		return
	}

	response := map[string]interface{}{
		"message": "You have write permissions",
		"user": map[string]interface{}{
			"user_id":     claims.UserID,
			"username":    claims.Username,
			"permissions": claims.Permissions,
		},
	}

	json.NewEncoder(w).Encode(response)
}
