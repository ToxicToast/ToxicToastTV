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

type CreateItemVariantCommand struct {
	cqrs.BaseCommand
	ItemID           string  `json:"item_id"`
	SizeID           string  `json:"size_id"`
	VariantName      string  `json:"variant_name"`
	Barcode          *string `json:"barcode"`
	MinSKU           int     `json:"min_sku"`
	MaxSKU           int     `json:"max_sku"`
	IsNormallyFrozen bool    `json:"is_normally_frozen"`
}

func (c *CreateItemVariantCommand) CommandName() string {
	return "create_item_variant"
}

func (c *CreateItemVariantCommand) Validate() error {
	if c.ItemID == "" || c.SizeID == "" || c.VariantName == "" {
		return errors.New("item_id, size_id, and variant_name are required")
	}
	if c.MinSKU < 0 || c.MaxSKU < 0 {
		return errors.New("SKU values cannot be negative")
	}
	if c.MaxSKU > 0 && c.MinSKU >= c.MaxSKU {
		return errors.New("minSKU must be less than maxSKU")
	}
	return nil
}

type UpdateItemVariantCommand struct {
	cqrs.BaseCommand
	VariantName      *string `json:"variant_name"`
	Barcode          *string `json:"barcode"`
	MinSKU           *int    `json:"min_sku"`
	MaxSKU           *int    `json:"max_sku"`
	IsNormallyFrozen *bool   `json:"is_normally_frozen"`
}

func (c *UpdateItemVariantCommand) CommandName() string {
	return "update_item_variant"
}

func (c *UpdateItemVariantCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("id is required")
	}
	if c.MinSKU != nil && *c.MinSKU < 0 {
		return errors.New("minSKU cannot be negative")
	}
	if c.MaxSKU != nil && *c.MaxSKU < 0 {
		return errors.New("maxSKU cannot be negative")
	}
	if c.MinSKU != nil && c.MaxSKU != nil && *c.MaxSKU > 0 && *c.MinSKU >= *c.MaxSKU {
		return errors.New("minSKU must be less than maxSKU")
	}
	return nil
}

type DeleteItemVariantCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteItemVariantCommand) CommandName() string {
	return "delete_item_variant"
}

func (c *DeleteItemVariantCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("id is required")
	}
	return nil
}

// ============================================================================
// Command Handlers
// ============================================================================

type CreateItemVariantHandler struct {
	variantRepo   interfaces.ItemVariantRepository
	itemRepo      interfaces.ItemRepository
	sizeRepo      interfaces.SizeRepository
	kafkaProducer *kafka.Producer
}

func NewCreateItemVariantHandler(
	variantRepo interfaces.ItemVariantRepository,
	itemRepo interfaces.ItemRepository,
	sizeRepo interfaces.SizeRepository,
	kafkaProducer *kafka.Producer,
) *CreateItemVariantHandler {
	return &CreateItemVariantHandler{
		variantRepo:   variantRepo,
		itemRepo:      itemRepo,
		sizeRepo:      sizeRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *CreateItemVariantHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	createCmd := cmd.(*CreateItemVariantCommand)

	// Validate item exists
	item, err := h.itemRepo.GetByID(ctx, createCmd.ItemID)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.New("item not found")
	}

	// Validate size exists
	size, err := h.sizeRepo.GetByID(ctx, createCmd.SizeID)
	if err != nil {
		return err
	}
	if size == nil {
		return errors.New("size not found")
	}

	// Check barcode uniqueness if provided
	if createCmd.Barcode != nil && *createCmd.Barcode != "" {
		existing, err := h.variantRepo.GetByBarcode(ctx, *createCmd.Barcode)
		if err != nil {
			return err
		}
		if existing != nil {
			return errors.New("barcode already exists")
		}
	}

	variant := &domain.ItemVariant{
		ItemID:           createCmd.ItemID,
		SizeID:           createCmd.SizeID,
		VariantName:      createCmd.VariantName,
		Barcode:          createCmd.Barcode,
		MinSKU:           createCmd.MinSKU,
		MaxSKU:           createCmd.MaxSKU,
		IsNormallyFrozen: createCmd.IsNormallyFrozen,
	}

	if err := h.variantRepo.Create(ctx, variant); err != nil {
		return err
	}

	createCmd.AggregateID = variant.ID

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.FoodfolioVariantCreatedEvent{
			VariantID:        variant.ID,
			ItemID:           variant.ItemID,
			SizeID:           variant.SizeID,
			VariantName:      variant.VariantName,
			Barcode:          variant.Barcode,
			MinSKU:           variant.MinSKU,
			MaxSKU:           variant.MaxSKU,
			IsNormallyFrozen: variant.IsNormallyFrozen,
			CreatedAt:        time.Now(),
		}
		if err := h.kafkaProducer.PublishFoodfolioVariantCreated("foodfolio.variant.created", event); err != nil {
			log.Printf("Warning: Failed to publish variant created event: %v", err)
		}
	}

	return nil
}

type UpdateItemVariantHandler struct {
	variantRepo   interfaces.ItemVariantRepository
	kafkaProducer *kafka.Producer
}

func NewUpdateItemVariantHandler(
	variantRepo interfaces.ItemVariantRepository,
	kafkaProducer *kafka.Producer,
) *UpdateItemVariantHandler {
	return &UpdateItemVariantHandler{
		variantRepo:   variantRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *UpdateItemVariantHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateItemVariantCommand)

	variant, err := h.variantRepo.GetByID(ctx, updateCmd.AggregateID)
	if err != nil {
		return err
	}
	if variant == nil {
		return errors.New("item variant not found")
	}

	if updateCmd.VariantName != nil {
		variant.VariantName = *updateCmd.VariantName
	}

	// Check barcode uniqueness if changed
	if updateCmd.Barcode != nil {
		if *updateCmd.Barcode != "" {
			if variant.Barcode == nil || *variant.Barcode != *updateCmd.Barcode {
				existing, err := h.variantRepo.GetByBarcode(ctx, *updateCmd.Barcode)
				if err != nil {
					return err
				}
				if existing != nil && existing.ID != updateCmd.AggregateID {
					return errors.New("barcode already exists")
				}
			}
		}
		variant.Barcode = updateCmd.Barcode
	}

	if updateCmd.MinSKU != nil {
		variant.MinSKU = *updateCmd.MinSKU
	}

	if updateCmd.MaxSKU != nil {
		variant.MaxSKU = *updateCmd.MaxSKU
	}

	if updateCmd.IsNormallyFrozen != nil {
		variant.IsNormallyFrozen = *updateCmd.IsNormallyFrozen
	}

	if err := h.variantRepo.Update(ctx, variant); err != nil {
		return err
	}

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.FoodfolioVariantUpdatedEvent{
			VariantID:        variant.ID,
			VariantName:      variant.VariantName,
			Barcode:          variant.Barcode,
			MinSKU:           variant.MinSKU,
			MaxSKU:           variant.MaxSKU,
			IsNormallyFrozen: variant.IsNormallyFrozen,
			UpdatedAt:        time.Now(),
		}
		if err := h.kafkaProducer.PublishFoodfolioVariantUpdated("foodfolio.variant.updated", event); err != nil {
			log.Printf("Warning: Failed to publish variant updated event: %v", err)
		}

		// Check stock levels after update
		stock, err := h.variantRepo.GetCurrentStock(ctx, variant.ID)
		if err == nil {
			if stock == 0 {
				// Publish stock empty event
				emptyEvent := kafka.FoodfolioVariantStockEmptyEvent{
					VariantID:   variant.ID,
					ItemID:      variant.ItemID,
					VariantName: variant.VariantName,
					DetectedAt:  time.Now(),
				}
				if err := h.kafkaProducer.PublishFoodfolioVariantStockEmpty("foodfolio.variant.stock.empty", emptyEvent); err != nil {
					log.Printf("Warning: Failed to publish variant stock empty event: %v", err)
				}
			} else if stock < variant.MinSKU {
				// Publish low stock event
				lowStockEvent := kafka.FoodfolioVariantStockLowEvent{
					VariantID:    variant.ID,
					ItemID:       variant.ItemID,
					VariantName:  variant.VariantName,
					CurrentStock: stock,
					MinSKU:       variant.MinSKU,
					DetectedAt:   time.Now(),
				}
				if err := h.kafkaProducer.PublishFoodfolioVariantStockLow("foodfolio.variant.stock.low", lowStockEvent); err != nil {
					log.Printf("Warning: Failed to publish variant stock low event: %v", err)
				}
			}
		}
	}

	return nil
}

type DeleteItemVariantHandler struct {
	variantRepo   interfaces.ItemVariantRepository
	kafkaProducer *kafka.Producer
}

func NewDeleteItemVariantHandler(
	variantRepo interfaces.ItemVariantRepository,
	kafkaProducer *kafka.Producer,
) *DeleteItemVariantHandler {
	return &DeleteItemVariantHandler{
		variantRepo:   variantRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *DeleteItemVariantHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteItemVariantCommand)

	_, err := h.variantRepo.GetByID(ctx, deleteCmd.AggregateID)
	if err != nil {
		return err
	}

	if err := h.variantRepo.Delete(ctx, deleteCmd.AggregateID); err != nil {
		return err
	}

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.FoodfolioVariantDeletedEvent{
			VariantID: deleteCmd.AggregateID,
			DeletedAt: time.Now(),
		}
		if err := h.kafkaProducer.PublishFoodfolioVariantDeleted("foodfolio.variant.deleted", event); err != nil {
			log.Printf("Warning: Failed to publish variant deleted event: %v", err)
		}
	}

	return nil
}
