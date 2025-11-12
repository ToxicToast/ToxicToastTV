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
	ErrItemDetailNotFound    = errors.New("item detail not found")
	ErrInvalidItemDetailData = errors.New("invalid item detail data")
	ErrItemAlreadyOpened     = errors.New("item already opened")
)

type ItemDetailUseCase interface {
	CreateItemDetail(ctx context.Context, variantID, warehouseID, locationID string, articleNumber *string, purchasePrice float64, purchaseDate time.Time, expiryDate *time.Time, hasDeposit, isFrozen bool) (*domain.ItemDetail, error)
	BatchCreateItemDetails(ctx context.Context, variantID, warehouseID, locationID string, articleNumber *string, purchasePrice float64, purchaseDate time.Time, expiryDate *time.Time, hasDeposit, isFrozen bool, quantity int) ([]*domain.ItemDetail, error)
	GetItemDetailByID(ctx context.Context, id string) (*domain.ItemDetail, error)
	ListItemDetails(ctx context.Context, page, pageSize int, variantID, warehouseID, locationID *string, isOpened, hasDeposit, isFrozen *bool, includeDeleted bool) ([]*domain.ItemDetail, int64, error)
	GetByVariant(ctx context.Context, variantID string) ([]*domain.ItemDetail, error)
	GetByLocation(ctx context.Context, locationID string, includeChildren bool) ([]*domain.ItemDetail, error)
	GetExpiringItems(ctx context.Context, days int, page, pageSize int) ([]*domain.ItemDetail, int64, error)
	GetExpiredItems(ctx context.Context, page, pageSize int) ([]*domain.ItemDetail, int64, error)
	GetItemsWithDeposit(ctx context.Context, page, pageSize int) ([]*domain.ItemDetail, int64, error)
	OpenItem(ctx context.Context, id string) (*domain.ItemDetail, error)
	MoveItems(ctx context.Context, itemIDs []string, newLocationID string) ([]*domain.ItemDetail, error)
	UpdateItemDetail(ctx context.Context, id, locationID string, articleNumber *string, purchasePrice float64, expiryDate *time.Time, hasDeposit, isFrozen bool) (*domain.ItemDetail, error)
	DeleteItemDetail(ctx context.Context, id string) error
}

type itemDetailUseCase struct {
	detailRepo    interfaces.ItemDetailRepository
	variantRepo   interfaces.ItemVariantRepository
	warehouseRepo interfaces.WarehouseRepository
	locationRepo  interfaces.LocationRepository
	kafkaProducer *kafka.Producer
}

func NewItemDetailUseCase(
	detailRepo interfaces.ItemDetailRepository,
	variantRepo interfaces.ItemVariantRepository,
	warehouseRepo interfaces.WarehouseRepository,
	locationRepo interfaces.LocationRepository,
	kafkaProducer *kafka.Producer,
) ItemDetailUseCase {
	return &itemDetailUseCase{
		detailRepo:    detailRepo,
		variantRepo:   variantRepo,
		warehouseRepo: warehouseRepo,
		locationRepo:  locationRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (uc *itemDetailUseCase) CreateItemDetail(ctx context.Context, variantID, warehouseID, locationID string, articleNumber *string, purchasePrice float64, purchaseDate time.Time, expiryDate *time.Time, hasDeposit, isFrozen bool) (*domain.ItemDetail, error) {
	// Validate
	if variantID == "" || warehouseID == "" || locationID == "" {
		return nil, ErrInvalidItemDetailData
	}

	if purchasePrice < 0 {
		return nil, errors.New("purchase price cannot be negative")
	}

	// Validate variant exists
	variant, err := uc.variantRepo.GetByID(ctx, variantID)
	if err != nil {
		return nil, err
	}
	if variant == nil {
		return nil, errors.New("item variant not found")
	}

	// Validate warehouse exists
	warehouse, err := uc.warehouseRepo.GetByID(ctx, warehouseID)
	if err != nil {
		return nil, err
	}
	if warehouse == nil {
		return nil, errors.New("warehouse not found")
	}

	// Validate location exists
	location, err := uc.locationRepo.GetByID(ctx, locationID)
	if err != nil {
		return nil, err
	}
	if location == nil {
		return nil, errors.New("location not found")
	}

	detail := &domain.ItemDetail{
		ItemVariantID: variantID,
		WarehouseID:   warehouseID,
		LocationID:    locationID,
		ArticleNumber: articleNumber,
		PurchasePrice: purchasePrice,
		PurchaseDate:  purchaseDate,
		ExpiryDate:    expiryDate,
		HasDeposit:    hasDeposit,
		IsFrozen:      isFrozen,
		IsOpened:      false,
	}

	if err := uc.detailRepo.Create(ctx, detail); err != nil {
		return nil, err
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		event := kafka.FoodfolioDetailCreatedEvent{
			DetailID:      detail.ID,
			VariantID:     detail.ItemVariantID,
			WarehouseID:   detail.WarehouseID,
			LocationID:    detail.LocationID,
			ArticleNumber: detail.ArticleNumber,
			PurchasePrice: detail.PurchasePrice,
			PurchaseDate:  detail.PurchaseDate,
			ExpiryDate:    detail.ExpiryDate,
			HasDeposit:    detail.HasDeposit,
			IsFrozen:      detail.IsFrozen,
			CreatedAt:     time.Now(),
		}
		if err := uc.kafkaProducer.PublishFoodfolioDetailCreated("foodfolio.detail.created", event); err != nil {
			log.Printf("Warning: Failed to publish detail created event: %v", err)
		}
	}

	return detail, nil
}

func (uc *itemDetailUseCase) BatchCreateItemDetails(ctx context.Context, variantID, warehouseID, locationID string, articleNumber *string, purchasePrice float64, purchaseDate time.Time, expiryDate *time.Time, hasDeposit, isFrozen bool, quantity int) ([]*domain.ItemDetail, error) {
	// Validate
	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than 0")
	}

	if quantity > 1000 {
		return nil, errors.New("quantity cannot exceed 1000 items per batch")
	}

	// Validate references (same as CreateItemDetail)
	if variantID == "" || warehouseID == "" || locationID == "" {
		return nil, ErrInvalidItemDetailData
	}

	if purchasePrice < 0 {
		return nil, errors.New("purchase price cannot be negative")
	}

	// Validate variant exists
	variant, err := uc.variantRepo.GetByID(ctx, variantID)
	if err != nil {
		return nil, err
	}
	if variant == nil {
		return nil, errors.New("item variant not found")
	}

	// Validate warehouse exists
	warehouse, err := uc.warehouseRepo.GetByID(ctx, warehouseID)
	if err != nil {
		return nil, err
	}
	if warehouse == nil {
		return nil, errors.New("warehouse not found")
	}

	// Validate location exists
	location, err := uc.locationRepo.GetByID(ctx, locationID)
	if err != nil {
		return nil, err
	}
	if location == nil {
		return nil, errors.New("location not found")
	}

	// Create details
	details := make([]*domain.ItemDetail, quantity)
	for i := 0; i < quantity; i++ {
		details[i] = &domain.ItemDetail{
			ItemVariantID: variantID,
			WarehouseID:   warehouseID,
			LocationID:    locationID,
			ArticleNumber: articleNumber,
			PurchasePrice: purchasePrice,
			PurchaseDate:  purchaseDate,
			ExpiryDate:    expiryDate,
			HasDeposit:    hasDeposit,
			IsFrozen:      isFrozen,
			IsOpened:      false,
		}
	}

	if err := uc.detailRepo.BatchCreate(ctx, details); err != nil {
		return nil, err
	}

	// Publish Kafka events for batch created items
	if uc.kafkaProducer != nil {
		for _, detail := range details {
			event := kafka.FoodfolioDetailCreatedEvent{
				DetailID:      detail.ID,
				VariantID:     detail.ItemVariantID,
				WarehouseID:   detail.WarehouseID,
				LocationID:    detail.LocationID,
				ArticleNumber: detail.ArticleNumber,
				PurchasePrice: detail.PurchasePrice,
				PurchaseDate:  detail.PurchaseDate,
				ExpiryDate:    detail.ExpiryDate,
				HasDeposit:    detail.HasDeposit,
				IsFrozen:      detail.IsFrozen,
				CreatedAt:     time.Now(),
			}
			if err := uc.kafkaProducer.PublishFoodfolioDetailCreated("foodfolio.detail.created", event); err != nil {
				log.Printf("Warning: Failed to publish detail created event for detail %s: %v", detail.ID, err)
			}
		}
	}

	return details, nil
}

func (uc *itemDetailUseCase) GetItemDetailByID(ctx context.Context, id string) (*domain.ItemDetail, error) {
	detail, err := uc.detailRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if detail == nil {
		return nil, ErrItemDetailNotFound
	}

	return detail, nil
}

func (uc *itemDetailUseCase) ListItemDetails(ctx context.Context, page, pageSize int, variantID, warehouseID, locationID *string, isOpened, hasDeposit, isFrozen *bool, includeDeleted bool) ([]*domain.ItemDetail, int64, error) {
	offset := (page - 1) * pageSize

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return uc.detailRepo.List(ctx, offset, pageSize, variantID, warehouseID, locationID, isOpened, hasDeposit, isFrozen, includeDeleted)
}

func (uc *itemDetailUseCase) GetByVariant(ctx context.Context, variantID string) ([]*domain.ItemDetail, error) {
	// Validate variant exists
	variant, err := uc.variantRepo.GetByID(ctx, variantID)
	if err != nil {
		return nil, err
	}
	if variant == nil {
		return nil, errors.New("item variant not found")
	}

	return uc.detailRepo.GetByVariant(ctx, variantID)
}

func (uc *itemDetailUseCase) GetByLocation(ctx context.Context, locationID string, includeChildren bool) ([]*domain.ItemDetail, error) {
	// Validate location exists
	location, err := uc.locationRepo.GetByID(ctx, locationID)
	if err != nil {
		return nil, err
	}
	if location == nil {
		return nil, errors.New("location not found")
	}

	return uc.detailRepo.GetByLocation(ctx, locationID, includeChildren)
}

func (uc *itemDetailUseCase) GetExpiringItems(ctx context.Context, days int, page, pageSize int) ([]*domain.ItemDetail, int64, error) {
	if days <= 0 {
		return nil, 0, errors.New("days must be greater than 0")
	}

	offset := (page - 1) * pageSize

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return uc.detailRepo.GetExpiringItems(ctx, days, offset, pageSize)
}

func (uc *itemDetailUseCase) GetExpiredItems(ctx context.Context, page, pageSize int) ([]*domain.ItemDetail, int64, error) {
	offset := (page - 1) * pageSize

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return uc.detailRepo.GetExpiredItems(ctx, offset, pageSize)
}

func (uc *itemDetailUseCase) GetItemsWithDeposit(ctx context.Context, page, pageSize int) ([]*domain.ItemDetail, int64, error) {
	offset := (page - 1) * pageSize

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return uc.detailRepo.GetItemsWithDeposit(ctx, offset, pageSize)
}

func (uc *itemDetailUseCase) OpenItem(ctx context.Context, id string) (*domain.ItemDetail, error) {
	detail, err := uc.GetItemDetailByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if detail.IsOpened {
		return nil, ErrItemAlreadyOpened
	}

	openedDate := time.Now()

	if err := uc.detailRepo.OpenItem(ctx, id, openedDate); err != nil {
		return nil, err
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		event := kafka.FoodfolioDetailOpenedEvent{
			DetailID:  id,
			VariantID: detail.ItemVariantID,
			OpenedAt:  openedDate,
		}
		if err := uc.kafkaProducer.PublishFoodfolioDetailOpened("foodfolio.detail.opened", event); err != nil {
			log.Printf("Warning: Failed to publish detail opened event: %v", err)
		}
	}

	// Reload to get updated data
	return uc.GetItemDetailByID(ctx, id)
}

func (uc *itemDetailUseCase) MoveItems(ctx context.Context, itemIDs []string, newLocationID string) ([]*domain.ItemDetail, error) {
	if len(itemIDs) == 0 {
		return nil, errors.New("no items to move")
	}

	// Validate location exists
	location, err := uc.locationRepo.GetByID(ctx, newLocationID)
	if err != nil {
		return nil, err
	}
	if location == nil {
		return nil, errors.New("location not found")
	}

	// Get old location IDs before moving
	oldLocationIDs := make(map[string]string)
	if uc.kafkaProducer != nil {
		for _, id := range itemIDs {
			detail, err := uc.detailRepo.GetByID(ctx, id)
			if err == nil && detail != nil {
				oldLocationIDs[id] = detail.LocationID
			}
		}
	}

	// Move items
	if err := uc.detailRepo.MoveItems(ctx, itemIDs, newLocationID); err != nil {
		return nil, err
	}

	// Publish Kafka events for moved items
	if uc.kafkaProducer != nil {
		for _, id := range itemIDs {
			if oldLocationID, exists := oldLocationIDs[id]; exists {
				detail, err := uc.detailRepo.GetByID(ctx, id)
				if err == nil && detail != nil {
					event := kafka.FoodfolioDetailMovedEvent{
						DetailID:      id,
						VariantID:     detail.ItemVariantID,
						OldLocationID: oldLocationID,
						NewLocationID: newLocationID,
						MovedAt:       time.Now(),
					}
					if err := uc.kafkaProducer.PublishFoodfolioDetailMoved("foodfolio.detail.moved", event); err != nil {
						log.Printf("Warning: Failed to publish detail moved event for detail %s: %v", id, err)
					}
				}
			}
		}
	}

	// Reload items to get updated data
	var details []*domain.ItemDetail
	for _, id := range itemIDs {
		detail, err := uc.GetItemDetailByID(ctx, id)
		if err != nil {
			continue // Skip if not found
		}
		details = append(details, detail)
	}

	return details, nil
}

func (uc *itemDetailUseCase) UpdateItemDetail(ctx context.Context, id, locationID string, articleNumber *string, purchasePrice float64, expiryDate *time.Time, hasDeposit, isFrozen bool) (*domain.ItemDetail, error) {
	detail, err := uc.GetItemDetailByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate
	if locationID == "" {
		return nil, ErrInvalidItemDetailData
	}

	if purchasePrice < 0 {
		return nil, errors.New("purchase price cannot be negative")
	}

	// Validate location exists
	location, err := uc.locationRepo.GetByID(ctx, locationID)
	if err != nil {
		return nil, err
	}
	if location == nil {
		return nil, errors.New("location not found")
	}

	// Track frozen/thawed state changes
	oldIsFrozen := detail.IsFrozen

	detail.LocationID = locationID
	detail.ArticleNumber = articleNumber
	detail.PurchasePrice = purchasePrice
	detail.ExpiryDate = expiryDate
	detail.HasDeposit = hasDeposit
	detail.IsFrozen = isFrozen

	if err := uc.detailRepo.Update(ctx, detail); err != nil {
		return nil, err
	}

	// Publish frozen/thawed events if state changed
	if uc.kafkaProducer != nil {
		if !oldIsFrozen && isFrozen {
			// Item was frozen
			event := kafka.FoodfolioDetailFrozenEvent{
				DetailID:  detail.ID,
				VariantID: detail.ItemVariantID,
				FrozenAt:  time.Now(),
			}
			if err := uc.kafkaProducer.PublishFoodfolioDetailFrozen("foodfolio.detail.frozen", event); err != nil {
				log.Printf("Warning: Failed to publish detail frozen event: %v", err)
			}
		} else if oldIsFrozen && !isFrozen {
			// Item was thawed
			event := kafka.FoodfolioDetailThawedEvent{
				DetailID:  detail.ID,
				VariantID: detail.ItemVariantID,
				ThawedAt:  time.Now(),
			}
			if err := uc.kafkaProducer.PublishFoodfolioDetailThawed("foodfolio.detail.thawed", event); err != nil {
				log.Printf("Warning: Failed to publish detail thawed event: %v", err)
			}
		}
	}

	return detail, nil
}

func (uc *itemDetailUseCase) DeleteItemDetail(ctx context.Context, id string) error {
	detail, err := uc.GetItemDetailByID(ctx, id)
	if err != nil {
		return err
	}

	if err := uc.detailRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Publish consumed event (deletion implies consumption)
	if uc.kafkaProducer != nil {
		event := kafka.FoodfolioDetailConsumedEvent{
			DetailID:   id,
			VariantID:  detail.ItemVariantID,
			ConsumedAt: time.Now(),
		}
		if err := uc.kafkaProducer.PublishFoodfolioDetailConsumed("foodfolio.detail.consumed", event); err != nil {
			log.Printf("Warning: Failed to publish detail consumed event: %v", err)
		}
	}

	return nil
}
