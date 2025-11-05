package interfaces

import (
	"context"
	"toxictoast/services/foodfolio-service/internal/domain"
)

// CompanyRepository defines the interface for company data access
type CompanyRepository interface {
	// Create creates a new company
	Create(ctx context.Context, company *domain.Company) error

	// GetByID retrieves a company by ID
	GetByID(ctx context.Context, id string) (*domain.Company, error)

	// GetBySlug retrieves a company by slug
	GetBySlug(ctx context.Context, slug string) (*domain.Company, error)

	// List retrieves companies with pagination and filtering
	List(ctx context.Context, offset, limit int, search string, includeDeleted bool) ([]*domain.Company, int64, error)

	// Update updates an existing company
	Update(ctx context.Context, company *domain.Company) error

	// Delete soft deletes a company
	Delete(ctx context.Context, id string) error

	// HardDelete permanently deletes a company
	HardDelete(ctx context.Context, id string) error
}
