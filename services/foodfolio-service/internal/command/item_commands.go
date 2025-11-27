package command

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"github.com/toxictoast/toxictoastgo/shared/kafka"

	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/interfaces"
)

// ============================================================================
// Commands
// ============================================================================

type CreateItemCommand struct {
	cqrs.BaseCommand
	Name       string `json:"name"`
	CategoryID string `json:"category_id"`
	CompanyID  string `json:"company_id"`
	TypeID     string `json:"type_id"`
}

func (c *CreateItemCommand) CommandName() string {
	return "create_item"
}

func (c *CreateItemCommand) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	if c.CategoryID == "" {
		return errors.New("category_id is required")
	}
	if c.CompanyID == "" {
		return errors.New("company_id is required")
	}
	if c.TypeID == "" {
		return errors.New("type_id is required")
	}
	return nil
}

type UpdateItemCommand struct {
	cqrs.BaseCommand
	Name       *string `json:"name"`
	CategoryID *string `json:"category_id"`
	CompanyID  *string `json:"company_id"`
	TypeID     *string `json:"type_id"`
}

func (c *UpdateItemCommand) CommandName() string {
	return "update_item"
}

func (c *UpdateItemCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("id is required")
	}
	return nil
}

type DeleteItemCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteItemCommand) CommandName() string {
	return "delete_item"
}

func (c *DeleteItemCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("id is required")
	}
	return nil
}

// ============================================================================
// Command Handlers
// ============================================================================

type CreateItemHandler struct {
	itemRepo      interfaces.ItemRepository
	categoryRepo  interfaces.CategoryRepository
	companyRepo   interfaces.CompanyRepository
	typeRepo      interfaces.TypeRepository
	kafkaProducer *kafka.Producer
}

func NewCreateItemHandler(
	itemRepo interfaces.ItemRepository,
	categoryRepo interfaces.CategoryRepository,
	companyRepo interfaces.CompanyRepository,
	typeRepo interfaces.TypeRepository,
	kafkaProducer *kafka.Producer,
) *CreateItemHandler {
	return &CreateItemHandler{
		itemRepo:      itemRepo,
		categoryRepo:  categoryRepo,
		companyRepo:   companyRepo,
		typeRepo:      typeRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *CreateItemHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	createCmd := cmd.(*CreateItemCommand)

	// Validate category exists
	category, err := h.categoryRepo.GetByID(ctx, createCmd.CategoryID)
	if err != nil {
		return err
	}
	if category == nil {
		return errors.New("category not found")
	}

	// Validate company exists
	company, err := h.companyRepo.GetByID(ctx, createCmd.CompanyID)
	if err != nil {
		return err
	}
	if company == nil {
		return errors.New("company not found")
	}

	// Validate type exists
	typeEntity, err := h.typeRepo.GetByID(ctx, createCmd.TypeID)
	if err != nil {
		return err
	}
	if typeEntity == nil {
		return errors.New("type not found")
	}

	item := &domain.Item{
		Name:       createCmd.Name,
		CategoryID: createCmd.CategoryID,
		CompanyID:  createCmd.CompanyID,
		TypeID:     createCmd.TypeID,
	}

	if err := h.itemRepo.Create(ctx, item); err != nil {
		return err
	}

	createCmd.AggregateID = item.ID

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.FoodfolioItemCreatedEvent{
			ItemID:     item.ID,
			Name:       item.Name,
			Slug:       item.Slug,
			CategoryID: item.CategoryID,
			CompanyID:  item.CompanyID,
			TypeID:     item.TypeID,
			CreatedAt:  time.Now(),
		}
		if err := h.kafkaProducer.PublishFoodfolioItemCreated("foodfolio.item.created", event); err != nil {
			log.Printf("Warning: Failed to publish item created event: %v", err)
		}
	}

	return nil
}

type UpdateItemHandler struct {
	itemRepo      interfaces.ItemRepository
	categoryRepo  interfaces.CategoryRepository
	companyRepo   interfaces.CompanyRepository
	typeRepo      interfaces.TypeRepository
	kafkaProducer *kafka.Producer
}

func NewUpdateItemHandler(
	itemRepo interfaces.ItemRepository,
	categoryRepo interfaces.CategoryRepository,
	companyRepo interfaces.CompanyRepository,
	typeRepo interfaces.TypeRepository,
	kafkaProducer *kafka.Producer,
) *UpdateItemHandler {
	return &UpdateItemHandler{
		itemRepo:      itemRepo,
		categoryRepo:  categoryRepo,
		companyRepo:   companyRepo,
		typeRepo:      typeRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *UpdateItemHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateItemCommand)

	item, err := h.itemRepo.GetByID(ctx, updateCmd.AggregateID)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.New("item not found")
	}

	if updateCmd.Name != nil {
		item.Name = *updateCmd.Name
	}

	if updateCmd.CategoryID != nil {
		// Validate category exists
		category, err := h.categoryRepo.GetByID(ctx, *updateCmd.CategoryID)
		if err != nil {
			return err
		}
		if category == nil {
			return errors.New("category not found")
		}
		item.CategoryID = *updateCmd.CategoryID
	}

	if updateCmd.CompanyID != nil {
		// Validate company exists
		company, err := h.companyRepo.GetByID(ctx, *updateCmd.CompanyID)
		if err != nil {
			return err
		}
		if company == nil {
			return errors.New("company not found")
		}
		item.CompanyID = *updateCmd.CompanyID
	}

	if updateCmd.TypeID != nil {
		// Validate type exists
		typeEntity, err := h.typeRepo.GetByID(ctx, *updateCmd.TypeID)
		if err != nil {
			return err
		}
		if typeEntity == nil {
			return errors.New("type not found")
		}
		item.TypeID = *updateCmd.TypeID
	}

	if err := h.itemRepo.Update(ctx, item); err != nil {
		return err
	}

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.FoodfolioItemUpdatedEvent{
			ItemID:     item.ID,
			Name:       item.Name,
			Slug:       item.Slug,
			CategoryID: item.CategoryID,
			CompanyID:  item.CompanyID,
			TypeID:     item.TypeID,
			UpdatedAt:  time.Now(),
		}
		if err := h.kafkaProducer.PublishFoodfolioItemUpdated("foodfolio.item.updated", event); err != nil {
			log.Printf("Warning: Failed to publish item updated event: %v", err)
		}
	}

	return nil
}

type DeleteItemHandler struct {
	itemRepo      interfaces.ItemRepository
	kafkaProducer *kafka.Producer
}

func NewDeleteItemHandler(
	itemRepo interfaces.ItemRepository,
	kafkaProducer *kafka.Producer,
) *DeleteItemHandler {
	return &DeleteItemHandler{
		itemRepo:      itemRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *DeleteItemHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteItemCommand)

	_, err := h.itemRepo.GetByID(ctx, deleteCmd.AggregateID)
	if err != nil {
		return err
	}

	if err := h.itemRepo.Delete(ctx, deleteCmd.AggregateID); err != nil {
		return err
	}

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.FoodfolioItemDeletedEvent{
			ItemID:    deleteCmd.AggregateID,
			DeletedAt: time.Now(),
		}
		if err := h.kafkaProducer.PublishFoodfolioItemDeleted("foodfolio.item.deleted", event); err != nil {
			log.Printf("Warning: Failed to publish item deleted event: %v", err)
		}
	}

	return nil
}
