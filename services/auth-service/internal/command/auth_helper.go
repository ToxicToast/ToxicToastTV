package command

import (
	"context"
	"fmt"

	"github.com/toxictoast/toxictoastgo/shared/jwt"
	"toxictoast/services/auth-service/internal/repository/interfaces"
)

// TokenHelper generates JWT tokens with roles and permissions
type TokenHelper struct {
	jwtHelper          *jwt.JWTHelper
	userRoleRepo       interfaces.UserRoleRepository
	rolePermissionRepo interfaces.RolePermissionRepository
}

// NewTokenHelper creates a new token helper
func NewTokenHelper(
	jwtHelper *jwt.JWTHelper,
	userRoleRepo interfaces.UserRoleRepository,
	rolePermissionRepo interfaces.RolePermissionRepository,
) *TokenHelper {
	return &TokenHelper{
		jwtHelper:          jwtHelper,
		userRoleRepo:       userRoleRepo,
		rolePermissionRepo: rolePermissionRepo,
	}
}

// GenerateTokens generates access and refresh tokens for a user
func (h *TokenHelper) GenerateTokens(ctx context.Context, userID, email, username string) (string, string, int64, error) {
	// Get user roles
	roles, err := h.userRoleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to get user roles: %w", err)
	}

	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}

	// Get user permissions
	permissions, err := h.rolePermissionRepo.GetUserPermissions(ctx, userID)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to get user permissions: %w", err)
	}

	permissionStrings := make([]string, 0, len(permissions))
	for _, perm := range permissions {
		permissionStrings = append(permissionStrings, perm.String())
	}

	// Generate access token
	accessToken, err := h.jwtHelper.GenerateAccessToken(userID, email, username, roleNames, permissionStrings)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := h.jwtHelper.GenerateRefreshToken(userID)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	expiresIn := int64(h.jwtHelper.GetAccessTokenDuration().Seconds())

	return accessToken, refreshToken, expiresIn, nil
}
