package mapper

import (
	"gorm.io/gorm"
	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/entity"
)

// CompanyToEntity converts domain model to database entity
func CompanyToEntity(company *domain.Company) *entity.CompanyEntity {
	if company == nil {
		return nil
	}

	e := &entity.CompanyEntity{
		ID:        company.ID,
		Name:      company.Name,
		Slug:      company.Slug,
		CreatedAt: company.CreatedAt,
		UpdatedAt: company.UpdatedAt,
	}

	// Convert DeletedAt
	if company.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *company.DeletedAt,
			Valid: true,
		}
	}

	return e
}

// CompanyToDomain converts database entity to domain model
func CompanyToDomain(e *entity.CompanyEntity) *domain.Company {
	if e == nil {
		return nil
	}

	company := &domain.Company{
		ID:        e.ID,
		Name:      e.Name,
		Slug:      e.Slug,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		company.DeletedAt = &deletedAt
	}

	return company
}

// CompaniesToDomain converts slice of entities to domain models
func CompaniesToDomain(entities []*entity.CompanyEntity) []*domain.Company {
	companies := make([]*domain.Company, 0, len(entities))
	for _, e := range entities {
		companies = append(companies, CompanyToDomain(e))
	}
	return companies
}
