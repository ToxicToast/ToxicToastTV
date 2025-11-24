package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/toxictoast/toxictoastgo/shared/kafka"
	"toxictoast/services/user-service/internal/domain"
	"toxictoast/services/user-service/internal/repository/interfaces"
)

// UserUseCase handles user business logic
type UserUseCase struct {
	userRepo      interfaces.UserRepository
	kafkaProducer *kafka.Producer
}

// NewUserUseCase creates a new user use case
func NewUserUseCase(userRepo interfaces.UserRepository, kafkaProducer *kafka.Producer) *UserUseCase {
	return &UserUseCase{
		userRepo:      userRepo,
		kafkaProducer: kafkaProducer,
	}
}

// CreateUser creates a new user
func (uc *UserUseCase) CreateUser(ctx context.Context, email, username, password string, firstName, lastName, avatarURL *string) (*domain.User, error) {
	// Check if email already exists
	existing, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing email: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("email already exists")
	}

	// Check if username already exists
	existing, err = uc.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing username: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("username already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		ID:           uuid.New().String(),
		Email:        email,
		Username:     username,
		PasswordHash: string(hashedPassword),
		Status:       domain.UserStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if firstName != nil {
		user.FirstName = *firstName
	}
	if lastName != nil {
		user.LastName = *lastName
	}
	if avatarURL != nil {
		user.AvatarURL = *avatarURL
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		eventData := map[string]interface{}{
			"user_id":    user.ID,
			"email":      user.Email,
			"username":   user.Username,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"status":     user.Status,
			"created_at": user.CreatedAt,
		}
		uc.kafkaProducer.PublishEvent("user.created", user.ID, eventData)
	}

	return user, nil
}

// GetUser retrieves a user by ID
func (uc *UserUseCase) GetUser(ctx context.Context, id string) (*domain.User, error) {
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (uc *UserUseCase) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

// GetUserByUsername retrieves a user by username
func (uc *UserUseCase) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	user, err := uc.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

// ListUsers retrieves users with pagination and filters
func (uc *UserUseCase) ListUsers(ctx context.Context, page, pageSize int, status *domain.UserStatus, search *string, sortBy, sortOrder string) ([]*domain.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	users, total, err := uc.userRepo.List(ctx, offset, pageSize, status, search, sortBy, sortOrder)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

// UpdateUser updates an existing user
func (uc *UserUseCase) UpdateUser(ctx context.Context, id string, email, username, firstName, lastName, avatarURL *string) (*domain.User, error) {
	// Get existing user
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update email
	if email != nil && *email != "" && *email != user.Email {
		// Check if new email already exists
		existing, err := uc.userRepo.GetByEmail(ctx, *email)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing email: %w", err)
		}
		if existing != nil && existing.ID != id {
			return nil, fmt.Errorf("email already exists")
		}
		user.Email = *email
	}

	// Update username
	if username != nil && *username != "" && *username != user.Username {
		// Check if new username already exists
		existing, err := uc.userRepo.GetByUsername(ctx, *username)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing username: %w", err)
		}
		if existing != nil && existing.ID != id {
			return nil, fmt.Errorf("username already exists")
		}
		user.Username = *username
	}

	// Update other fields
	if firstName != nil {
		user.FirstName = *firstName
	}
	if lastName != nil {
		user.LastName = *lastName
	}
	if avatarURL != nil {
		user.AvatarURL = *avatarURL
	}

	user.UpdatedAt = time.Now()

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		eventData := map[string]interface{}{
			"user_id":    user.ID,
			"email":      user.Email,
			"username":   user.Username,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"avatar_url": user.AvatarURL,
			"updated_at": user.UpdatedAt,
		}
		uc.kafkaProducer.PublishEvent("user.updated", user.ID, eventData)
	}

	return user, nil
}

// UpdatePassword updates a user's password
func (uc *UserUseCase) UpdatePassword(ctx context.Context, userID, password string) error {
	// Check if user exists
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Hash password before storing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := uc.userRepo.UpdatePassword(ctx, userID, string(hashedPassword)); err != nil {
		return err
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		eventData := map[string]interface{}{
			"user_id":    userID,
			"changed_at": time.Now(),
		}
		uc.kafkaProducer.PublishEvent("user.password.changed", userID, eventData)
	}

	return nil
}

// VerifyPassword checks if a password hash matches the user's password
func (uc *UserUseCase) VerifyPassword(ctx context.Context, userID, password string) (bool, error) {
	// Get user
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return false, fmt.Errorf("user not found")
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return false, nil
	}

	return true, nil
}

// ActivateUser activates a user account
func (uc *UserUseCase) ActivateUser(ctx context.Context, userID string) (*domain.User, error) {
	// Get user
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update status and last login
	if err := uc.userRepo.UpdateStatus(ctx, userID, domain.UserStatusActive); err != nil {
		return nil, fmt.Errorf("failed to activate user: %w", err)
	}

	if err := uc.userRepo.UpdateLastLogin(ctx, userID); err != nil {
		// Log but don't fail
	}

	// Get updated user
	updatedUser, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		eventData := map[string]interface{}{
			"user_id":      userID,
			"email":        updatedUser.Email,
			"username":     updatedUser.Username,
			"activated_at": time.Now(),
		}
		uc.kafkaProducer.PublishEvent("user.activated", userID, eventData)
	}

	return updatedUser, nil
}

// DeactivateUser deactivates a user account
func (uc *UserUseCase) DeactivateUser(ctx context.Context, userID string) (*domain.User, error) {
	// Get user
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update status
	if err := uc.userRepo.UpdateStatus(ctx, userID, domain.UserStatusInactive); err != nil {
		return nil, fmt.Errorf("failed to deactivate user: %w", err)
	}

	// Get updated user
	updatedUser, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		eventData := map[string]interface{}{
			"user_id":        userID,
			"email":          updatedUser.Email,
			"username":       updatedUser.Username,
			"deactivated_at": time.Now(),
		}
		uc.kafkaProducer.PublishEvent("user.deactivated", userID, eventData)
	}

	return updatedUser, nil
}

// DeleteUser deletes a user (soft or hard delete)
func (uc *UserUseCase) DeleteUser(ctx context.Context, id string, hardDelete bool) error {
	// Check if user exists
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	if hardDelete {
		if err = uc.userRepo.HardDelete(ctx, id); err != nil {
			return err
		}
	} else {
		if err = uc.userRepo.Delete(ctx, id); err != nil {
			return err
		}
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		eventData := map[string]interface{}{
			"user_id":     id,
			"email":       user.Email,
			"username":    user.Username,
			"hard_delete": hardDelete,
			"deleted_at":  time.Now(),
		}
		uc.kafkaProducer.PublishEvent("user.deleted", id, eventData)
	}

	return nil
}
