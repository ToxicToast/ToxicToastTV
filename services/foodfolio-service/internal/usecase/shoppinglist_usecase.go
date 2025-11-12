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
	ErrShoppinglistNotFound    = errors.New("shoppinglist not found")
	ErrShoppinglistItemNotFound = errors.New("shoppinglist item not found")
	ErrInvalidShoppinglistData  = errors.New("invalid shoppinglist data")
)

type ShoppinglistUseCase interface {
	CreateShoppinglist(ctx context.Context, name string) (*domain.Shoppinglist, error)
	GetShoppinglistByID(ctx context.Context, id string) (*domain.Shoppinglist, error)
	ListShoppinglists(ctx context.Context, page, pageSize int, includeDeleted bool) ([]*domain.Shoppinglist, int64, error)
	UpdateShoppinglist(ctx context.Context, id, name string) (*domain.Shoppinglist, error)
	DeleteShoppinglist(ctx context.Context, id string) error
	AddItemToShoppinglist(ctx context.Context, shoppinglistID, variantID string, quantity int) (*domain.ShoppinglistItem, error)
	RemoveItemFromShoppinglist(ctx context.Context, shoppinglistID, itemID string) error
	UpdateShoppinglistItem(ctx context.Context, itemID string, quantity int, isPurchased bool) (*domain.ShoppinglistItem, error)
	MarkItemPurchased(ctx context.Context, itemID string) (*domain.ShoppinglistItem, error)
	MarkAllItemsPurchased(ctx context.Context, shoppinglistID string) (int, error)
	ClearPurchasedItems(ctx context.Context, shoppinglistID string) (int, error)
	GenerateFromLowStock(ctx context.Context, name string) (*domain.Shoppinglist, int, error)
}

type shoppinglistUseCase struct {
	shoppinglistRepo interfaces.ShoppinglistRepository
	variantRepo      interfaces.ItemVariantRepository
	kafkaProducer    *kafka.Producer
}

func NewShoppinglistUseCase(
	shoppinglistRepo interfaces.ShoppinglistRepository,
	variantRepo interfaces.ItemVariantRepository,
	kafkaProducer *kafka.Producer,
) ShoppinglistUseCase {
	return &shoppinglistUseCase{
		shoppinglistRepo: shoppinglistRepo,
		variantRepo:      variantRepo,
		kafkaProducer:    kafkaProducer,
	}
}

func (uc *shoppinglistUseCase) CreateShoppinglist(ctx context.Context, name string) (*domain.Shoppinglist, error) {
	if name == "" {
		return nil, ErrInvalidShoppinglistData
	}

	shoppinglist := &domain.Shoppinglist{
		Name: name,
	}

	if err := uc.shoppinglistRepo.Create(ctx, shoppinglist); err != nil {
		return nil, err
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		event := kafka.FoodfolioShoppinglistCreatedEvent{
			ShoppinglistID: shoppinglist.ID,
			Name:           shoppinglist.Name,
			CreatedAt:      time.Now(),
		}
		if err := uc.kafkaProducer.PublishFoodfolioShoppinglistCreated("foodfolio.shoppinglist.created", event); err != nil {
			log.Printf("Warning: Failed to publish shoppinglist created event: %v", err)
		}
	}

	return shoppinglist, nil
}

func (uc *shoppinglistUseCase) GetShoppinglistByID(ctx context.Context, id string) (*domain.Shoppinglist, error) {
	shoppinglist, err := uc.shoppinglistRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if shoppinglist == nil {
		return nil, ErrShoppinglistNotFound
	}

	return shoppinglist, nil
}

func (uc *shoppinglistUseCase) ListShoppinglists(ctx context.Context, page, pageSize int, includeDeleted bool) ([]*domain.Shoppinglist, int64, error) {
	offset := (page - 1) * pageSize

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return uc.shoppinglistRepo.List(ctx, offset, pageSize, includeDeleted)
}

func (uc *shoppinglistUseCase) UpdateShoppinglist(ctx context.Context, id, name string) (*domain.Shoppinglist, error) {
	shoppinglist, err := uc.GetShoppinglistByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name == "" {
		return nil, ErrInvalidShoppinglistData
	}

	shoppinglist.Name = name

	if err := uc.shoppinglistRepo.Update(ctx, shoppinglist); err != nil {
		return nil, err
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		event := kafka.FoodfolioShoppinglistUpdatedEvent{
			ShoppinglistID: shoppinglist.ID,
			Name:           shoppinglist.Name,
			UpdatedAt:      time.Now(),
		}
		if err := uc.kafkaProducer.PublishFoodfolioShoppinglistUpdated("foodfolio.shoppinglist.updated", event); err != nil {
			log.Printf("Warning: Failed to publish shoppinglist updated event: %v", err)
		}
	}

	return shoppinglist, nil
}

func (uc *shoppinglistUseCase) DeleteShoppinglist(ctx context.Context, id string) error {
	_, err := uc.GetShoppinglistByID(ctx, id)
	if err != nil {
		return err
	}

	if err := uc.shoppinglistRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		event := kafka.FoodfolioShoppinglistDeletedEvent{
			ShoppinglistID: id,
			DeletedAt:      time.Now(),
		}
		if err := uc.kafkaProducer.PublishFoodfolioShoppinglistDeleted("foodfolio.shoppinglist.deleted", event); err != nil {
			log.Printf("Warning: Failed to publish shoppinglist deleted event: %v", err)
		}
	}

	return nil
}

func (uc *shoppinglistUseCase) AddItemToShoppinglist(ctx context.Context, shoppinglistID, variantID string, quantity int) (*domain.ShoppinglistItem, error) {
	// Validate
	if shoppinglistID == "" || variantID == "" {
		return nil, ErrInvalidShoppinglistData
	}

	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than 0")
	}

	// Validate shoppinglist exists
	_, err := uc.GetShoppinglistByID(ctx, shoppinglistID)
	if err != nil {
		return nil, err
	}

	// Validate variant exists
	variant, err := uc.variantRepo.GetByID(ctx, variantID)
	if err != nil {
		return nil, err
	}
	if variant == nil {
		return nil, errors.New("item variant not found")
	}

	item := &domain.ShoppinglistItem{
		ShoppinglistID: shoppinglistID,
		ItemVariantID:  variantID,
		Quantity:       quantity,
		IsPurchased:    false,
	}

	if err := uc.shoppinglistRepo.AddItem(ctx, item); err != nil {
		return nil, err
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		event := kafka.FoodfolioShoppinglistItemAddedEvent{
			ShoppinglistID: shoppinglistID,
			ItemID:         item.ID,
			VariantID:      variantID,
			Quantity:       quantity,
			AddedAt:        time.Now(),
		}
		if err := uc.kafkaProducer.PublishFoodfolioShoppinglistItemAdded("foodfolio.shoppinglist.item.added", event); err != nil {
			log.Printf("Warning: Failed to publish shoppinglist item added event: %v", err)
		}
	}

	return item, nil
}

func (uc *shoppinglistUseCase) RemoveItemFromShoppinglist(ctx context.Context, shoppinglistID, itemID string) error {
	// Validate shoppinglist exists
	_, err := uc.GetShoppinglistByID(ctx, shoppinglistID)
	if err != nil {
		return err
	}

	if err := uc.shoppinglistRepo.RemoveItem(ctx, shoppinglistID, itemID); err != nil {
		return err
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		event := kafka.FoodfolioShoppinglistItemRemovedEvent{
			ShoppinglistID: shoppinglistID,
			ItemID:         itemID,
			RemovedAt:      time.Now(),
		}
		if err := uc.kafkaProducer.PublishFoodfolioShoppinglistItemRemoved("foodfolio.shoppinglist.item.removed", event); err != nil {
			log.Printf("Warning: Failed to publish shoppinglist item removed event: %v", err)
		}
	}

	return nil
}

func (uc *shoppinglistUseCase) UpdateShoppinglistItem(ctx context.Context, itemID string, quantity int, isPurchased bool) (*domain.ShoppinglistItem, error) {
	// Get item
	item, err := uc.shoppinglistRepo.GetItem(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrShoppinglistItemNotFound
	}

	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than 0")
	}

	item.Quantity = quantity
	item.IsPurchased = isPurchased

	if err := uc.shoppinglistRepo.UpdateItem(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (uc *shoppinglistUseCase) MarkItemPurchased(ctx context.Context, itemID string) (*domain.ShoppinglistItem, error) {
	// Get item before marking purchased to get details for event
	item, err := uc.shoppinglistRepo.GetItem(ctx, itemID)
	if err != nil {
		return nil, err
	}

	if err := uc.shoppinglistRepo.MarkItemPurchased(ctx, itemID); err != nil {
		return nil, err
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil && item != nil {
		event := kafka.FoodfolioShoppinglistItemPurchasedEvent{
			ShoppinglistID: item.ShoppinglistID,
			ItemID:         itemID,
			VariantID:      item.ItemVariantID,
			Quantity:       item.Quantity,
			PurchasedAt:    time.Now(),
		}
		if err := uc.kafkaProducer.PublishFoodfolioShoppinglistItemPurchased("foodfolio.shoppinglist.item.purchased", event); err != nil {
			log.Printf("Warning: Failed to publish shoppinglist item purchased event: %v", err)
		}
	}

	// Reload item
	return uc.shoppinglistRepo.GetItem(ctx, itemID)
}

func (uc *shoppinglistUseCase) MarkAllItemsPurchased(ctx context.Context, shoppinglistID string) (int, error) {
	// Validate shoppinglist exists
	_, err := uc.GetShoppinglistByID(ctx, shoppinglistID)
	if err != nil {
		return 0, err
	}

	return uc.shoppinglistRepo.MarkAllItemsPurchased(ctx, shoppinglistID)
}

func (uc *shoppinglistUseCase) ClearPurchasedItems(ctx context.Context, shoppinglistID string) (int, error) {
	// Validate shoppinglist exists
	_, err := uc.GetShoppinglistByID(ctx, shoppinglistID)
	if err != nil {
		return 0, err
	}

	return uc.shoppinglistRepo.ClearPurchasedItems(ctx, shoppinglistID)
}

func (uc *shoppinglistUseCase) GenerateFromLowStock(ctx context.Context, name string) (*domain.Shoppinglist, int, error) {
	if name == "" {
		return nil, 0, ErrInvalidShoppinglistData
	}

	// Create shoppinglist
	shoppinglist := &domain.Shoppinglist{
		Name: name,
	}

	if err := uc.shoppinglistRepo.Create(ctx, shoppinglist); err != nil {
		return nil, 0, err
	}

	// Get low stock variants
	lowStockVariants, _, err := uc.variantRepo.GetLowStockVariants(ctx, 0, 1000) // Get all low stock items
	if err != nil {
		return nil, 0, err
	}

	// Add items to shoppinglist
	itemsAdded := 0
	for _, variant := range lowStockVariants {
		// Calculate how many to buy (bring to MinSKU)
		currentStock, err := uc.variantRepo.GetCurrentStock(ctx, variant.ID)
		if err != nil {
			continue
		}

		quantityNeeded := variant.MinSKU - currentStock
		if quantityNeeded <= 0 {
			continue
		}

		item := &domain.ShoppinglistItem{
			ShoppinglistID: shoppinglist.ID,
			ItemVariantID:  variant.ID,
			Quantity:       quantityNeeded,
			IsPurchased:    false,
		}

		if err := uc.shoppinglistRepo.AddItem(ctx, item); err != nil {
			continue // Skip if error
		}

		itemsAdded++
	}

	return shoppinglist, itemsAdded, nil
}
