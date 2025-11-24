package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authpb "toxictoast/services/auth-service/api/proto"
	"toxictoast/services/auth-service/internal/usecase"
)

// AuthHandler implements the gRPC auth service
type AuthHandler struct {
	authpb.UnimplementedAuthServiceServer
	authUseCase       *usecase.AuthUseCase
	roleUseCase       *usecase.RoleUseCase
	permissionUseCase *usecase.PermissionUseCase
	rbacUseCase       *usecase.RBACUseCase
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	authUseCase *usecase.AuthUseCase,
	roleUseCase *usecase.RoleUseCase,
	permissionUseCase *usecase.PermissionUseCase,
	rbacUseCase *usecase.RBACUseCase,
) *AuthHandler {
	return &AuthHandler{
		authUseCase:       authUseCase,
		roleUseCase:       roleUseCase,
		permissionUseCase: permissionUseCase,
		rbacUseCase:       rbacUseCase,
	}
}

// Register creates a new user and returns JWT tokens
func (h *AuthHandler) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.AuthResponse, error) {
	firstName := ""
	if req.FirstName != nil {
		firstName = *req.FirstName
	}

	lastName := ""
	if req.LastName != nil {
		lastName = *req.LastName
	}

	accessToken, refreshToken, expiresIn, err := h.authUseCase.Register(
		ctx,
		req.Email,
		req.Username,
		req.Password,
		firstName,
		lastName,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register user: %v", err)
	}

	// Validate the token to get user claims
	claims, err := h.authUseCase.ValidateToken(ctx, accessToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to validate token: %v", err)
	}

	return &authpb.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		User: &authpb.UserClaims{
			UserId:      claims.UserID,
			Email:       claims.Email,
			Username:    claims.Username,
			Roles:       claims.Roles,
			Permissions: claims.Permissions,
		},
	}, nil
}

// Login authenticates a user and returns JWT tokens
func (h *AuthHandler) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.AuthResponse, error) {
	accessToken, refreshToken, expiresIn, err := h.authUseCase.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}

	// Validate the token to get user claims
	claims, err := h.authUseCase.ValidateToken(ctx, accessToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to validate token: %v", err)
	}

	return &authpb.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		User: &authpb.UserClaims{
			UserId:      claims.UserID,
			Email:       claims.Email,
			Username:    claims.Username,
			Roles:       claims.Roles,
			Permissions: claims.Permissions,
		},
	}, nil
}

// ValidateToken validates a JWT token and returns user info
func (h *AuthHandler) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	claims, err := h.authUseCase.ValidateToken(ctx, req.Token)
	if err != nil {
		return &authpb.ValidateTokenResponse{
			Valid: false,
		}, nil
	}

	return &authpb.ValidateTokenResponse{
		Valid: true,
		User: &authpb.UserClaims{
			UserId:      claims.UserID,
			Email:       claims.Email,
			Username:    claims.Username,
			Roles:       claims.Roles,
			Permissions: claims.Permissions,
		},
		Roles:       claims.Roles,
		Permissions: claims.Permissions,
	}, nil
}

// RefreshToken generates new tokens from a refresh token
func (h *AuthHandler) RefreshToken(ctx context.Context, req *authpb.RefreshTokenRequest) (*authpb.AuthResponse, error) {
	accessToken, refreshToken, expiresIn, err := h.authUseCase.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid refresh token")
	}

	// Validate the token to get user claims
	claims, err := h.authUseCase.ValidateToken(ctx, accessToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to validate token: %v", err)
	}

	return &authpb.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		User: &authpb.UserClaims{
			UserId:      claims.UserID,
			Email:       claims.Email,
			Username:    claims.Username,
			Roles:       claims.Roles,
			Permissions: claims.Permissions,
		},
	}, nil
}
