package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"toxictoast/services/user-service/internal/domain"
	"toxictoast/services/user-service/internal/repository/mocks"
)

func TestUserUseCase_CreateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("successful user creation", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		uc := NewUserUseCase(mockRepo, nil)

		firstName := "John"
		lastName := "Doe"
		user, err := uc.CreateUser(ctx, "john@example.com", "johndoe", "password123", &firstName, &lastName, nil)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if user == nil {
			t.Fatal("expected user to be created")
		}
		if user.Email != "john@example.com" {
			t.Errorf("expected email john@example.com, got %s", user.Email)
		}
		if user.Username != "johndoe" {
			t.Errorf("expected username johndoe, got %s", user.Username)
		}
		if user.FirstName != "John" {
			t.Errorf("expected first name John, got %s", user.FirstName)
		}
		if user.Status != domain.UserStatusActive {
			t.Errorf("expected status active, got %s", user.Status)
		}
		// Verify password was hashed
		if user.PasswordHash == "password123" {
			t.Error("password should be hashed")
		}
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("password123"))
		if err != nil {
			t.Error("password hash verification failed")
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		existingUser := &domain.User{
			ID:       "existing-id",
			Email:    "john@example.com",
			Username: "existinguser",
			Status:   domain.UserStatusActive,
		}
		mockRepo.AddUser(existingUser)
		uc := NewUserUseCase(mockRepo, nil)

		_, err := uc.CreateUser(ctx, "john@example.com", "newuser", "password123", nil, nil, nil)

		if err == nil {
			t.Fatal("expected error for duplicate email")
		}
		if err.Error() != "email already exists" {
			t.Errorf("expected 'email already exists' error, got %v", err)
		}
	})

	t.Run("duplicate username", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		existingUser := &domain.User{
			ID:       "existing-id",
			Email:    "existing@example.com",
			Username: "johndoe",
			Status:   domain.UserStatusActive,
		}
		mockRepo.AddUser(existingUser)
		uc := NewUserUseCase(mockRepo, nil)

		_, err := uc.CreateUser(ctx, "john@example.com", "johndoe", "password123", nil, nil, nil)

		if err == nil {
			t.Fatal("expected error for duplicate username")
		}
		if err.Error() != "username already exists" {
			t.Errorf("expected 'username already exists' error, got %v", err)
		}
	})

	t.Run("repository error on email check", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		mockRepo.GetByEmailError = errors.New("database error")
		uc := NewUserUseCase(mockRepo, nil)

		_, err := uc.CreateUser(ctx, "john@example.com", "johndoe", "password123", nil, nil, nil)

		if err == nil {
			t.Fatal("expected error from repository")
		}
	})

	t.Run("repository error on create", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		mockRepo.CreateError = errors.New("database error")
		uc := NewUserUseCase(mockRepo, nil)

		_, err := uc.CreateUser(ctx, "john@example.com", "johndoe", "password123", nil, nil, nil)

		if err == nil {
			t.Fatal("expected error from repository")
		}
	})
}

func TestUserUseCase_GetUser(t *testing.T) {
	ctx := context.Background()

	t.Run("successful user retrieval", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		existingUser := &domain.User{
			ID:       "test-id",
			Email:    "john@example.com",
			Username: "johndoe",
			Status:   domain.UserStatusActive,
		}
		mockRepo.AddUser(existingUser)
		uc := NewUserUseCase(mockRepo, nil)

		user, err := uc.GetUser(ctx, "test-id")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if user.ID != "test-id" {
			t.Errorf("expected ID test-id, got %s", user.ID)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		uc := NewUserUseCase(mockRepo, nil)

		_, err := uc.GetUser(ctx, "nonexistent-id")

		if err == nil {
			t.Fatal("expected error for nonexistent user")
		}
		if err.Error() != "user not found" {
			t.Errorf("expected 'user not found' error, got %v", err)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		mockRepo.GetByIDError = errors.New("database error")
		uc := NewUserUseCase(mockRepo, nil)

		_, err := uc.GetUser(ctx, "test-id")

		if err == nil {
			t.Fatal("expected error from repository")
		}
	})
}

func TestUserUseCase_GetUserByEmail(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval by email", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		existingUser := &domain.User{
			ID:       "test-id",
			Email:    "john@example.com",
			Username: "johndoe",
		}
		mockRepo.AddUser(existingUser)
		uc := NewUserUseCase(mockRepo, nil)

		user, err := uc.GetUserByEmail(ctx, "john@example.com")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if user.Email != "john@example.com" {
			t.Errorf("expected email john@example.com, got %s", user.Email)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		uc := NewUserUseCase(mockRepo, nil)

		_, err := uc.GetUserByEmail(ctx, "nonexistent@example.com")

		if err == nil {
			t.Fatal("expected error for nonexistent user")
		}
	})
}

func TestUserUseCase_GetUserByUsername(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval by username", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		existingUser := &domain.User{
			ID:       "test-id",
			Email:    "john@example.com",
			Username: "johndoe",
		}
		mockRepo.AddUser(existingUser)
		uc := NewUserUseCase(mockRepo, nil)

		user, err := uc.GetUserByUsername(ctx, "johndoe")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if user.Username != "johndoe" {
			t.Errorf("expected username johndoe, got %s", user.Username)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		uc := NewUserUseCase(mockRepo, nil)

		_, err := uc.GetUserByUsername(ctx, "nonexistent")

		if err == nil {
			t.Fatal("expected error for nonexistent user")
		}
	})
}

func TestUserUseCase_ListUsers(t *testing.T) {
	ctx := context.Background()

	t.Run("successful list with pagination", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		for i := 0; i < 5; i++ {
			mockRepo.AddUser(&domain.User{
				ID:       string(rune(i)),
				Email:    "user" + string(rune(i)) + "@example.com",
				Username: "user" + string(rune(i)),
				Status:   domain.UserStatusActive,
			})
		}
		uc := NewUserUseCase(mockRepo, nil)

		users, total, err := uc.ListUsers(ctx, 1, 10, nil, nil, "", "")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if total != 5 {
			t.Errorf("expected total 5, got %d", total)
		}
		if len(users) != 5 {
			t.Errorf("expected 5 users, got %d", len(users))
		}
	})

	t.Run("pagination with default values", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		uc := NewUserUseCase(mockRepo, nil)

		users, total, err := uc.ListUsers(ctx, 0, 0, nil, nil, "", "")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if total != 0 {
			t.Errorf("expected total 0, got %d", total)
		}
		if len(users) != 0 {
			t.Errorf("expected 0 users, got %d", len(users))
		}
	})

	t.Run("filter by status", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		mockRepo.AddUser(&domain.User{
			ID:     "1",
			Status: domain.UserStatusActive,
		})
		mockRepo.AddUser(&domain.User{
			ID:     "2",
			Status: domain.UserStatusInactive,
		})
		uc := NewUserUseCase(mockRepo, nil)

		status := domain.UserStatusActive
		users, total, err := uc.ListUsers(ctx, 1, 10, &status, nil, "", "")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
		if len(users) != 1 {
			t.Errorf("expected 1 user, got %d", len(users))
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		mockRepo.ListError = errors.New("database error")
		uc := NewUserUseCase(mockRepo, nil)

		_, _, err := uc.ListUsers(ctx, 1, 10, nil, nil, "", "")

		if err == nil {
			t.Fatal("expected error from repository")
		}
	})
}

func TestUserUseCase_UpdateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		existingUser := &domain.User{
			ID:        "test-id",
			Email:     "old@example.com",
			Username:  "olduser",
			FirstName: "Old",
			LastName:  "User",
			CreatedAt: time.Now(),
		}
		mockRepo.AddUser(existingUser)
		uc := NewUserUseCase(mockRepo, nil)

		newEmail := "new@example.com"
		newUsername := "newuser"
		newFirstName := "New"
		user, err := uc.UpdateUser(ctx, "test-id", &newEmail, &newUsername, &newFirstName, nil, nil)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if user.Email != "new@example.com" {
			t.Errorf("expected email new@example.com, got %s", user.Email)
		}
		if user.Username != "newuser" {
			t.Errorf("expected username newuser, got %s", user.Username)
		}
		if user.FirstName != "New" {
			t.Errorf("expected first name New, got %s", user.FirstName)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		uc := NewUserUseCase(mockRepo, nil)

		newEmail := "new@example.com"
		_, err := uc.UpdateUser(ctx, "nonexistent-id", &newEmail, nil, nil, nil, nil)

		if err == nil {
			t.Fatal("expected error for nonexistent user")
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		mockRepo.AddUser(&domain.User{
			ID:       "user1",
			Email:    "user1@example.com",
			Username: "user1",
		})
		mockRepo.AddUser(&domain.User{
			ID:       "user2",
			Email:    "user2@example.com",
			Username: "user2",
		})
		uc := NewUserUseCase(mockRepo, nil)

		newEmail := "user2@example.com"
		_, err := uc.UpdateUser(ctx, "user1", &newEmail, nil, nil, nil, nil)

		if err == nil {
			t.Fatal("expected error for duplicate email")
		}
		if err.Error() != "email already exists" {
			t.Errorf("expected 'email already exists' error, got %v", err)
		}
	})

	t.Run("duplicate username", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		mockRepo.AddUser(&domain.User{
			ID:       "user1",
			Email:    "user1@example.com",
			Username: "user1",
		})
		mockRepo.AddUser(&domain.User{
			ID:       "user2",
			Email:    "user2@example.com",
			Username: "user2",
		})
		uc := NewUserUseCase(mockRepo, nil)

		newUsername := "user2"
		_, err := uc.UpdateUser(ctx, "user1", nil, &newUsername, nil, nil, nil)

		if err == nil {
			t.Fatal("expected error for duplicate username")
		}
		if err.Error() != "username already exists" {
			t.Errorf("expected 'username already exists' error, got %v", err)
		}
	})
}

func TestUserUseCase_UpdatePassword(t *testing.T) {
	ctx := context.Background()

	t.Run("successful password update", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)
		existingUser := &domain.User{
			ID:           "test-id",
			Email:        "john@example.com",
			PasswordHash: string(hashedPassword),
		}
		mockRepo.AddUser(existingUser)
		uc := NewUserUseCase(mockRepo, nil)

		err := uc.UpdatePassword(ctx, "test-id", "newpassword")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify password was updated
		updatedUser, _ := mockRepo.GetByID(ctx, "test-id")
		err = bcrypt.CompareHashAndPassword([]byte(updatedUser.PasswordHash), []byte("newpassword"))
		if err != nil {
			t.Error("new password hash verification failed")
		}
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		uc := NewUserUseCase(mockRepo, nil)

		err := uc.UpdatePassword(ctx, "nonexistent-id", "newpassword")

		if err == nil {
			t.Fatal("expected error for nonexistent user")
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		mockRepo.AddUser(&domain.User{
			ID:    "test-id",
			Email: "john@example.com",
		})
		mockRepo.UpdatePasswordError = errors.New("database error")
		uc := NewUserUseCase(mockRepo, nil)

		err := uc.UpdatePassword(ctx, "test-id", "newpassword")

		if err == nil {
			t.Fatal("expected error from repository")
		}
	})
}

func TestUserUseCase_VerifyPassword(t *testing.T) {
	ctx := context.Background()

	t.Run("correct password", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
		existingUser := &domain.User{
			ID:           "test-id",
			PasswordHash: string(hashedPassword),
		}
		mockRepo.AddUser(existingUser)
		uc := NewUserUseCase(mockRepo, nil)

		valid, err := uc.VerifyPassword(ctx, "test-id", "correctpassword")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !valid {
			t.Error("expected password to be valid")
		}
	})

	t.Run("incorrect password", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
		existingUser := &domain.User{
			ID:           "test-id",
			PasswordHash: string(hashedPassword),
		}
		mockRepo.AddUser(existingUser)
		uc := NewUserUseCase(mockRepo, nil)

		valid, err := uc.VerifyPassword(ctx, "test-id", "wrongpassword")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if valid {
			t.Error("expected password to be invalid")
		}
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		uc := NewUserUseCase(mockRepo, nil)

		_, err := uc.VerifyPassword(ctx, "nonexistent-id", "password")

		if err == nil {
			t.Fatal("expected error for nonexistent user")
		}
	})
}

func TestUserUseCase_ActivateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("successful activation", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		existingUser := &domain.User{
			ID:        "test-id",
			Email:     "john@example.com",
			Username:  "johndoe",
			Status:    domain.UserStatusInactive,
			CreatedAt: time.Now(),
		}
		mockRepo.AddUser(existingUser)
		uc := NewUserUseCase(mockRepo, nil)

		user, err := uc.ActivateUser(ctx, "test-id")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if user.Status != domain.UserStatusActive {
			t.Errorf("expected status active, got %s", user.Status)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		uc := NewUserUseCase(mockRepo, nil)

		_, err := uc.ActivateUser(ctx, "nonexistent-id")

		if err == nil {
			t.Fatal("expected error for nonexistent user")
		}
	})
}

func TestUserUseCase_DeactivateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("successful deactivation", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		existingUser := &domain.User{
			ID:        "test-id",
			Email:     "john@example.com",
			Username:  "johndoe",
			Status:    domain.UserStatusActive,
			CreatedAt: time.Now(),
		}
		mockRepo.AddUser(existingUser)
		uc := NewUserUseCase(mockRepo, nil)

		user, err := uc.DeactivateUser(ctx, "test-id")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if user.Status != domain.UserStatusInactive {
			t.Errorf("expected status inactive, got %s", user.Status)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		uc := NewUserUseCase(mockRepo, nil)

		_, err := uc.DeactivateUser(ctx, "nonexistent-id")

		if err == nil {
			t.Fatal("expected error for nonexistent user")
		}
	})
}

func TestUserUseCase_DeleteUser(t *testing.T) {
	ctx := context.Background()

	t.Run("soft delete", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		existingUser := &domain.User{
			ID:        "test-id",
			Email:     "john@example.com",
			Username:  "johndoe",
			Status:    domain.UserStatusActive,
			CreatedAt: time.Now(),
		}
		mockRepo.AddUser(existingUser)
		uc := NewUserUseCase(mockRepo, nil)

		err := uc.DeleteUser(ctx, "test-id", false)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		// User should still exist but be marked deleted
		if mockRepo.UserCount() != 1 {
			t.Error("expected user to still exist after soft delete")
		}
	})

	t.Run("hard delete", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		existingUser := &domain.User{
			ID:        "test-id",
			Email:     "john@example.com",
			Username:  "johndoe",
			Status:    domain.UserStatusActive,
			CreatedAt: time.Now(),
		}
		mockRepo.AddUser(existingUser)
		uc := NewUserUseCase(mockRepo, nil)

		err := uc.DeleteUser(ctx, "test-id", true)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		// User should be completely removed
		if mockRepo.UserCount() != 0 {
			t.Error("expected user to be removed after hard delete")
		}
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository()
		uc := NewUserUseCase(mockRepo, nil)

		err := uc.DeleteUser(ctx, "nonexistent-id", false)

		if err == nil {
			t.Fatal("expected error for nonexistent user")
		}
	})
}
