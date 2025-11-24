package domain

import (
	"testing"
	"time"
)

func TestPermission_String(t *testing.T) {
	tests := []struct {
		name       string
		permission *Permission
		expected   string
	}{
		{
			name: "blog create permission",
			permission: &Permission{
				Resource: "blog",
				Action:   "create",
			},
			expected: "blog:create",
		},
		{
			name: "user delete permission",
			permission: &Permission{
				Resource: "user",
				Action:   "delete",
			},
			expected: "user:delete",
		},
		{
			name: "comment read permission",
			permission: &Permission{
				Resource: "comment",
				Action:   "read",
			},
			expected: "comment:read",
		},
		{
			name: "post update permission",
			permission: &Permission{
				Resource: "post",
				Action:   "update",
			},
			expected: "post:update",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.permission.String()
			if result != tt.expected {
				t.Errorf("String() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestPermission_BasicFields(t *testing.T) {
	now := time.Now()
	perm := &Permission{
		ID:          "test-id",
		Resource:    "blog",
		Action:      "create",
		Description: "Create blog posts",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if perm.ID != "test-id" {
		t.Errorf("expected ID test-id, got %s", perm.ID)
	}
	if perm.Resource != "blog" {
		t.Errorf("expected resource blog, got %s", perm.Resource)
	}
	if perm.Action != "create" {
		t.Errorf("expected action create, got %s", perm.Action)
	}
	if perm.Description != "Create blog posts" {
		t.Errorf("expected description 'Create blog posts', got %s", perm.Description)
	}
}

func TestPermission_DeletedAt(t *testing.T) {
	now := time.Now()

	t.Run("active permission", func(t *testing.T) {
		perm := &Permission{
			ID:        "test-id",
			DeletedAt: nil,
		}
		if perm.DeletedAt != nil {
			t.Error("DeletedAt should be nil for active permission")
		}
	})

	t.Run("deleted permission", func(t *testing.T) {
		perm := &Permission{
			ID:        "test-id",
			DeletedAt: &now,
		}
		if perm.DeletedAt == nil {
			t.Error("DeletedAt should not be nil for deleted permission")
		}
		if !perm.DeletedAt.Equal(now) {
			t.Errorf("expected DeletedAt %v, got %v", now, *perm.DeletedAt)
		}
	})
}

func TestPermission_Creation(t *testing.T) {
	perm := &Permission{
		ID:          "perm-1",
		Resource:    "user",
		Action:      "read",
		Description: "Read user information",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if perm.ID == "" {
		t.Error("ID should not be empty")
	}
	if perm.Resource == "" {
		t.Error("Resource should not be empty")
	}
	if perm.Action == "" {
		t.Error("Action should not be empty")
	}
	if perm.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if perm.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

func TestPermission_StringFormat(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		action   string
		expected string
	}{
		{"basic", "blog", "create", "blog:create"},
		{"update", "user", "update", "user:update"},
		{"delete", "comment", "delete", "comment:delete"},
		{"read", "post", "read", "post:read"},
		{"complex resource", "user-profile", "edit", "user-profile:edit"},
		{"complex action", "blog", "publish-draft", "blog:publish-draft"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perm := &Permission{
				Resource: tt.resource,
				Action:   tt.action,
			}
			result := perm.String()
			if result != tt.expected {
				t.Errorf("String() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestPermission_Update(t *testing.T) {
	now := time.Now()
	perm := &Permission{
		ID:          "test-id",
		Resource:    "blog",
		Action:      "create",
		Description: "Old description",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Simulate update
	laterTime := now.Add(1 * time.Hour)
	perm.Description = "New description"
	perm.UpdatedAt = laterTime

	if perm.Description != "New description" {
		t.Errorf("expected description 'New description', got %s", perm.Description)
	}
	if !perm.UpdatedAt.After(perm.CreatedAt) {
		t.Error("UpdatedAt should be after CreatedAt")
	}
}
