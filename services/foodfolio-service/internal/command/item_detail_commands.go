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
// Commands - Fokus auf wichtigste Operations
// ============================================================================

type CreateItemDetailCommand struct {
	cqrs.BaseCommand
	ItemVariantID string     `json:"item_variant_id"`
	WarehouseID   string     `json:"warehouse_id"`
	LocationID    string     `json:"location_id"`
	ArticleNumber *string    `json:"article_number"`
	PurchasePrice float64    `json:"purchase_price"`
	PurchaseDate  time.Time  `json:"purchase_date"`
	ExpiryDate    *time.Time `json:"expiry_date"`
	HasDeposit    bool       `json:"has_deposit"`
	IsFrozen      bool       `json:"is_frozen"`
}

type BatchCreateItemDetailsCommand struct {
	cqrs.BaseCommand
	ItemVariantID string     `json:"item_variant_id"`
	WarehouseID   string     `json:"warehouse_id"`
	LocationID    string     `json:"location_id"`
	ArticleNumber *string    `json:"article_number"`
	PurchasePrice float64    `json:"purchase_price"`
	PurchaseDate  time.Time  `json:"purchase_date"`
	ExpiryDate    *time.Time `json:"expiry_date"`
	HasDeposit    bool       `json:"has_deposit"`
	IsFrozen      bool       `json:"is_frozen"`
	Quantity      int        `json:"quantity"`
	CreatedIDs    []string   `json:"created_ids"`
}

func (c *BatchCreateItemDetailsCommand) CommandName() string {
	return "batch_create_item_details"
}

func (c *BatchCreateItemDetailsCommand) Validate() error {
	if c.ItemVariantID == "" || c.WarehouseID == "" || c.LocationID == "" {
		return errors.New("item_variant_id, warehouse_id, and location_id are required")
	}
	if c.Quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}
	if c.Quantity > 1000 {
		return errors.New("quantity cannot exceed 1000 items per batch")
	}
	return nil
}

type MoveItemsCommand struct {
	cqrs.BaseCommand
	ItemIDs       []string `json:"item_ids"`
	NewLocationID string   `json:"new_location_id"`
}

func (c *MoveItemsCommand) CommandName() string {
	return "move_items"
}

func (c *MoveItemsCommand) Validate() error {
	if len(c.ItemIDs) == 0 {
		return errors.New("item_ids are required")
	}
	if c.NewLocationID == "" {
		return errors.New("new_location_id is required")
	}
	return nil
}

func (c *CreateItemDetailCommand) CommandName() string {
	return "create_item_detail"
}

func (c *CreateItemDetailCommand) Validate() error {
	if c.ItemVariantID == "" || c.WarehouseID == "" || c.LocationID == "" {
		return errors.New("item_variant_id, warehouse_id, and location_id are required")
	}
	return nil
}

type OpenItemCommand struct {
	cqrs.BaseCommand
}

func (c *OpenItemCommand) CommandName() string {
	return "open_item"
}

func (c *OpenItemCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("id is required")
	}
	return nil
}

type UpdateItemDetailCommand struct {
	cqrs.BaseCommand
	LocationID    string     `json:"location_id"`
	ArticleNumber *string    `json:"article_number"`
	PurchasePrice float64    `json:"purchase_price"`
	ExpiryDate    *time.Time `json:"expiry_date"`
	HasDeposit    bool       `json:"has_deposit"`
	IsFrozen      bool       `json:"is_frozen"`
}

func (c *UpdateItemDetailCommand) CommandName() string {
	return "update_item_detail"
}

func (c *UpdateItemDetailCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("id is required")
	}
	return nil
}

type DeleteItemDetailCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteItemDetailCommand) CommandName() string {
	return "delete_item_detail"
}

func (c *DeleteItemDetailCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("id is required")
	}
	return nil
}

// ============================================================================
// Command Handlers
// ============================================================================

type CreateItemDetailHandler struct {
	detailRepo    interfaces.ItemDetailRepository
	variantRepo   interfaces.ItemVariantRepository
	warehouseRepo interfaces.WarehouseRepository
	locationRepo  interfaces.LocationRepository
	kafkaProducer *kafka.Producer
}

func NewCreateItemDetailHandler(
	detailRepo interfaces.ItemDetailRepository,
	variantRepo interfaces.ItemVariantRepository,
	warehouseRepo interfaces.WarehouseRepository,
	locationRepo interfaces.LocationRepository,
	kafkaProducer *kafka.Producer,
) *CreateItemDetailHandler {
	return &CreateItemDetailHandler{
		detailRepo:    detailRepo,
		variantRepo:   variantRepo,
		warehouseRepo: warehouseRepo,
		locationRepo:  locationRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *CreateItemDetailHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	createCmd := cmd.(*CreateItemDetailCommand)

	// Validate references exist
	if _, err := h.variantRepo.GetByID(ctx, createCmd.ItemVariantID); err != nil {
		return errors.New("variant not found")
	}
	if _, err := h.warehouseRepo.GetByID(ctx, createCmd.WarehouseID); err != nil {
		return errors.New("warehouse not found")
	}
	if _, err := h.locationRepo.GetByID(ctx, createCmd.LocationID); err != nil {
		return errors.New("location not found")
	}

	detail := &domain.ItemDetail{
		ItemVariantID: createCmd.ItemVariantID,
		WarehouseID:   createCmd.WarehouseID,
		LocationID:    createCmd.LocationID,
		ArticleNumber: createCmd.ArticleNumber,
		PurchasePrice: createCmd.PurchasePrice,
		PurchaseDate:  createCmd.PurchaseDate,
		ExpiryDate:    createCmd.ExpiryDate,
		HasDeposit:    createCmd.HasDeposit,
		IsFrozen:      createCmd.IsFrozen,
		IsOpened:      false,
	}

	if err := h.detailRepo.Create(ctx, detail); err != nil {
		return err
	}

	createCmd.AggregateID = detail.ID

	// Publish Kafka event
	if h.kafkaProducer != nil {
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
		if err := h.kafkaProducer.PublishFoodfolioDetailCreated("foodfolio.detail.created", event); err != nil {
			log.Printf("Warning: Failed to publish detail created event: %v", err)
		}
	}

	return nil
}

type OpenItemHandler struct {
	detailRepo    interfaces.ItemDetailRepository
	kafkaProducer *kafka.Producer
}

func NewOpenItemHandler(detailRepo interfaces.ItemDetailRepository, kafkaProducer *kafka.Producer) *OpenItemHandler {
	return &OpenItemHandler{
		detailRepo:    detailRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *OpenItemHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	openCmd := cmd.(*OpenItemCommand)

	detail, err := h.detailRepo.GetByID(ctx, openCmd.AggregateID)
	if err != nil || detail == nil {
		return errors.New("item detail not found")
	}

	if detail.IsOpened {
		return errors.New("item is already opened")
	}

	detail.IsOpened = true
	openedDate := time.Now()
	detail.OpenedDate = &openedDate

	if err := h.detailRepo.Update(ctx, detail); err != nil {
		return err
	}

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.FoodfolioDetailOpenedEvent{
			DetailID:  detail.ID,
			VariantID: detail.ItemVariantID,
			OpenedAt:  openedDate,
		}
		if err := h.kafkaProducer.PublishFoodfolioDetailOpened("foodfolio.detail.opened", event); err != nil {
			log.Printf("Warning: Failed to publish detail opened event: %v", err)
		}
	}

	return nil
}

type UpdateItemDetailHandler struct {
	detailRepo    interfaces.ItemDetailRepository
	locationRepo  interfaces.LocationRepository
	kafkaProducer *kafka.Producer
}

func NewUpdateItemDetailHandler(
	detailRepo interfaces.ItemDetailRepository,
	locationRepo interfaces.LocationRepository,
	kafkaProducer *kafka.Producer,
) *UpdateItemDetailHandler {
	return &UpdateItemDetailHandler{
		detailRepo:    detailRepo,
		locationRepo:  locationRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *UpdateItemDetailHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateItemDetailCommand)

	detail, err := h.detailRepo.GetByID(ctx, updateCmd.AggregateID)
	if err != nil || detail == nil {
		return errors.New("item detail not found")
	}

	// Validate location exists
	if _, err := h.locationRepo.GetByID(ctx, updateCmd.LocationID); err != nil {
		return errors.New("location not found")
	}

	// Track frozen/thawed state changes
	oldIsFrozen := detail.IsFrozen

	detail.LocationID = updateCmd.LocationID
	detail.ArticleNumber = updateCmd.ArticleNumber
	detail.PurchasePrice = updateCmd.PurchasePrice
	detail.ExpiryDate = updateCmd.ExpiryDate
	detail.HasDeposit = updateCmd.HasDeposit
	detail.IsFrozen = updateCmd.IsFrozen

	if err := h.detailRepo.Update(ctx, detail); err != nil {
		return err
	}

	// Publish frozen/thawed events if state changed
	if h.kafkaProducer != nil {
		if !oldIsFrozen && updateCmd.IsFrozen {
			// Item was frozen
			event := kafka.FoodfolioDetailFrozenEvent{
				DetailID:  detail.ID,
				VariantID: detail.ItemVariantID,
				FrozenAt:  time.Now(),
			}
			if err := h.kafkaProducer.PublishFoodfolioDetailFrozen("foodfolio.detail.frozen", event); err != nil {
				log.Printf("Warning: Failed to publish detail frozen event: %v", err)
			}
		} else if oldIsFrozen && !updateCmd.IsFrozen {
			// Item was thawed
			event := kafka.FoodfolioDetailThawedEvent{
				DetailID:  detail.ID,
				VariantID: detail.ItemVariantID,
				ThawedAt:  time.Now(),
			}
			if err := h.kafkaProducer.PublishFoodfolioDetailThawed("foodfolio.detail.thawed", event); err != nil {
				log.Printf("Warning: Failed to publish detail thawed event: %v", err)
			}
		}
	}

	return nil
}

type DeleteItemDetailHandler struct {
	detailRepo    interfaces.ItemDetailRepository
	kafkaProducer *kafka.Producer
}

func NewDeleteItemDetailHandler(detailRepo interfaces.ItemDetailRepository, kafkaProducer *kafka.Producer) *DeleteItemDetailHandler {
	return &DeleteItemDetailHandler{
		detailRepo:    detailRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *DeleteItemDetailHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteItemDetailCommand)

	detail, err := h.detailRepo.GetByID(ctx, deleteCmd.AggregateID)
	if err != nil {
		return err
	}

	if err := h.detailRepo.Delete(ctx, deleteCmd.AggregateID); err != nil {
		return err
	}

	// Publish consumed event (deletion implies consumption)
	if h.kafkaProducer != nil {
		event := kafka.FoodfolioDetailConsumedEvent{
			DetailID:   deleteCmd.AggregateID,
			VariantID:  detail.ItemVariantID,
			ConsumedAt: time.Now(),
		}
		if err := h.kafkaProducer.PublishFoodfolioDetailConsumed("foodfolio.detail.consumed", event); err != nil {
			log.Printf("Warning: Failed to publish detail consumed event: %v", err)
		}
	}

	return nil
}

// BatchCreateItemDetailsHandler handles batch creation of item details
type BatchCreateItemDetailsHandler struct {
	detailRepo    interfaces.ItemDetailRepository
	variantRepo   interfaces.ItemVariantRepository
	warehouseRepo interfaces.WarehouseRepository
	locationRepo  interfaces.LocationRepository
	kafkaProducer *kafka.Producer
}

func NewBatchCreateItemDetailsHandler(
	detailRepo interfaces.ItemDetailRepository,
	variantRepo interfaces.ItemVariantRepository,
	warehouseRepo interfaces.WarehouseRepository,
	locationRepo interfaces.LocationRepository,
	kafkaProducer *kafka.Producer,
) *BatchCreateItemDetailsHandler {
	return &BatchCreateItemDetailsHandler{
		detailRepo:    detailRepo,
		variantRepo:   variantRepo,
		warehouseRepo: warehouseRepo,
		locationRepo:  locationRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *BatchCreateItemDetailsHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	batchCmd := cmd.(*BatchCreateItemDetailsCommand)

	// Validate variant exists
	if _, err := h.variantRepo.GetByID(ctx, batchCmd.ItemVariantID); err != nil {
		return errors.New("variant not found")
	}

	// Validate warehouse exists
	if _, err := h.warehouseRepo.GetByID(ctx, batchCmd.WarehouseID); err != nil {
		return errors.New("warehouse not found")
	}

	// Validate location exists
	if _, err := h.locationRepo.GetByID(ctx, batchCmd.LocationID); err != nil {
		return errors.New("location not found")
	}

	// Create details
	details := make([]*domain.ItemDetail, batchCmd.Quantity)
	for i := 0; i < batchCmd.Quantity; i++ {
		details[i] = &domain.ItemDetail{
			ItemVariantID: batchCmd.ItemVariantID,
			WarehouseID:   batchCmd.WarehouseID,
			LocationID:    batchCmd.LocationID,
			ArticleNumber: batchCmd.ArticleNumber,
			PurchasePrice: batchCmd.PurchasePrice,
			PurchaseDate:  batchCmd.PurchaseDate,
			ExpiryDate:    batchCmd.ExpiryDate,
			HasDeposit:    batchCmd.HasDeposit,
			IsFrozen:      batchCmd.IsFrozen,
			IsOpened:      false,
		}
	}

	if err := h.detailRepo.BatchCreate(ctx, details); err != nil {
		return err
	}

	// Collect created IDs
	batchCmd.CreatedIDs = make([]string, len(details))
	for i, detail := range details {
		batchCmd.CreatedIDs[i] = detail.ID
	}

	// Publish Kafka events for batch created items
	if h.kafkaProducer != nil {
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
			if err := h.kafkaProducer.PublishFoodfolioDetailCreated("foodfolio.detail.created", event); err != nil {
				log.Printf("Warning: Failed to publish detail created event for detail %s: %v", detail.ID, err)
			}
		}
	}

	return nil
}

// MoveItemsHandler handles moving multiple items to a new location
type MoveItemsHandler struct {
	detailRepo    interfaces.ItemDetailRepository
	locationRepo  interfaces.LocationRepository
	kafkaProducer *kafka.Producer
}

func NewMoveItemsHandler(
	detailRepo interfaces.ItemDetailRepository,
	locationRepo interfaces.LocationRepository,
	kafkaProducer *kafka.Producer,
) *MoveItemsHandler {
	return &MoveItemsHandler{
		detailRepo:    detailRepo,
		locationRepo:  locationRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *MoveItemsHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	moveCmd := cmd.(*MoveItemsCommand)

	// Validate location exists
	if _, err := h.locationRepo.GetByID(ctx, moveCmd.NewLocationID); err != nil {
		return errors.New("location not found")
	}

	// Get old location IDs before moving
	oldLocationIDs := make(map[string]string)
	if h.kafkaProducer != nil {
		for _, id := range moveCmd.ItemIDs {
			detail, err := h.detailRepo.GetByID(ctx, id)
			if err == nil && detail != nil {
				oldLocationIDs[id] = detail.LocationID
			}
		}
	}

	// Move items
	if err := h.detailRepo.MoveItems(ctx, moveCmd.ItemIDs, moveCmd.NewLocationID); err != nil {
		return err
	}

	// Publish Kafka events for moved items
	if h.kafkaProducer != nil {
		for _, id := range moveCmd.ItemIDs {
			if oldLocationID, exists := oldLocationIDs[id]; exists {
				detail, err := h.detailRepo.GetByID(ctx, id)
				if err == nil && detail != nil {
					event := kafka.FoodfolioDetailMovedEvent{
						DetailID:      id,
						VariantID:     detail.ItemVariantID,
						OldLocationID: oldLocationID,
						NewLocationID: moveCmd.NewLocationID,
						MovedAt:       time.Now(),
					}
					if err := h.kafkaProducer.PublishFoodfolioDetailMoved("foodfolio.detail.moved", event); err != nil {
						log.Printf("Warning: Failed to publish detail moved event for detail %s: %v", id, err)
					}
				}
			}
		}
	}

	return nil
}
