package command

import (
	"context"
	"errors"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"

	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/interfaces"
)

// ============================================================================
// Commands
// ============================================================================

type CreateCompanyCommand struct {
	cqrs.BaseCommand
	Name string `json:"name"`
}

func (c *CreateCompanyCommand) CommandName() string {
	return "create_company"
}

func (c *CreateCompanyCommand) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

type UpdateCompanyCommand struct {
	cqrs.BaseCommand
	Name string `json:"name"`
}

func (c *UpdateCompanyCommand) CommandName() string {
	return "update_company"
}

func (c *UpdateCompanyCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("id is required")
	}
	if c.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

type DeleteCompanyCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteCompanyCommand) CommandName() string {
	return "delete_company"
}

func (c *DeleteCompanyCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("id is required")
	}
	return nil
}

// ============================================================================
// Command Handlers
// ============================================================================

type CreateCompanyHandler struct {
	companyRepo interfaces.CompanyRepository
}

func NewCreateCompanyHandler(companyRepo interfaces.CompanyRepository) *CreateCompanyHandler {
	return &CreateCompanyHandler{
		companyRepo: companyRepo,
	}
}

func (h *CreateCompanyHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	createCmd := cmd.(*CreateCompanyCommand)

	companyEntity := &domain.Company{
		Name: createCmd.Name,
	}

	if err := h.companyRepo.Create(ctx, companyEntity); err != nil {
		return err
	}

	createCmd.AggregateID = companyEntity.ID
	return nil
}

type UpdateCompanyHandler struct {
	companyRepo interfaces.CompanyRepository
}

func NewUpdateCompanyHandler(companyRepo interfaces.CompanyRepository) *UpdateCompanyHandler {
	return &UpdateCompanyHandler{
		companyRepo: companyRepo,
	}
}

func (h *UpdateCompanyHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateCompanyCommand)

	companyEntity, err := h.companyRepo.GetByID(ctx, updateCmd.AggregateID)
	if err != nil {
		return err
	}

	if companyEntity == nil {
		return errors.New("company not found")
	}

	companyEntity.Name = updateCmd.Name

	if err := h.companyRepo.Update(ctx, companyEntity); err != nil {
		return err
	}

	return nil
}

type DeleteCompanyHandler struct {
	companyRepo interfaces.CompanyRepository
}

func NewDeleteCompanyHandler(companyRepo interfaces.CompanyRepository) *DeleteCompanyHandler {
	return &DeleteCompanyHandler{
		companyRepo: companyRepo,
	}
}

func (h *DeleteCompanyHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteCompanyCommand)

	_, err := h.companyRepo.GetByID(ctx, deleteCmd.AggregateID)
	if err != nil {
		return err
	}

	return h.companyRepo.Delete(ctx, deleteCmd.AggregateID)
}
