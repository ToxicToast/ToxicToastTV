package mocks

import (
	"context"
	"sync"

	"toxictoast/services/auth-service/internal/domain"
)

// MockRoleRepository is a mock implementation of RoleRepository for testing
type MockRoleRepository struct {
	mu    sync.RWMutex
	roles map[string]*domain.Role

	// Error injection for testing
	CreateError     error
	GetByIDError    error
	GetByNameError  error
	ListError       error
	UpdateError     error
	DeleteError     error
	HardDeleteError error
}

// NewMockRoleRepository creates a new mock repository
func NewMockRoleRepository() *MockRoleRepository {
	return &MockRoleRepository{
		roles: make(map[string]*domain.Role),
	}
}

func (m *MockRoleRepository) Create(ctx context.Context, role *domain.Role) error {
	if m.CreateError != nil {
		return m.CreateError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.roles[role.ID] = role
	return nil
}

func (m *MockRoleRepository) GetByID(ctx context.Context, id string) (*domain.Role, error) {
	if m.GetByIDError != nil {
		return nil, m.GetByIDError
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	role, exists := m.roles[id]
	if !exists {
		return nil, nil
	}
	return role, nil
}

func (m *MockRoleRepository) GetByName(ctx context.Context, name string) (*domain.Role, error) {
	if m.GetByNameError != nil {
		return nil, m.GetByNameError
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, role := range m.roles {
		if role.Name == name {
			return role, nil
		}
	}
	return nil, nil
}

func (m *MockRoleRepository) List(ctx context.Context, offset, limit int) ([]*domain.Role, int64, error) {
	if m.ListError != nil {
		return nil, 0, m.ListError
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	var roles []*domain.Role
	for _, role := range m.roles {
		roles = append(roles, role)
	}

	total := int64(len(roles))

	// Apply pagination
	if offset < len(roles) {
		end := offset + limit
		if end > len(roles) {
			end = len(roles)
		}
		roles = roles[offset:end]
	} else {
		roles = []*domain.Role{}
	}

	return roles, total, nil
}

func (m *MockRoleRepository) Update(ctx context.Context, role *domain.Role) error {
	if m.UpdateError != nil {
		return m.UpdateError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.roles[role.ID]; !exists {
		return nil
	}
	m.roles[role.ID] = role
	return nil
}

func (m *MockRoleRepository) Delete(ctx context.Context, id string) error {
	if m.DeleteError != nil {
		return m.DeleteError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if role, exists := m.roles[id]; exists {
		now := role.CreatedAt
		role.DeletedAt = &now
	}
	return nil
}

func (m *MockRoleRepository) HardDelete(ctx context.Context, id string) error {
	if m.HardDeleteError != nil {
		return m.HardDeleteError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.roles, id)
	return nil
}

// Helper methods for tests

func (m *MockRoleRepository) AddRole(role *domain.Role) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.roles[role.ID] = role
}

func (m *MockRoleRepository) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.roles = make(map[string]*domain.Role)
}

func (m *MockRoleRepository) RoleCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.roles)
}
