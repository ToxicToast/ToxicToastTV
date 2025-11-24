package usecase

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/toxictoast/toxictoastgo/shared/jwt"
	"github.com/toxictoast/toxictoastgo/shared/kafka"
	"toxictoast/services/auth-service/internal/repository/interfaces"
	userpb "toxictoast/services/user-service/api/proto"
)

// AuthUseCase handles authentication business logic
type AuthUseCase struct {
	userRoleRepo       interfaces.UserRoleRepository
	rolePermissionRepo interfaces.RolePermissionRepository
	jwtHelper          *jwt.JWTHelper
	userServiceAddr    string
	kafkaProducer      *kafka.Producer
}

// NewAuthUseCase creates a new auth use case
func NewAuthUseCase(
	userRoleRepo interfaces.UserRoleRepository,
	rolePermissionRepo interfaces.RolePermissionRepository,
	jwtHelper *jwt.JWTHelper,
	userServiceAddr string,
	kafkaProducer *kafka.Producer,
) *AuthUseCase {
	return &AuthUseCase{
		userRoleRepo:       userRoleRepo,
		rolePermissionRepo: rolePermissionRepo,
		jwtHelper:          jwtHelper,
		userServiceAddr:    userServiceAddr,
		kafkaProducer:      kafkaProducer,
	}
}

// Register creates a new user and returns JWT tokens
func (uc *AuthUseCase) Register(ctx context.Context, email, username, password, firstName, lastName string) (string, string, int64, error) {
	// Connect to user-service
	conn, err := grpc.NewClient(uc.userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to connect to user-service: %w", err)
	}
	defer conn.Close()

	userClient := userpb.NewUserServiceClient(conn)

	// Create user via user-service (user-service will handle password hashing)
	createReq := &userpb.CreateUserRequest{
		Email:     email,
		Username:  username,
		Password:  password,
		FirstName: &firstName,
		LastName:  &lastName,
	}

	userResp, err := userClient.CreateUser(ctx, createReq)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	accessToken, refreshToken, expiresIn, err := uc.generateTokens(ctx, userResp.User.Id, userResp.User.Email, userResp.User.Username)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		eventData := map[string]interface{}{
			"user_id":       userResp.User.Id,
			"email":         email,
			"username":      username,
			"registered_at": time.Now(),
		}
		uc.kafkaProducer.PublishEvent("auth.registered", userResp.User.Id, eventData)
	}

	return accessToken, refreshToken, expiresIn, nil
}

// Login authenticates a user and returns JWT tokens
func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (string, string, int64, error) {
	// Connect to user-service
	conn, err := grpc.NewClient(uc.userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to connect to user-service: %w", err)
	}
	defer conn.Close()

	userClient := userpb.NewUserServiceClient(conn)

	// Get user by email
	userResp, err := userClient.GetUserByEmail(ctx, &userpb.GetUserByEmailRequest{Email: email})
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid credentials")
	}

	if userResp.User == nil {
		return "", "", 0, fmt.Errorf("invalid credentials")
	}

	// Verify password (pass plain password, user-service will compare with bcrypt)
	verifyResp, err := userClient.VerifyPassword(ctx, &userpb.VerifyPasswordRequest{
		UserId:       userResp.User.Id,
		PasswordHash: password,
	})
	if err != nil || !verifyResp.Valid {
		return "", "", 0, fmt.Errorf("invalid credentials")
	}

	// Check if user is active
	if userResp.User.Status != userpb.UserStatus_USER_STATUS_ACTIVE {
		return "", "", 0, fmt.Errorf("user account is not active")
	}

	// Update last login
	_, err = userClient.ActivateUser(ctx, &userpb.ActivateUserRequest{UserId: userResp.User.Id})
	if err != nil {
		// Log but don't fail login
	}

	// Generate tokens
	accessToken, refreshToken, expiresIn, err := uc.generateTokens(ctx, userResp.User.Id, userResp.User.Email, userResp.User.Username)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		eventData := map[string]interface{}{
			"user_id":   userResp.User.Id,
			"email":     email,
			"username":  userResp.User.Username,
			"logged_at": time.Now(),
		}
		uc.kafkaProducer.PublishEvent("auth.login", userResp.User.Id, eventData)
	}

	return accessToken, refreshToken, expiresIn, nil
}

// ValidateToken validates a JWT token and returns user info
func (uc *AuthUseCase) ValidateToken(ctx context.Context, token string) (*jwt.Claims, error) {
	claims, err := uc.jwtHelper.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return claims, nil
}

// RefreshToken generates new tokens from a refresh token
func (uc *AuthUseCase) RefreshToken(ctx context.Context, refreshToken string) (string, string, int64, error) {
	// Validate refresh token
	claims, err := uc.jwtHelper.ValidateToken(refreshToken)
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Connect to user-service to get fresh user data
	conn, err := grpc.NewClient(uc.userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to connect to user-service: %w", err)
	}
	defer conn.Close()

	userClient := userpb.NewUserServiceClient(conn)
	userResp, err := userClient.GetUser(ctx, &userpb.GetUserRequest{Id: claims.UserID})
	if err != nil {
		return "", "", 0, fmt.Errorf("user not found")
	}

	// Generate new tokens
	accessToken, refreshToken, expiresIn, err := uc.generateTokens(ctx, userResp.User.Id, userResp.User.Email, userResp.User.Username)
	if err != nil {
		return "", "", 0, err
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		eventData := map[string]interface{}{
			"user_id":      userResp.User.Id,
			"email":        userResp.User.Email,
			"username":     userResp.User.Username,
			"refreshed_at": time.Now(),
		}
		uc.kafkaProducer.PublishEvent("auth.token.refreshed", userResp.User.Id, eventData)
	}

	return accessToken, refreshToken, expiresIn, nil
}

// generateTokens is a helper to generate access and refresh tokens
func (uc *AuthUseCase) generateTokens(ctx context.Context, userID, email, username string) (string, string, int64, error) {
	// Get user roles
	roles, err := uc.userRoleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to get user roles: %w", err)
	}

	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}

	// Get user permissions
	permissions, err := uc.rolePermissionRepo.GetUserPermissions(ctx, userID)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to get user permissions: %w", err)
	}

	permissionStrings := make([]string, 0, len(permissions))
	for _, perm := range permissions {
		permissionStrings = append(permissionStrings, perm.String())
	}

	// Generate access token
	accessToken, err := uc.jwtHelper.GenerateAccessToken(userID, email, username, roleNames, permissionStrings)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := uc.jwtHelper.GenerateRefreshToken(userID)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	expiresIn := int64(uc.jwtHelper.GetAccessTokenDuration().Seconds())

	return accessToken, refreshToken, expiresIn, nil
}
