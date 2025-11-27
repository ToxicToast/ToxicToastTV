package query

import (
	"context"
	"errors"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"

	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/interfaces"
)

// ============================================================================
// Queries
// ============================================================================

type GetCompanyByIDQuery struct {
	cqrs.BaseQuery
	ID string `json:"id"`
}

func (q *GetCompanyByIDQuery) QueryName() string {
	return "get_company_by_id"
}

func (q *GetCompanyByIDQuery) Validate() error {
	if q.ID == "" {
		return errors.New("id is required")
	}
	return nil
}

type GetCompanyBySlugQuery struct {
	cqrs.BaseQuery
	Slug string `json:"slug"`
}

func (q *GetCompanyBySlugQuery) QueryName() string {
	return "get_company_by_slug"
}

func (q *GetCompanyBySlugQuery) Validate() error {
	if q.Slug == "" {
		return errors.New("slug is required")
	}
	return nil
}

type ListCompaniesQuery struct {
	cqrs.BaseQuery
	Page           int    `json:"page"`
	PageSize       int    `json:"page_size"`
	Search         string `json:"search"`
	IncludeDeleted bool   `json:"include_deleted"`
}

func (q *ListCompaniesQuery) QueryName() string {
	return "list_companies"
}

func (q *ListCompaniesQuery) Validate() error {
	return nil
}

// ============================================================================
// Query Results
// ============================================================================

type ListCompaniesResult struct {
	Companies []*domain.Company
	Total     int64
}

// ============================================================================
// Query Handlers
// ============================================================================

type GetCompanyByIDHandler struct {
	companyRepo interfaces.CompanyRepository
}

func NewGetCompanyByIDHandler(companyRepo interfaces.CompanyRepository) *GetCompanyByIDHandler {
	return &GetCompanyByIDHandler{
		companyRepo: companyRepo,
	}
}

func (h *GetCompanyByIDHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetCompanyByIDQuery)

	companyEntity, err := h.companyRepo.GetByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	if companyEntity == nil {
		return nil, errors.New("company not found")
	}

	return companyEntity, nil
}

type GetCompanyBySlugHandler struct {
	companyRepo interfaces.CompanyRepository
}

func NewGetCompanyBySlugHandler(companyRepo interfaces.CompanyRepository) *GetCompanyBySlugHandler {
	return &GetCompanyBySlugHandler{
		companyRepo: companyRepo,
	}
}

func (h *GetCompanyBySlugHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetCompanyBySlugQuery)

	companyEntity, err := h.companyRepo.GetBySlug(ctx, q.Slug)
	if err != nil {
		return nil, err
	}

	if companyEntity == nil {
		return nil, errors.New("company not found")
	}

	return companyEntity, nil
}

type ListCompaniesHandler struct {
	companyRepo interfaces.CompanyRepository
}

func NewListCompaniesHandler(companyRepo interfaces.CompanyRepository) *ListCompaniesHandler {
	return &ListCompaniesHandler{
		companyRepo: companyRepo,
	}
}

func (h *ListCompaniesHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*ListCompaniesQuery)

	page := q.Page
	pageSize := q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	companies, total, err := h.companyRepo.List(ctx, offset, pageSize, q.Search, q.IncludeDeleted)
	if err != nil {
		return nil, err
	}

	return &ListCompaniesResult{
		Companies: companies,
		Total:     total,
	}, nil
}
