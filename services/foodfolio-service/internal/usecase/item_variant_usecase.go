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
	ErrItemVariantNotFound    = errors.New("item variant not found")
	ErrInvalidItemVariantData = errors.New("invalid item variant data")
	ErrBarcodeAlreadyExists   = errors.New("barcode already exists")
)

type ItemVariantUseCase interface {
	CreateItemVariant(ctx context.Context, itemID, sizeID, variantName string, barcode *string, minSKU, maxSKU int, isNormallyFrozen bool) (*domain.ItemVariant, error)
	GetItemVariantByID(ctx context.Context, id string) (*domain.ItemVariant, error)
	GetItemVariantByBarcode(ctx context.Context, barcode string) (*domain.ItemVariant, error)
	ListItemVariants(ctx context.Context, page, pageSize int, itemID, sizeID *string, isNormallyFrozen *bool, includeDeleted bool) ([]*domain.ItemVariant, int64, error)
	GetByItem(ctx context.Context, itemID string) ([]*domain.ItemVariant, error)
	GetLowStockVariants(ctx context.Context, page, pageSize int) ([]*domain.ItemVariant, int64, error)
	GetOverstockedVariants(ctx context.Context, page, pageSize int) ([]*domain.ItemVariant, int64, error)
	GetCurrentStock(ctx context.Context, variantID string) (int, bool, bool, error)
	UpdateItemVariant(ctx context.Context, id, variantName string, barcode *string, minSKU, maxSKU int, isNormallyFrozen bool) (*domain.ItemVariant, error)
	DeleteItemVariant(ctx context.Context, id string) error
}

type itemVariantUseCase struct {
	variantRepo   interfaces.ItemVariantRepository
	itemRepo      interfaces.ItemRepository
	sizeRepo      interfaces.SizeRepository
	kafkaProducer *kafka.Producer
}

func NewItemVariantUseCase(
	variantRepo interfaces.ItemVariantRepository,
	itemRepo interfaces.ItemRepository,
	sizeRepo interfaces.SizeRepository,
	kafkaProducer *kafka.Producer,
) ItemVariantUseCase {
	return &itemVariantUseCase{
		variantRepo:   variantRepo,
		itemRepo:      itemRepo,
		sizeRepo:      sizeRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (uc *itemVariantUseCase) CreateItemVariant(ctx context.Context, itemID, sizeID, variantName string, barcode *string, minSKU, maxSKU int, isNormallyFrozen bool) (*domain.ItemVariant, error) {
	// Validate
	if itemID == "" || sizeID == "" || variantName == "" {
		return nil, ErrInvalidItemVariantData
	}

	if minSKU < 0 || maxSKU < 0 {
		return nil, errors.New("SKU values cannot be negative")
	}

	if maxSKU > 0 && minSKU >= maxSKU {
		return nil, errors.New("minSKU must be less than maxSKU")
	}

	// Validate item exists
	item, err := uc.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("item not found")
	}

	// Validate size exists
	size, err := uc.sizeRepo.GetByID(ctx, sizeID)
	if err != nil {
		return nil, err
	}
	if size == nil {
		return nil, errors.New("size not found")
	}

	// Check barcode uniqueness if provided
	if barcode != nil && *barcode != "" {
		existing, err := uc.variantRepo.GetByBarcode(ctx, *barcode)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, ErrBarcodeAlreadyExists
		}
	}

	variant := &domain.ItemVariant{
		ItemID:           itemID,
		SizeID:           sizeID,
		VariantName:      variantName,
		Barcode:          barcode,
		MinSKU:           minSKU,
		MaxSKU:           maxSKU,
		IsNormallyFrozen: isNormallyFrozen,
	}

	if err := uc.variantRepo.Create(ctx, variant); err != nil {
		return nil, err
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
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
		if err := uc.kafkaProducer.PublishFoodfolioVariantCreated("foodfolio.variant.created", event); err != nil {
			log.Printf("Warning: Failed to publish variant created event: %v", err)
		}
	}

	return variant, nil
}

func (uc *itemVariantUseCase) GetItemVariantByID(ctx context.Context, id string) (*domain.ItemVariant, error) {
	variant, err := uc.variantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if variant == nil {
		return nil, ErrItemVariantNotFound
	}

	return variant, nil
}

func (uc *itemVariantUseCase) GetItemVariantByBarcode(ctx context.Context, barcode string) (*domain.ItemVariant, error) {
	if barcode == "" {
		return nil, errors.New("barcode cannot be empty")
	}

	variant, err := uc.variantRepo.GetByBarcode(ctx, barcode)
	if err != nil {
		return nil, err
	}

	if variant == nil {
		return nil, ErrItemVariantNotFound
	}

	return variant, nil
}

func (uc *itemVariantUseCase) ListItemVariants(ctx context.Context, page, pageSize int, itemID, sizeID *string, isNormallyFrozen *bool, includeDeleted bool) ([]*domain.ItemVariant, int64, error) {
	offset := (page - 1) * pageSize

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return uc.variantRepo.List(ctx, offset, pageSize, itemID, sizeID, isNormallyFrozen, includeDeleted)
}

func (uc *itemVariantUseCase) GetByItem(ctx context.Context, itemID string) ([]*domain.ItemVariant, error) {
	// Validate item exists
	item, err := uc.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("item not found")
	}

	return uc.variantRepo.GetByItem(ctx, itemID)
}

func (uc *itemVariantUseCase) GetLowStockVariants(ctx context.Context, page, pageSize int) ([]*domain.ItemVariant, int64, error) {
	offset := (page - 1) * pageSize

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return uc.variantRepo.GetLowStockVariants(ctx, offset, pageSize)
}

func (uc *itemVariantUseCase) GetOverstockedVariants(ctx context.Context, page, pageSize int) ([]*domain.ItemVariant, int64, error) {
	offset := (page - 1) * pageSize

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return uc.variantRepo.GetOverstockedVariants(ctx, offset, pageSize)
}

func (uc *itemVariantUseCase) GetCurrentStock(ctx context.Context, variantID string) (int, bool, bool, error) {
	// Get variant to check thresholds
	variant, err := uc.GetItemVariantByID(ctx, variantID)
	if err != nil {
		return 0, false, false, err
	}

	// Get current stock
	stock, err := uc.variantRepo.GetCurrentStock(ctx, variantID)
	if err != nil {
		return 0, false, false, err
	}

	// Check if needs restock
	needsRestock := stock < variant.MinSKU

	// Check if overstocked
	isOverstocked := variant.MaxSKU > 0 && stock > variant.MaxSKU

	return stock, needsRestock, isOverstocked, nil
}

func (uc *itemVariantUseCase) UpdateItemVariant(ctx context.Context, id, variantName string, barcode *string, minSKU, maxSKU int, isNormallyFrozen bool) (*domain.ItemVariant, error) {
	variant, err := uc.GetItemVariantByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate
	if variantName == "" {
		return nil, ErrInvalidItemVariantData
	}

	if minSKU < 0 || maxSKU < 0 {
		return nil, errors.New("SKU values cannot be negative")
	}

	if maxSKU > 0 && minSKU >= maxSKU {
		return nil, errors.New("minSKU must be less than maxSKU")
	}

	// Check barcode uniqueness if changed
	if barcode != nil && *barcode != "" {
		if variant.Barcode == nil || *variant.Barcode != *barcode {
			existing, err := uc.variantRepo.GetByBarcode(ctx, *barcode)
			if err != nil {
				return nil, err
			}
			if existing != nil && existing.ID != id {
				return nil, ErrBarcodeAlreadyExists
			}
		}
	}

	variant.VariantName = variantName
	variant.Barcode = barcode
	variant.MinSKU = minSKU
	variant.MaxSKU = maxSKU
	variant.IsNormallyFrozen = isNormallyFrozen

	if err := uc.variantRepo.Update(ctx, variant); err != nil {
		return nil, err
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		event := kafka.FoodfolioVariantUpdatedEvent{
			VariantID:        variant.ID,
			VariantName:      variant.VariantName,
			Barcode:          variant.Barcode,
			MinSKU:           variant.MinSKU,
			MaxSKU:           variant.MaxSKU,
			IsNormallyFrozen: variant.IsNormallyFrozen,
			UpdatedAt:        time.Now(),
		}
		if err := uc.kafkaProducer.PublishFoodfolioVariantUpdated("foodfolio.variant.updated", event); err != nil {
			log.Printf("Warning: Failed to publish variant updated event: %v", err)
		}
	}

	// Check stock levels after update
	if uc.kafkaProducer != nil {
		stock, needsRestock, _, err := uc.GetCurrentStock(ctx, variant.ID)
		if err == nil {
			if stock == 0 {
				// Publish stock empty event
				emptyEvent := kafka.FoodfolioVariantStockEmptyEvent{
					VariantID:   variant.ID,
					ItemID:      variant.ItemID,
					VariantName: variant.VariantName,
					DetectedAt:  time.Now(),
				}
				if err := uc.kafkaProducer.PublishFoodfolioVariantStockEmpty("foodfolio.variant.stock.empty", emptyEvent); err != nil {
					log.Printf("Warning: Failed to publish variant stock empty event: %v", err)
				}
			} else if needsRestock {
				// Publish low stock event
				lowStockEvent := kafka.FoodfolioVariantStockLowEvent{
					VariantID:    variant.ID,
					ItemID:       variant.ItemID,
					VariantName:  variant.VariantName,
					CurrentStock: stock,
					MinSKU:       variant.MinSKU,
					DetectedAt:   time.Now(),
				}
				if err := uc.kafkaProducer.PublishFoodfolioVariantStockLow("foodfolio.variant.stock.low", lowStockEvent); err != nil {
					log.Printf("Warning: Failed to publish variant stock low event: %v", err)
				}
			}
		}
	}

	return variant, nil
}

func (uc *itemVariantUseCase) DeleteItemVariant(ctx context.Context, id string) error {
	_, err := uc.GetItemVariantByID(ctx, id)
	if err != nil {
		return err
	}

	if err := uc.variantRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		event := kafka.FoodfolioVariantDeletedEvent{
			VariantID: id,
			DeletedAt: time.Now(),
		}
		if err := uc.kafkaProducer.PublishFoodfolioVariantDeleted("foodfolio.variant.deleted", event); err != nil {
			log.Printf("Warning: Failed to publish variant deleted event: %v", err)
		}
	}

	return nil
}
