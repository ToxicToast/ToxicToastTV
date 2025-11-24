package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"

	"github.com/toxictoast/toxictoastgo/shared/jwt"
)

const (
	// MetadataKeyUserID is the key for user ID in gRPC metadata
	MetadataKeyUserID = "x-user-id"
	// MetadataKeyEmail is the key for user email in gRPC metadata
	MetadataKeyEmail = "x-user-email"
	// MetadataKeyUsername is the key for username in gRPC metadata
	MetadataKeyUsername = "x-user-username"
	// MetadataKeyRoles is the key for user roles in gRPC metadata (comma-separated)
	MetadataKeyRoles = "x-user-roles"
	// MetadataKeyPermissions is the key for user permissions in gRPC metadata (comma-separated)
	MetadataKeyPermissions = "x-user-permissions"
)

// contextKey is a custom type for context keys
type contextKey string

const (
	// UserContextKey is the key for storing user info in context
	UserContextKey contextKey = "user"
)

// UserInfo contains extracted user information from gRPC metadata
type UserInfo struct {
	UserID      string
	Email       string
	Username    string
	Roles       []string
	Permissions []string
}

// InjectClaimsIntoMetadata adds JWT claims to outgoing gRPC metadata
func InjectClaimsIntoMetadata(ctx context.Context, claims *jwt.Claims) context.Context {
	if claims == nil {
		return ctx
	}

	md := metadata.Pairs(
		MetadataKeyUserID, claims.UserID,
		MetadataKeyEmail, claims.Email,
		MetadataKeyUsername, claims.Username,
	)

	// Add roles (comma-separated)
	if len(claims.Roles) > 0 {
		rolesStr := ""
		for i, role := range claims.Roles {
			if i > 0 {
				rolesStr += ","
			}
			rolesStr += role
		}
		md.Append(MetadataKeyRoles, rolesStr)
	}

	// Add permissions (comma-separated)
	if len(claims.Permissions) > 0 {
		permsStr := ""
		for i, perm := range claims.Permissions {
			if i > 0 {
				permsStr += ","
			}
			permsStr += perm
		}
		md.Append(MetadataKeyPermissions, permsStr)
	}

	return metadata.NewOutgoingContext(ctx, md)
}

// ExtractUserFromMetadata extracts user information from incoming gRPC metadata
func ExtractUserFromMetadata(ctx context.Context) (*UserInfo, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no metadata in context")
	}

	userID := getMetadataValue(md, MetadataKeyUserID)
	if userID == "" {
		return nil, fmt.Errorf("user ID not found in metadata")
	}

	email := getMetadataValue(md, MetadataKeyEmail)
	username := getMetadataValue(md, MetadataKeyUsername)

	// Parse roles
	rolesStr := getMetadataValue(md, MetadataKeyRoles)
	var roles []string
	if rolesStr != "" {
		roles = splitCommaSeparated(rolesStr)
	}

	// Parse permissions
	permsStr := getMetadataValue(md, MetadataKeyPermissions)
	var permissions []string
	if permsStr != "" {
		permissions = splitCommaSeparated(permsStr)
	}

	return &UserInfo{
		UserID:      userID,
		Email:       email,
		Username:    username,
		Roles:       roles,
		Permissions: permissions,
	}, nil
}

// GetUserFromContext extracts user info from context (after interceptor)
func GetUserFromContext(ctx context.Context) (*UserInfo, bool) {
	user, ok := ctx.Value(UserContextKey).(*UserInfo)
	return user, ok
}

// InjectUserIntoContext stores user info in context
func InjectUserIntoContext(ctx context.Context, user *UserInfo) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}

// getMetadataValue extracts a single value from metadata
func getMetadataValue(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// splitCommaSeparated splits a comma-separated string into a slice
func splitCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}
	result := []string{}
	parts := splitString(s, ',')
	for _, part := range parts {
		trimmed := trimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// splitString splits a string by a delimiter
func splitString(s string, delimiter rune) []string {
	var result []string
	var current string
	for _, c := range s {
		if c == delimiter {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" || len(result) > 0 {
		result = append(result, current)
	}
	return result
}

// trimSpace removes leading and trailing whitespace
func trimSpace(s string) string {
	start := 0
	end := len(s)

	// Trim leading whitespace
	for start < end && isSpace(s[start]) {
		start++
	}

	// Trim trailing whitespace
	for end > start && isSpace(s[end-1]) {
		end--
	}

	return s[start:end]
}

// isSpace checks if a byte is a whitespace character
func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}
