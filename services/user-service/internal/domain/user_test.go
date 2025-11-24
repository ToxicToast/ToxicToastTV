package domain

import (
	"testing"
	"time"
)

func TestUser_FullName(t *testing.T) {
	tests := []struct {
		name      string
		user      *User
		expected  string
	}{
		{
			name: "both first and last name",
			user: &User{
				FirstName: "John",
				LastName:  "Doe",
				Username:  "johndoe",
			},
			expected: "John Doe",
		},
		{
			name: "only first name",
			user: &User{
				FirstName: "John",
				Username:  "johndoe",
			},
			expected: "John",
		},
		{
			name: "only last name",
			user: &User{
				LastName:  "Doe",
				Username:  "johndoe",
			},
			expected: "Doe",
		},
		{
			name: "no names, return username",
			user: &User{
				Username: "johndoe",
			},
			expected: "johndoe",
		},
		{
			name: "empty strings, return username",
			user: &User{
				FirstName: "",
				LastName:  "",
				Username:  "johndoe",
			},
			expected: "johndoe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.FullName()
			if result != tt.expected {
				t.Errorf("FullName() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestUser_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		status   UserStatus
		expected bool
	}{
		{
			name:     "active user",
			status:   UserStatusActive,
			expected: true,
		},
		{
			name:     "inactive user",
			status:   UserStatusInactive,
			expected: false,
		},
		{
			name:     "suspended user",
			status:   UserStatusSuspended,
			expected: false,
		},
		{
			name:     "deleted user",
			status:   UserStatusDeleted,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				Status: tt.status,
			}
			result := user.IsActive()
			if result != tt.expected {
				t.Errorf("IsActive() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestUserStatus_Constants(t *testing.T) {
	tests := []struct {
		name     string
		status   UserStatus
		expected string
	}{
		{"active status", UserStatusActive, "active"},
		{"inactive status", UserStatusInactive, "inactive"},
		{"suspended status", UserStatusSuspended, "suspended"},
		{"deleted status", UserStatusDeleted, "deleted"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("Status = %v, expected %v", tt.status, tt.expected)
			}
		})
	}
}

func TestUser_CreatedAt(t *testing.T) {
	now := time.Now()
	user := &User{
		ID:        "test-id",
		Email:     "test@example.com",
		Username:  "testuser",
		CreatedAt: now,
	}

	if !user.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, expected %v", user.CreatedAt, now)
	}
}

func TestUser_LastLogin(t *testing.T) {
	now := time.Now()

	t.Run("user with last login", func(t *testing.T) {
		user := &User{
			LastLogin: &now,
		}
		if user.LastLogin == nil {
			t.Error("LastLogin should not be nil")
		}
		if !user.LastLogin.Equal(now) {
			t.Errorf("LastLogin = %v, expected %v", user.LastLogin, now)
		}
	})

	t.Run("user without last login", func(t *testing.T) {
		user := &User{
			LastLogin: nil,
		}
		if user.LastLogin != nil {
			t.Error("LastLogin should be nil")
		}
	})
}

func TestUser_DeletedAt(t *testing.T) {
	now := time.Now()

	t.Run("deleted user", func(t *testing.T) {
		user := &User{
			DeletedAt: &now,
			Status:    UserStatusDeleted,
		}
		if user.DeletedAt == nil {
			t.Error("DeletedAt should not be nil for deleted user")
		}
	})

	t.Run("active user", func(t *testing.T) {
		user := &User{
			DeletedAt: nil,
			Status:    UserStatusActive,
		}
		if user.DeletedAt != nil {
			t.Error("DeletedAt should be nil for active user")
		}
	})
}
