package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"toxictoast/services/auth-service/internal/domain"
	"toxictoast/services/auth-service/internal/repository/mocks"
)

func TestRoleUseCase_CreateRole(t *testing.T) {
	ctx := context.Background()

	t.Run("successful role creation", func(t *testing.T) {
		mockRepo := mocks.NewMockRoleRepository()
		uc := NewRoleUseCase(mockRepo)

		role, err := uc.CreateRole(ctx, "admin", "Administrator role")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if role == nil {
			t.Fatal("expected role to be created")
		}
		if role.Name != "admin" {
			t.Errorf("expected name admin, got %s", role.Name)
		}
		if role.Description != "Administrator role" {
			t.Errorf("expected description 'Administrator role', got %s", role.Description)
		}
		if role.ID == "" {
			t.Error("expected ID to be generated")
		}
	})

	t.Run("duplicate role name", func(t *testing.T) {
		mockRepo := mocks.NewMockRoleRepository()
		existingRole := &domain.Role{
			ID:        "existing-id",
			Name:      "admin",
			CreatedAt: time.Now(),
		}
		mockRepo.AddRole(existingRole)
		uc := NewRoleUseCase(mockRepo)

		_, err := uc.CreateRole(ctx, "admin", "Another admin role")

		if err == nil {
			t.Fatal("expected error for duplicate role name")
		}
		if err.Error() != "role with name 'admin' already exists" {
			t.Errorf("expected 'role with name 'admin' already exists' error, got %v", err)
		}
	})

	t.Run("repository error on name check", func(t *testing.T) {
		mockRepo := mocks.NewMockRoleRepository()
		mockRepo.GetByNameError = errors.New("database error")
		uc := NewRoleUseCase(mockRepo)

		_, err := uc.CreateRole(ctx, "admin", "Administrator role")

		if err == nil {
			t.Fatal("expected error from repository")
		}
	})

	t.Run("repository error on create", func(t *testing.T) {
		mockRepo := mocks.NewMockRoleRepository()
		mockRepo.CreateError = errors.New("database error")
		uc := NewRoleUseCase(mockRepo)

		_, err := uc.CreateRole(ctx, "admin", "Administrator role")

		if err == nil {
			t.Fatal("expected error from repository")
		}
	})
}

func TestRoleUseCase_GetRole(t *testing.T) {
	ctx := context.Background()

	t.Run("successful role retrieval", func(t *testing.T) {
		mockRepo := mocks.NewMockRoleRepository()
		existingRole := &domain.Role{
			ID:          "test-id",
			Name:        "admin",
			Description: "Administrator role",
			CreatedAt:   time.Now(),
		}
		mockRepo.AddRole(existingRole)
		uc := NewRoleUseCase(mockRepo)

		role, err := uc.GetRole(ctx, "test-id")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if role.ID != "test-id" {
			t.Errorf("expected ID test-id, got %s", role.ID)
		}
		if role.Name != "admin" {
			t.Errorf("expected name admin, got %s", role.Name)
		}
	})

	t.Run("role not found", func(t *testing.T) {
		mockRepo := mocks.NewMockRoleRepository()
		uc := NewRoleUseCase(mockRepo)

		_, err := uc.GetRole(ctx, "nonexistent-id")

		if err == nil {
			t.Fatal("expected error for nonexistent role")
		}
		if err.Error() != "role not found" {
			t.Errorf("expected 'role not found' error, got %v", err)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := mocks.NewMockRoleRepository()
		mockRepo.GetByIDError = errors.New("database error")
		uc := NewRoleUseCase(mockRepo)

		_, err := uc.GetRole(ctx, "test-id")

		if err == nil {
			t.Fatal("expected error from repository")
		}
	})
}

func TestRoleUseCase_ListRoles(t *testing.T) {
	ctx := context.Background()

	t.Run("successful list with pagination", func(t *testing.T) {
		mockRepo := mocks.NewMockRoleRepository()
		for i := 0; i < 5; i++ {
			mockRepo.AddRole(&domain.Role{
				ID:        string(rune(i)),
				Name:      "role" + string(rune(i)),
				CreatedAt: time.Now(),
			})
		}
		uc := NewRoleUseCase(mockRepo)

		roles, total, err := uc.ListRoles(ctx, 1, 10)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if total != 5 {
			t.Errorf("expected total 5, got %d", total)
		}
		if len(roles) != 5 {
			t.Errorf("expected 5 roles, got %d", len(roles))
		}
	})

	t.Run("pagination with default values", func(t *testing.T) {
		mockRepo := mocks.NewMockRoleRepository()
		uc := NewRoleUseCase(mockRepo)

		roles, total, err := uc.ListRoles(ctx, 0, 0)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if total != 0 {
			t.Errorf("expected total 0, got %d", total)
		}
		if len(roles) != 0 {
			t.Errorf("expected 0 roles, got %d", len(roles))
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := mocks.NewMockRoleRepository()
		mockRepo.ListError = errors.New("database error")
		uc := NewRoleUseCase(mockRepo)

		_, _, err := uc.ListRoles(ctx, 1, 10)

		if err == nil {
			t.Fatal("expected error from repository")
		}
	})
}

func TestRoleUseCase_UpdateRole(t *testing.T) {
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		mockRepo := mocks.NewMockRoleRepository()
		existingRole := &domain.Role{
			ID:          "test-id",
			Name:        "old-name",
			Description: "Old description",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		mockRepo.AddRole(existingRole)
		uc := NewRoleUseCase(mockRepo)

		newName := "new-name"
		newDesc := "New description"
		role, err := uc.UpdateRole(ctx, "test-id", &newName, &newDesc)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if role.Name != "new-name" {
			t.Errorf("expected name new-name, got %s", role.Name)
		}
		if role.Description != "New description" {
			t.Errorf("expected description 'New description', got %s", role.Description)
		}
	})

	t.Run("role not found", func(t *testing.T) {
		mockRepo := mocks.NewMockRoleRepository()
		uc := NewRoleUseCase(mockRepo)

		newName := "new-name"
		_, err := uc.UpdateRole(ctx, "nonexistent-id", &newName, nil)

		if err == nil {
			t.Fatal("expected error for nonexistent role")
		}
	})

	t.Run("duplicate name", func(t *testing.T) {
		mockRepo := mocks.NewMockRoleRepository()
		mockRepo.AddRole(&domain.Role{
			ID:        "role1",
			Name:      "role1",
			CreatedAt: time.Now(),
		})
		mockRepo.AddRole(&domain.Role{
			ID:        "role2",
			Name:      "role2",
			CreatedAt: time.Now(),
		})
		uc := NewRoleUseCase(mockRepo)

		newName := "role2"
		_, err := uc.UpdateRole(ctx, "role1", &newName, nil)

		if err == nil {
			t.Fatal("expected error for duplicate name")
		}
		if err.Error() != "role with name 'role2' already exists" {
			t.Errorf("expected 'role with name 'role2' already exists' error, got %v", err)
		}
	})

	t.Run("update only description", func(t *testing.T) {
		mockRepo := mocks.NewMockRoleRepository()
		existingRole := &domain.Role{
			ID:          "test-id",
			Name:        "admin",
			Description: "Old description",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		mockRepo.AddRole(existingRole)
		uc := NewRoleUseCase(mockRepo)

		newDesc := "New description"
		role, err := uc.UpdateRole(ctx, "test-id", nil, &newDesc)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if role.Name != "admin" {
			t.Errorf("expected name to remain admin, got %s", role.Name)
		}
		if role.Description != "New description" {
			t.Errorf("expected description 'New description', got %s", role.Description)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := mocks.NewMockRoleRepository()
		mockRepo.AddRole(&domain.Role{
			ID:        "test-id",
			Name:      "admin",
			CreatedAt: time.Now(),
		})
		mockRepo.UpdateError = errors.New("database error")
		uc := NewRoleUseCase(mockRepo)

		newName := "new-name"
		_, err := uc.UpdateRole(ctx, "test-id", &newName, nil)

		if err == nil {
			t.Fatal("expected error from repository")
		}
	})
}

func TestRoleUseCase_DeleteRole(t *testing.T) {
	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		mockRepo := mocks.NewMockRoleRepository()
		existingRole := &domain.Role{
			ID:        "test-id",
			Name:      "admin",
			CreatedAt: time.Now(),
		}
		mockRepo.AddRole(existingRole)
		uc := NewRoleUseCase(mockRepo)

		err := uc.DeleteRole(ctx, "test-id")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("role not found", func(t *testing.T) {
		mockRepo := mocks.NewMockRoleRepository()
		uc := NewRoleUseCase(mockRepo)

		err := uc.DeleteRole(ctx, "nonexistent-id")

		if err == nil {
			t.Fatal("expected error for nonexistent role")
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := mocks.NewMockRoleRepository()
		mockRepo.AddRole(&domain.Role{
			ID:        "test-id",
			Name:      "admin",
			CreatedAt: time.Now(),
		})
		mockRepo.DeleteError = errors.New("database error")
		uc := NewRoleUseCase(mockRepo)

		err := uc.DeleteRole(ctx, "test-id")

		if err == nil {
			t.Fatal("expected error from repository")
		}
	})
}
