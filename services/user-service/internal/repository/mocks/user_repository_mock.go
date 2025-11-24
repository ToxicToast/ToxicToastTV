package mocks

import (
	"context"
	"sync"

	"toxictoast/services/user-service/internal/domain"
)

// MockUserRepository is a mock implementation of UserRepository for testing
type MockUserRepository struct {
	mu    sync.RWMutex
	users map[string]*domain.User

	// Error injection for testing
	CreateError           error
	GetByIDError          error
	GetByEmailError       error
	GetByUsernameError    error
	ListError             error
	UpdateError           error
	UpdatePasswordError   error
	UpdateLastLoginError  error
	UpdateStatusError     error
	DeleteError           error
	HardDeleteError       error
}

// NewMockUserRepository creates a new mock repository
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*domain.User),
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	if m.CreateError != nil {
		return m.CreateError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if m.GetByIDError != nil {
		return nil, m.GetByIDError
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	user, exists := m.users[id]
	if !exists {
		return nil, nil
	}
	return user, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.GetByEmailError != nil {
		return nil, m.GetByEmailError
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, nil
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	if m.GetByUsernameError != nil {
		return nil, m.GetByUsernameError
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, nil
}

func (m *MockUserRepository) List(ctx context.Context, offset, limit int, status *domain.UserStatus, search *string, sortBy, sortOrder string) ([]*domain.User, int64, error) {
	if m.ListError != nil {
		return nil, 0, m.ListError
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	var users []*domain.User
	for _, user := range m.users {
		// Filter by status if provided
		if status != nil && user.Status != *status {
			continue
		}
		// Filter by search if provided (simple implementation)
		if search != nil && *search != "" {
			// Simple search in email and username
			if user.Email != *search && user.Username != *search {
				continue
			}
		}
		users = append(users, user)
	}

	total := int64(len(users))

	// Apply pagination
	if offset < len(users) {
		end := offset + limit
		if end > len(users) {
			end = len(users)
		}
		users = users[offset:end]
	} else {
		users = []*domain.User{}
	}

	return users, total, nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	if m.UpdateError != nil {
		return m.UpdateError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.users[user.ID]; !exists {
		return nil
	}
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	if m.UpdatePasswordError != nil {
		return m.UpdatePasswordError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if user, exists := m.users[userID]; exists {
		user.PasswordHash = passwordHash
	}
	return nil
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	if m.UpdateLastLoginError != nil {
		return m.UpdateLastLoginError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if user, exists := m.users[userID]; exists {
		now := user.CreatedAt // Use any time
		user.LastLogin = &now
	}
	return nil
}

func (m *MockUserRepository) UpdateStatus(ctx context.Context, userID string, status domain.UserStatus) error {
	if m.UpdateStatusError != nil {
		return m.UpdateStatusError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if user, exists := m.users[userID]; exists {
		user.Status = status
	}
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	if m.DeleteError != nil {
		return m.DeleteError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if user, exists := m.users[id]; exists {
		user.Status = domain.UserStatusDeleted
		now := user.CreatedAt
		user.DeletedAt = &now
	}
	return nil
}

func (m *MockUserRepository) HardDelete(ctx context.Context, id string) error {
	if m.HardDeleteError != nil {
		return m.HardDeleteError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.users, id)
	return nil
}

// Helper methods for tests

func (m *MockUserRepository) AddUser(user *domain.User) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[user.ID] = user
}

func (m *MockUserRepository) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users = make(map[string]*domain.User)
}

func (m *MockUserRepository) UserCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.users)
}
