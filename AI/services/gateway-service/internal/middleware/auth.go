package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/toxictoast/toxictoastgo/shared/auth"
)

type contextKey string

const UserContextKey contextKey = "user"

// Auth middleware validates JWT tokens using Keycloak
func Auth(keycloakAuth *auth.KeycloakAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing authorization header", http.StatusUnauthorized)
				return
			}

			// Remove "Bearer " prefix
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == authHeader {
				http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
				return
			}

			// Validate token
			claims, err := keycloakAuth.ValidateToken(token)
			if err != nil {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Add user context to request
			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth validates token if present but allows requests without token
func OptionalAuth(keycloakAuth *auth.KeycloakAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				if token != authHeader {
					claims, err := keycloakAuth.ValidateToken(token)
					if err == nil {
						ctx := context.WithValue(r.Context(), UserContextKey, claims)
						r = r.WithContext(ctx)
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
