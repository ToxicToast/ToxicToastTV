package domain

import (
	"testing"
	"time"
)

func TestRole_BasicFields(t *testing.T) {
	now := time.Now()
	role := &Role{
		ID:          "test-id",
		Name:        "admin",
		Description: "Administrator role",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if role.ID != "test-id" {
		t.Errorf("expected ID test-id, got %s", role.ID)
	}
	if role.Name != "admin" {
		t.Errorf("expected name admin, got %s", role.Name)
	}
	if role.Description != "Administrator role" {
		t.Errorf("expected description 'Administrator role', got %s", role.Description)
	}
	if !role.CreatedAt.Equal(now) {
		t.Errorf("expected CreatedAt %v, got %v", now, role.CreatedAt)
	}
}

func TestRole_DeletedAt(t *testing.T) {
	now := time.Now()

	t.Run("active role", func(t *testing.T) {
		role := &Role{
			ID:        "test-id",
			DeletedAt: nil,
		}
		if role.DeletedAt != nil {
			t.Error("DeletedAt should be nil for active role")
		}
	})

	t.Run("deleted role", func(t *testing.T) {
		role := &Role{
			ID:        "test-id",
			DeletedAt: &now,
		}
		if role.DeletedAt == nil {
			t.Error("DeletedAt should not be nil for deleted role")
		}
		if !role.DeletedAt.Equal(now) {
			t.Errorf("expected DeletedAt %v, got %v", now, *role.DeletedAt)
		}
	})
}

func TestRole_Creation(t *testing.T) {
	role := &Role{
		ID:          "role-1",
		Name:        "user",
		Description: "Regular user role",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if role.ID == "" {
		t.Error("ID should not be empty")
	}
	if role.Name == "" {
		t.Error("Name should not be empty")
	}
	if role.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if role.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

func TestRole_Update(t *testing.T) {
	now := time.Now()
	role := &Role{
		ID:          "test-id",
		Name:        "user",
		Description: "Old description",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Simulate update
	laterTime := now.Add(1 * time.Hour)
	role.Description = "New description"
	role.UpdatedAt = laterTime

	if role.Description != "New description" {
		t.Errorf("expected description 'New description', got %s", role.Description)
	}
	if !role.UpdatedAt.After(role.CreatedAt) {
		t.Error("UpdatedAt should be after CreatedAt")
	}
}
