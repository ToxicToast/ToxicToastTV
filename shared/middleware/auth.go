package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/toxictoast/toxictoastgo/shared/auth"
	"github.com/toxictoast/toxictoastgo/shared/jwt"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// ClaimsContextKey is the key for storing JWT claims in context
	ClaimsContextKey contextKey = "jwt_claims"
)

// AuthMiddleware provides JWT authentication middleware for HTTP handlers
type AuthMiddleware struct {
	jwtHelper      *jwt.JWTHelper
	tokenBlacklist *auth.TokenBlacklist
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(jwtHelper *jwt.JWTHelper) *AuthMiddleware {
	return &AuthMiddleware{
		jwtHelper:      jwtHelper,
		tokenBlacklist: auth.NewTokenBlacklist(),
	}
}

// GetTokenBlacklist returns the token blacklist instance
func (m *AuthMiddleware) GetTokenBlacklist() *auth.TokenBlacklist {
	return m.tokenBlacklist
}

// Authenticate is a middleware that validates JWT tokens from the Authorization header
// It extracts the token, validates it, and stores the claims in the request context
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		// Check if token is revoked
		if m.tokenBlacklist.IsRevoked(tokenString) {
			http.Error(w, "Token has been revoked", http.StatusUnauthorized)
			return
		}

		// Validate token
		claims, err := m.jwtHelper.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Store claims in context
		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AuthenticateOptional is a middleware that validates JWT tokens if present
// Unlike Authenticate, it doesn't fail if the token is missing
func (m *AuthMiddleware) AuthenticateOptional(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No token provided - continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Check for Bearer token format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Invalid format - continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := m.jwtHelper.ValidateToken(tokenString)
		if err != nil {
			// Invalid token - continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Store claims in context
		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole is a middleware that checks if the authenticated user has a specific role
// Must be used after Authenticate middleware
func (m *AuthMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r.Context())
			if claims == nil {
				http.Error(w, "Unauthorized: authentication required", http.StatusUnauthorized)
				return
			}

			if !HasRole(claims, role) {
				http.Error(w, "Forbidden: required role '"+role+"' not found", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyRole is a middleware that checks if the authenticated user has any of the specified roles
// Must be used after Authenticate middleware
func (m *AuthMiddleware) RequireAnyRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r.Context())
			if claims == nil {
				http.Error(w, "Unauthorized: authentication required", http.StatusUnauthorized)
				return
			}

			for _, role := range roles {
				if HasRole(claims, role) {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, "Forbidden: none of the required roles found", http.StatusForbidden)
		})
	}
}

// RequirePermission is a middleware that checks if the authenticated user has a specific permission
// Must be used after Authenticate middleware
func (m *AuthMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r.Context())
			if claims == nil {
				http.Error(w, "Unauthorized: authentication required", http.StatusUnauthorized)
				return
			}

			if !HasPermission(claims, permission) {
				http.Error(w, "Forbidden: required permission '"+permission+"' not found", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission is a middleware that checks if the authenticated user has any of the specified permissions
// Must be used after Authenticate middleware
func (m *AuthMiddleware) RequireAnyPermission(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r.Context())
			if claims == nil {
				http.Error(w, "Unauthorized: authentication required", http.StatusUnauthorized)
				return
			}

			for _, permission := range permissions {
				if HasPermission(claims, permission) {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, "Forbidden: none of the required permissions found", http.StatusForbidden)
		})
	}
}

// GetClaims extracts JWT claims from the request context
func GetClaims(ctx context.Context) *jwt.Claims {
	claims, ok := ctx.Value(ClaimsContextKey).(*jwt.Claims)
	if !ok {
		return nil
	}
	return claims
}

// HasRole checks if the claims contain a specific role
func HasRole(claims *jwt.Claims, role string) bool {
	if claims == nil {
		return false
	}
	for _, r := range claims.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasPermission checks if the claims contain a specific permission
func HasPermission(claims *jwt.Claims, permission string) bool {
	if claims == nil {
		return false
	}
	for _, p := range claims.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the claims contain any of the specified roles
func HasAnyRole(claims *jwt.Claims, roles ...string) bool {
	if claims == nil {
		return false
	}
	for _, role := range roles {
		if HasRole(claims, role) {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if the claims contain any of the specified permissions
func HasAnyPermission(claims *jwt.Claims, permissions ...string) bool {
	if claims == nil {
		return false
	}
	for _, permission := range permissions {
		if HasPermission(claims, permission) {
			return true
		}
	}
	return false
}
