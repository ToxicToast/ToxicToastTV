package command

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"github.com/toxictoast/toxictoastgo/shared/jwt"
	"github.com/toxictoast/toxictoastgo/shared/kafka"
	"toxictoast/services/auth-service/internal/repository/interfaces"
	userpb "toxictoast/services/user-service/api/proto"
)

// RegisterCommand registers a new user
type RegisterCommand struct {
	cqrs.BaseCommand
	Email     string `json:"email"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (c *RegisterCommand) CommandName() string {
	return "register"
}

func (c *RegisterCommand) Validate() error {
	if c.Email == "" {
		return errors.New("email is required")
	}
	if c.Username == "" {
		return errors.New("username is required")
	}
	if c.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

// LoginCommand authenticates a user
type LoginCommand struct {
	cqrs.BaseCommand
	Email    string `json:"email"`
	Password string `json:"password"`
	// Populated by handler
	Username string `json:"-"`
}

func (c *LoginCommand) CommandName() string {
	return "login"
}

func (c *LoginCommand) Validate() error {
	if c.Email == "" {
		return errors.New("email is required")
	}
	if c.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

// RefreshTokenCommand generates new tokens from a refresh token
type RefreshTokenCommand struct {
	cqrs.BaseCommand
	RefreshToken string `json:"refresh_token"`
	// Populated by handler
	Email    string `json:"-"`
	Username string `json:"-"`
}

func (c *RefreshTokenCommand) CommandName() string {
	return "refresh_token"
}

func (c *RefreshTokenCommand) Validate() error {
	if c.RefreshToken == "" {
		return errors.New("refresh_token is required")
	}
	return nil
}

// AuthResult contains the result of authentication commands
type AuthResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	UserID       string
}

// Command Handlers

// RegisterHandler handles user registration
type RegisterHandler struct {
	userRoleRepo       interfaces.UserRoleRepository
	rolePermissionRepo interfaces.RolePermissionRepository
	jwtHelper          *jwt.JWTHelper
	userServiceAddr    string
	kafkaProducer      *kafka.Producer
}

func NewRegisterHandler(
	userRoleRepo interfaces.UserRoleRepository,
	rolePermissionRepo interfaces.RolePermissionRepository,
	jwtHelper *jwt.JWTHelper,
	userServiceAddr string,
	kafkaProducer *kafka.Producer,
) *RegisterHandler {
	return &RegisterHandler{
		userRoleRepo:       userRoleRepo,
		rolePermissionRepo: rolePermissionRepo,
		jwtHelper:          jwtHelper,
		userServiceAddr:    userServiceAddr,
		kafkaProducer:      kafkaProducer,
	}
}

func (h *RegisterHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	registerCmd := cmd.(*RegisterCommand)

	// Connect to user-service
	conn, err := grpc.NewClient(h.userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to user-service: %w", err)
	}
	defer conn.Close()

	userClient := userpb.NewUserServiceClient(conn)

	// Create user via user-service (user-service will handle password hashing)
	createReq := &userpb.CreateUserRequest{
		Email:     registerCmd.Email,
		Username:  registerCmd.Username,
		Password:  registerCmd.Password,
		FirstName: &registerCmd.FirstName,
		LastName:  &registerCmd.LastName,
	}

	userResp, err := userClient.CreateUser(ctx, createReq)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Publish Kafka event
	if h.kafkaProducer != nil {
		eventData := map[string]interface{}{
			"user_id":       userResp.User.Id,
			"email":         registerCmd.Email,
			"username":      registerCmd.Username,
			"registered_at": time.Now(),
		}
		h.kafkaProducer.PublishEvent("auth.registered", userResp.User.Id, eventData)
	}

	// Store user ID in command result (for handler to generate tokens)
	cmd.(*RegisterCommand).AggregateID = userResp.User.Id

	return nil
}

// LoginHandler handles user authentication
type LoginHandler struct {
	userServiceAddr string
}

func NewLoginHandler(userServiceAddr string) *LoginHandler {
	return &LoginHandler{userServiceAddr: userServiceAddr}
}

func (h *LoginHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	loginCmd := cmd.(*LoginCommand)

	// Connect to user-service
	conn, err := grpc.NewClient(h.userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to user-service: %w", err)
	}
	defer conn.Close()

	userClient := userpb.NewUserServiceClient(conn)

	// Get user by email
	userResp, err := userClient.GetUserByEmail(ctx, &userpb.GetUserByEmailRequest{Email: loginCmd.Email})
	if err != nil {
		return fmt.Errorf("invalid credentials")
	}

	if userResp.User == nil {
		return fmt.Errorf("invalid credentials")
	}

	// Verify password (pass plain password, user-service will compare with bcrypt)
	verifyResp, err := userClient.VerifyPassword(ctx, &userpb.VerifyPasswordRequest{
		UserId:       userResp.User.Id,
		PasswordHash: loginCmd.Password,
	})
	if err != nil || !verifyResp.Valid {
		return fmt.Errorf("invalid credentials")
	}

	// Check if user is active
	if userResp.User.Status != userpb.UserStatus_USER_STATUS_ACTIVE {
		return fmt.Errorf("user account is not active")
	}

	// Update last login
	_, err = userClient.ActivateUser(ctx, &userpb.ActivateUserRequest{UserId: userResp.User.Id})
	if err != nil {
		// Log but don't fail login
	}

	// Store user info in command for token generation
	loginCmd.AggregateID = userResp.User.Id
	loginCmd.Username = userResp.User.Username

	return nil
}

// RefreshTokenHandler handles token refresh
type RefreshTokenHandler struct {
	jwtHelper       *jwt.JWTHelper
	userServiceAddr string
}

func NewRefreshTokenHandler(jwtHelper *jwt.JWTHelper, userServiceAddr string) *RefreshTokenHandler {
	return &RefreshTokenHandler{
		jwtHelper:       jwtHelper,
		userServiceAddr: userServiceAddr,
	}
}

func (h *RefreshTokenHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	refreshCmd := cmd.(*RefreshTokenCommand)

	// Validate refresh token
	claims, err := h.jwtHelper.ValidateToken(refreshCmd.RefreshToken)
	if err != nil {
		return fmt.Errorf("invalid refresh token: %w", err)
	}

	// Connect to user-service to get fresh user data
	conn, err := grpc.NewClient(h.userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to user-service: %w", err)
	}
	defer conn.Close()

	userClient := userpb.NewUserServiceClient(conn)
	userResp, err := userClient.GetUser(ctx, &userpb.GetUserRequest{Id: claims.UserID})
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Store user info in command for token generation
	refreshCmd.AggregateID = userResp.User.Id
	refreshCmd.Email = userResp.User.Email
	refreshCmd.Username = userResp.User.Username

	return nil
}
