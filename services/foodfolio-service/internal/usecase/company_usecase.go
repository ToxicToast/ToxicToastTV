package usecase

import (
	"context"
	"errors"

	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/interfaces"
)

var (
	ErrCompanyNotFound      = errors.New("company not found")
	ErrCompanyAlreadyExists = errors.New("company already exists")
	ErrInvalidCompanyData   = errors.New("invalid company data")
)

type CompanyUseCase interface {
	CreateCompany(ctx context.Context, name string) (*domain.Company, error)
	GetCompanyByID(ctx context.Context, id string) (*domain.Company, error)
	GetCompanyBySlug(ctx context.Context, slug string) (*domain.Company, error)
	ListCompanies(ctx context.Context, page, pageSize int, search string, includeDeleted bool) ([]*domain.Company, int64, error)
	UpdateCompany(ctx context.Context, id, name string) (*domain.Company, error)
	DeleteCompany(ctx context.Context, id string) error
}

type companyUseCase struct {
	companyRepo interfaces.CompanyRepository
}

func NewCompanyUseCase(companyRepo interfaces.CompanyRepository) CompanyUseCase {
	return &companyUseCase{
		companyRepo: companyRepo,
	}
}

func (uc *companyUseCase) CreateCompany(ctx context.Context, name string) (*domain.Company, error) {
	// Validate
	if name == "" {
		return nil, ErrInvalidCompanyData
	}

	// Create company
	company := &domain.Company{
		Name: name,
	}

	if err := uc.companyRepo.Create(ctx, company); err != nil {
		return nil, err
	}

	return company, nil
}

func (uc *companyUseCase) GetCompanyByID(ctx context.Context, id string) (*domain.Company, error) {
	company, err := uc.companyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if company == nil {
		return nil, ErrCompanyNotFound
	}

	return company, nil
}

func (uc *companyUseCase) GetCompanyBySlug(ctx context.Context, slug string) (*domain.Company, error) {
	company, err := uc.companyRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	if company == nil {
		return nil, ErrCompanyNotFound
	}

	return company, nil
}

func (uc *companyUseCase) ListCompanies(ctx context.Context, page, pageSize int, search string, includeDeleted bool) ([]*domain.Company, int64, error) {
	// Calculate offset
	offset := (page - 1) * pageSize

	// Set default page size
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return uc.companyRepo.List(ctx, offset, pageSize, search, includeDeleted)
}

func (uc *companyUseCase) UpdateCompany(ctx context.Context, id, name string) (*domain.Company, error) {
	// Get existing company
	company, err := uc.GetCompanyByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate
	if name == "" {
		return nil, ErrInvalidCompanyData
	}

	// Update
	company.Name = name

	if err := uc.companyRepo.Update(ctx, company); err != nil {
		return nil, err
	}

	return company, nil
}

func (uc *companyUseCase) DeleteCompany(ctx context.Context, id string) error {
	// Check if exists
	_, err := uc.GetCompanyByID(ctx, id)
	if err != nil {
		return err
	}

	return uc.companyRepo.Delete(ctx, id)
}
