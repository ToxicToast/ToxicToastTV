package command

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"github.com/toxictoast/toxictoastgo/shared/eventstore"
	"toxictoast/services/user-service/internal/aggregate"
	"toxictoast/services/user-service/internal/repository/interfaces"
)

// CreateUserCommand creates a new user
type CreateUserCommand struct {
	cqrs.BaseCommand
	Email     string  `json:"email"`
	Username  string  `json:"username"`
	Password  string  `json:"password"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

func (c *CreateUserCommand) CommandName() string {
	return "create_user"
}

func (c *CreateUserCommand) Validate() error {
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

// ChangeEmailCommand changes a user's email
type ChangeEmailCommand struct {
	cqrs.BaseCommand
	NewEmail string `json:"new_email"`
}

func (c *ChangeEmailCommand) CommandName() string {
	return "change_email"
}

func (c *ChangeEmailCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("user_id is required")
	}
	if c.NewEmail == "" {
		return errors.New("new_email is required")
	}
	return nil
}

// ChangePasswordCommand changes a user's password
type ChangePasswordCommand struct {
	cqrs.BaseCommand
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (c *ChangePasswordCommand) CommandName() string {
	return "change_password"
}

func (c *ChangePasswordCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("user_id is required")
	}
	if c.NewPassword == "" {
		return errors.New("new_password is required")
	}
	return nil
}

// UpdatePasswordHashCommand updates a user's password with a pre-hashed password
// This is used when the password is already hashed by auth-service
type UpdatePasswordHashCommand struct {
	cqrs.BaseCommand
	NewPasswordHash string `json:"new_password_hash"`
}

func (c *UpdatePasswordHashCommand) CommandName() string {
	return "update_password_hash"
}

func (c *UpdatePasswordHashCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("user_id is required")
	}
	if c.NewPasswordHash == "" {
		return errors.New("new_password_hash is required")
	}
	return nil
}

// UpdateProfileCommand updates a user's profile
type UpdateProfileCommand struct {
	cqrs.BaseCommand
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

func (c *UpdateProfileCommand) CommandName() string {
	return "update_profile"
}

func (c *UpdateProfileCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("user_id is required")
	}
	return nil
}

// ActivateUserCommand activates a user
type ActivateUserCommand struct {
	cqrs.BaseCommand
}

func (c *ActivateUserCommand) CommandName() string {
	return "activate_user"
}

func (c *ActivateUserCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("user_id is required")
	}
	return nil
}

// DeactivateUserCommand deactivates a user
type DeactivateUserCommand struct {
	cqrs.BaseCommand
}

func (c *DeactivateUserCommand) CommandName() string {
	return "deactivate_user"
}

func (c *DeactivateUserCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("user_id is required")
	}
	return nil
}

// DeleteUserCommand soft deletes a user
type DeleteUserCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteUserCommand) CommandName() string {
	return "delete_user"
}

func (c *DeleteUserCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("user_id is required")
	}
	return nil
}

// Command Handlers

// CreateUserHandler handles user creation
type CreateUserHandler struct {
	aggRepo  *eventstore.AggregateRepository
	userRepo interfaces.UserRepository // For uniqueness checks
}

func NewCreateUserHandler(aggRepo *eventstore.AggregateRepository, userRepo interfaces.UserRepository) *CreateUserHandler {
	return &CreateUserHandler{
		aggRepo:  aggRepo,
		userRepo: userRepo,
	}
}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	createCmd := cmd.(*CreateUserCommand)

	// Check uniqueness (using existing repository)
	existing, err := h.userRepo.GetByEmail(ctx, createCmd.Email)
	if err != nil {
		return fmt.Errorf("failed to check email uniqueness: %w", err)
	}
	if existing != nil {
		return aggregate.ErrEmailAlreadyExists
	}

	existing, err = h.userRepo.GetByUsername(ctx, createCmd.Username)
	if err != nil {
		return fmt.Errorf("failed to check username uniqueness: %w", err)
	}
	if existing != nil {
		return aggregate.ErrUsernameAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(createCmd.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create aggregate
	userID := uuid.New().String()
	user := aggregate.NewUserAggregate(userID)

	if err := user.CreateUser(
		createCmd.Email,
		createCmd.Username,
		string(hashedPassword),
		createCmd.FirstName,
		createCmd.LastName,
		createCmd.AvatarURL,
	); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Save events
	if err := h.aggRepo.Save(ctx, user); err != nil {
		return fmt.Errorf("failed to save user events: %w", err)
	}

	return nil
}

// ChangeEmailHandler handles email changes
type ChangeEmailHandler struct {
	aggRepo  *eventstore.AggregateRepository
	userRepo interfaces.UserRepository
}

func NewChangeEmailHandler(aggRepo *eventstore.AggregateRepository, userRepo interfaces.UserRepository) *ChangeEmailHandler {
	return &ChangeEmailHandler{
		aggRepo:  aggRepo,
		userRepo: userRepo,
	}
}

func (h *ChangeEmailHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	changeCmd := cmd.(*ChangeEmailCommand)

	// Check email uniqueness
	existing, err := h.userRepo.GetByEmail(ctx, changeCmd.NewEmail)
	if err != nil {
		return fmt.Errorf("failed to check email uniqueness: %w", err)
	}
	if existing != nil && existing.ID != changeCmd.AggregateID {
		return aggregate.ErrEmailAlreadyExists
	}

	// Load aggregate
	user := aggregate.NewUserAggregate(changeCmd.AggregateID)
	if err := h.aggRepo.Load(ctx, user); err != nil {
		return fmt.Errorf("failed to load user: %w", err)
	}

	// Change email
	if err := user.ChangeEmail(changeCmd.NewEmail); err != nil {
		return fmt.Errorf("failed to change email: %w", err)
	}

	// Save events
	if err := h.aggRepo.Save(ctx, user); err != nil {
		return fmt.Errorf("failed to save user events: %w", err)
	}

	return nil
}

// ChangePasswordHandler handles password changes
type ChangePasswordHandler struct {
	aggRepo *eventstore.AggregateRepository
}

func NewChangePasswordHandler(aggRepo *eventstore.AggregateRepository) *ChangePasswordHandler {
	return &ChangePasswordHandler{
		aggRepo: aggRepo,
	}
}

func (h *ChangePasswordHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	changeCmd := cmd.(*ChangePasswordCommand)

	// Load aggregate
	user := aggregate.NewUserAggregate(changeCmd.AggregateID)
	if err := h.aggRepo.Load(ctx, user); err != nil {
		return fmt.Errorf("failed to load user: %w", err)
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(changeCmd.OldPassword)); err != nil {
		return errors.New("invalid old password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(changeCmd.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Change password
	if err := user.ChangePassword(string(hashedPassword)); err != nil {
		return fmt.Errorf("failed to change password: %w", err)
	}

	// Save events
	if err := h.aggRepo.Save(ctx, user); err != nil {
		return fmt.Errorf("failed to save user events: %w", err)
	}

	return nil
}

// UpdatePasswordHashHandler handles password updates with pre-hashed passwords
type UpdatePasswordHashHandler struct {
	aggRepo *eventstore.AggregateRepository
}

func NewUpdatePasswordHashHandler(aggRepo *eventstore.AggregateRepository) *UpdatePasswordHashHandler {
	return &UpdatePasswordHashHandler{
		aggRepo: aggRepo,
	}
}

func (h *UpdatePasswordHashHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdatePasswordHashCommand)

	// Load aggregate
	user := aggregate.NewUserAggregate(updateCmd.AggregateID)
	if err := h.aggRepo.Load(ctx, user); err != nil {
		return fmt.Errorf("failed to load user: %w", err)
	}

	// Change password (already hashed)
	if err := user.ChangePassword(updateCmd.NewPasswordHash); err != nil {
		return fmt.Errorf("failed to change password: %w", err)
	}

	// Save events
	if err := h.aggRepo.Save(ctx, user); err != nil {
		return fmt.Errorf("failed to save user events: %w", err)
	}

	return nil
}

// UpdateProfileHandler handles profile updates
type UpdateProfileHandler struct {
	aggRepo *eventstore.AggregateRepository
}

func NewUpdateProfileHandler(aggRepo *eventstore.AggregateRepository) *UpdateProfileHandler {
	return &UpdateProfileHandler{
		aggRepo: aggRepo,
	}
}

func (h *UpdateProfileHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateProfileCommand)

	// Load aggregate
	user := aggregate.NewUserAggregate(updateCmd.AggregateID)
	if err := h.aggRepo.Load(ctx, user); err != nil {
		return fmt.Errorf("failed to load user: %w", err)
	}

	// Update profile
	if err := user.UpdateProfile(updateCmd.FirstName, updateCmd.LastName, updateCmd.AvatarURL); err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	// Save events
	if err := h.aggRepo.Save(ctx, user); err != nil {
		return fmt.Errorf("failed to save user events: %w", err)
	}

	return nil
}

// ActivateUserHandler handles user activation
type ActivateUserHandler struct {
	aggRepo *eventstore.AggregateRepository
}

func NewActivateUserHandler(aggRepo *eventstore.AggregateRepository) *ActivateUserHandler {
	return &ActivateUserHandler{
		aggRepo: aggRepo,
	}
}

func (h *ActivateUserHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	activateCmd := cmd.(*ActivateUserCommand)

	// Load aggregate
	user := aggregate.NewUserAggregate(activateCmd.AggregateID)
	if err := h.aggRepo.Load(ctx, user); err != nil {
		return fmt.Errorf("failed to load user: %w", err)
	}

	// Activate
	if err := user.Activate(); err != nil {
		return fmt.Errorf("failed to activate user: %w", err)
	}

	// Save events
	if err := h.aggRepo.Save(ctx, user); err != nil {
		return fmt.Errorf("failed to save user events: %w", err)
	}

	return nil
}

// DeactivateUserHandler handles user deactivation
type DeactivateUserHandler struct {
	aggRepo *eventstore.AggregateRepository
}

func NewDeactivateUserHandler(aggRepo *eventstore.AggregateRepository) *DeactivateUserHandler {
	return &DeactivateUserHandler{
		aggRepo: aggRepo,
	}
}

func (h *DeactivateUserHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deactivateCmd := cmd.(*DeactivateUserCommand)

	// Load aggregate
	user := aggregate.NewUserAggregate(deactivateCmd.AggregateID)
	if err := h.aggRepo.Load(ctx, user); err != nil {
		return fmt.Errorf("failed to load user: %w", err)
	}

	// Deactivate
	if err := user.Deactivate(); err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	// Save events
	if err := h.aggRepo.Save(ctx, user); err != nil {
		return fmt.Errorf("failed to save user events: %w", err)
	}

	return nil
}

// DeleteUserHandler handles user deletion
type DeleteUserHandler struct {
	aggRepo *eventstore.AggregateRepository
}

func NewDeleteUserHandler(aggRepo *eventstore.AggregateRepository) *DeleteUserHandler {
	return &DeleteUserHandler{
		aggRepo: aggRepo,
	}
}

func (h *DeleteUserHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteUserCommand)

	// Load aggregate
	user := aggregate.NewUserAggregate(deleteCmd.AggregateID)
	if err := h.aggRepo.Load(ctx, user); err != nil {
		return fmt.Errorf("failed to load user: %w", err)
	}

	// Delete
	if err := user.Delete(); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Save events
	if err := h.aggRepo.Save(ctx, user); err != nil {
		return fmt.Errorf("failed to save user events: %w", err)
	}

	return nil
}
