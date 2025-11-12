package usecase

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/toxictoast/toxictoastgo/shared/kafka"
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
	companyRepo   interfaces.CompanyRepository
	kafkaProducer *kafka.Producer
}

func NewCompanyUseCase(companyRepo interfaces.CompanyRepository, kafkaProducer *kafka.Producer) CompanyUseCase {
	return &companyUseCase{
		companyRepo:   companyRepo,
		kafkaProducer: kafkaProducer,
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

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		event := kafka.FoodfolioCompanyCreatedEvent{
			CompanyID: company.ID,
			Name:      company.Name,
			Slug:      company.Slug,
			CreatedAt: time.Now(),
		}
		if err := uc.kafkaProducer.PublishFoodfolioCompanyCreated("foodfolio.company.created", event); err != nil {
			log.Printf("Warning: Failed to publish company created event: %v", err)
		}
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

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		event := kafka.FoodfolioCompanyUpdatedEvent{
			CompanyID: company.ID,
			Name:      company.Name,
			Slug:      company.Slug,
			UpdatedAt: time.Now(),
		}
		if err := uc.kafkaProducer.PublishFoodfolioCompanyUpdated("foodfolio.company.updated", event); err != nil {
			log.Printf("Warning: Failed to publish company updated event: %v", err)
		}
	}

	return company, nil
}

func (uc *companyUseCase) DeleteCompany(ctx context.Context, id string) error {
	// Check if exists
	_, err := uc.GetCompanyByID(ctx, id)
	if err != nil {
		return err
	}

	if err := uc.companyRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		event := kafka.FoodfolioCompanyDeletedEvent{
			CompanyID: id,
			DeletedAt: time.Now(),
		}
		if err := uc.kafkaProducer.PublishFoodfolioCompanyDeleted("foodfolio.company.deleted", event); err != nil {
			log.Printf("Warning: Failed to publish company deleted event: %v", err)
		}
	}

	return nil
}
