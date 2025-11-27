package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"github.com/toxictoast/toxictoastgo/shared/jwt"
	"github.com/toxictoast/toxictoastgo/shared/kafka"
	authpb "toxictoast/services/auth-service/api/proto"
	"toxictoast/services/auth-service/internal/command"
	"toxictoast/services/auth-service/internal/query"
	"toxictoast/services/auth-service/internal/repository/interfaces"
)

// AuthHandler implements the gRPC auth service
type AuthHandler struct {
	authpb.UnimplementedAuthServiceServer
	commandBus    *cqrs.CommandBus
	queryBus      *cqrs.QueryBus
	tokenHelper   *command.TokenHelper
	kafkaProducer *kafka.Producer
}

// NewAuthHandler creates a new auth handler with CQRS support
func NewAuthHandler(
	commandBus *cqrs.CommandBus,
	queryBus *cqrs.QueryBus,
	jwtHelper *jwt.JWTHelper,
	userRoleRepo interfaces.UserRoleRepository,
	rolePermissionRepo interfaces.RolePermissionRepository,
	kafkaProducer *kafka.Producer,
) *AuthHandler {
	return &AuthHandler{
		commandBus:    commandBus,
		queryBus:      queryBus,
		tokenHelper:   command.NewTokenHelper(jwtHelper, userRoleRepo, rolePermissionRepo),
		kafkaProducer: kafkaProducer,
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

	// Create command
	cmd := &command.RegisterCommand{
		BaseCommand: cqrs.BaseCommand{},
		Email:       req.Email,
		Username:    req.Username,
		Password:    req.Password,
		FirstName:   firstName,
		LastName:    lastName,
	}

	// Dispatch command (creates user in user-service)
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register user: %v", err)
	}

	// Get user ID from command (set by handler)
	userID := cmd.AggregateID

	// Generate tokens
	accessToken, refreshToken, expiresIn, err := h.tokenHelper.GenerateTokens(ctx, userID, req.Email, req.Username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate tokens: %v", err)
	}

	// Validate the token to get user claims
	validateQuery := &query.ValidateTokenQuery{
		BaseQuery: cqrs.BaseQuery{},
		Token:     accessToken,
	}

	result, err := h.queryBus.Dispatch(ctx, validateQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to validate token: %v", err)
	}

	claims := result.(*jwt.Claims)

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
	// Create command
	cmd := &command.LoginCommand{
		BaseCommand: cqrs.BaseCommand{},
		Email:       req.Email,
		Password:    req.Password,
	}

	// Dispatch command (validates credentials via user-service)
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}

	// Get user info from command (set by handler)
	userID := cmd.AggregateID
	username := cmd.Username

	// Generate tokens
	accessToken, refreshToken, expiresIn, err := h.tokenHelper.GenerateTokens(ctx, userID, req.Email, username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate tokens: %v", err)
	}

	// Publish Kafka event
	if h.kafkaProducer != nil {
		eventData := map[string]interface{}{
			"user_id":   userID,
			"email":     req.Email,
			"logged_at": time.Now(),
		}
		h.kafkaProducer.PublishEvent("auth.login", userID, eventData)
	}

	// Validate the token to get user claims
	validateQuery := &query.ValidateTokenQuery{
		BaseQuery: cqrs.BaseQuery{},
		Token:     accessToken,
	}

	result, err := h.queryBus.Dispatch(ctx, validateQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to validate token: %v", err)
	}

	claims := result.(*jwt.Claims)

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
	qry := &query.ValidateTokenQuery{
		BaseQuery: cqrs.BaseQuery{},
		Token:     req.Token,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return &authpb.ValidateTokenResponse{
			Valid: false,
		}, nil
	}

	claims := result.(*jwt.Claims)

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
	// Create command
	cmd := &command.RefreshTokenCommand{
		BaseCommand:  cqrs.BaseCommand{},
		RefreshToken: req.RefreshToken,
	}

	// Dispatch command (validates refresh token)
	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid refresh token")
	}

	// Get user info from command (set by handler)
	userID := cmd.AggregateID
	email := cmd.Email
	username := cmd.Username

	// Generate new tokens
	accessToken, refreshToken, expiresIn, err := h.tokenHelper.GenerateTokens(ctx, userID, email, username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate tokens: %v", err)
	}

	// Publish Kafka event
	if h.kafkaProducer != nil {
		eventData := map[string]interface{}{
			"user_id":      userID,
			"refreshed_at": time.Now(),
		}
		h.kafkaProducer.PublishEvent("auth.token.refreshed", userID, eventData)
	}

	// Validate the token to get user claims
	validateQuery := &query.ValidateTokenQuery{
		BaseQuery: cqrs.BaseQuery{},
		Token:     accessToken,
	}

	result, err := h.queryBus.Dispatch(ctx, validateQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to validate token: %v", err)
	}

	claims := result.(*jwt.Claims)

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
