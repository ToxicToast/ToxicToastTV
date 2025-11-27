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
// Shoppinglist Commands
// ============================================================================

type CreateShoppinglistCommand struct {
	cqrs.BaseCommand
	Name string `json:"name"`
}

func (c *CreateShoppinglistCommand) CommandName() string {
	return "create_shoppinglist"
}

func (c *CreateShoppinglistCommand) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

type UpdateShoppinglistCommand struct {
	cqrs.BaseCommand
	Name string `json:"name"`
}

func (c *UpdateShoppinglistCommand) CommandName() string {
	return "update_shoppinglist"
}

func (c *UpdateShoppinglistCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("id is required")
	}
	if c.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

type DeleteShoppinglistCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteShoppinglistCommand) CommandName() string {
	return "delete_shoppinglist"
}

func (c *DeleteShoppinglistCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("id is required")
	}
	return nil
}

type GenerateFromLowStockCommand struct {
	cqrs.BaseCommand
	Name       string `json:"name"`
	ItemsAdded int    `json:"items_added"`
}

func (c *GenerateFromLowStockCommand) CommandName() string {
	return "generate_from_low_stock"
}

func (c *GenerateFromLowStockCommand) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

// ============================================================================
// Shoppinglist Item Commands
// ============================================================================

type AddItemToShoppinglistCommand struct {
	cqrs.BaseCommand
	ShoppinglistID string `json:"shoppinglist_id"`
	VariantID      string `json:"variant_id"`
	Quantity       int    `json:"quantity"`
	ItemID         string `json:"item_id"`
}

func (c *AddItemToShoppinglistCommand) CommandName() string {
	return "add_item_to_shoppinglist"
}

func (c *AddItemToShoppinglistCommand) Validate() error {
	if c.ShoppinglistID == "" || c.VariantID == "" {
		return errors.New("shoppinglist_id and variant_id are required")
	}
	if c.Quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}
	return nil
}

type RemoveItemFromShoppinglistCommand struct {
	cqrs.BaseCommand
	ShoppinglistID string `json:"shoppinglist_id"`
	ItemID         string `json:"item_id"`
}

func (c *RemoveItemFromShoppinglistCommand) CommandName() string {
	return "remove_item_from_shoppinglist"
}

func (c *RemoveItemFromShoppinglistCommand) Validate() error {
	if c.ShoppinglistID == "" || c.ItemID == "" {
		return errors.New("shoppinglist_id and item_id are required")
	}
	return nil
}

type UpdateShoppinglistItemCommand struct {
	cqrs.BaseCommand
	Quantity    int  `json:"quantity"`
	IsPurchased bool `json:"is_purchased"`
}

func (c *UpdateShoppinglistItemCommand) CommandName() string {
	return "update_shoppinglist_item"
}

func (c *UpdateShoppinglistItemCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("item_id is required")
	}
	if c.Quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}
	return nil
}

type MarkItemPurchasedCommand struct {
	cqrs.BaseCommand
}

func (c *MarkItemPurchasedCommand) CommandName() string {
	return "mark_item_purchased"
}

func (c *MarkItemPurchasedCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("item_id is required")
	}
	return nil
}

type MarkAllItemsPurchasedCommand struct {
	cqrs.BaseCommand
	ShoppinglistID string `json:"shoppinglist_id"`
	ItemsMarked    int    `json:"items_marked"`
}

func (c *MarkAllItemsPurchasedCommand) CommandName() string {
	return "mark_all_items_purchased"
}

func (c *MarkAllItemsPurchasedCommand) Validate() error {
	if c.ShoppinglistID == "" {
		return errors.New("shoppinglist_id is required")
	}
	return nil
}

type ClearPurchasedItemsCommand struct {
	cqrs.BaseCommand
	ShoppinglistID string `json:"shoppinglist_id"`
	ItemsCleared   int    `json:"items_cleared"`
}

func (c *ClearPurchasedItemsCommand) CommandName() string {
	return "clear_purchased_items"
}

func (c *ClearPurchasedItemsCommand) Validate() error {
	if c.ShoppinglistID == "" {
		return errors.New("shoppinglist_id is required")
	}
	return nil
}

// ============================================================================
// Shoppinglist Command Handlers
// ============================================================================

type CreateShoppinglistHandler struct {
	shoppinglistRepo interfaces.ShoppinglistRepository
	kafkaProducer    *kafka.Producer
}

func NewCreateShoppinglistHandler(
	shoppinglistRepo interfaces.ShoppinglistRepository,
	kafkaProducer *kafka.Producer,
) *CreateShoppinglistHandler {
	return &CreateShoppinglistHandler{
		shoppinglistRepo: shoppinglistRepo,
		kafkaProducer:    kafkaProducer,
	}
}

func (h *CreateShoppinglistHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	createCmd := cmd.(*CreateShoppinglistCommand)

	shoppinglist := &domain.Shoppinglist{
		Name: createCmd.Name,
	}

	if err := h.shoppinglistRepo.Create(ctx, shoppinglist); err != nil {
		return err
	}

	createCmd.AggregateID = shoppinglist.ID

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.FoodfolioShoppinglistCreatedEvent{
			ShoppinglistID: shoppinglist.ID,
			Name:           shoppinglist.Name,
			CreatedAt:      time.Now(),
		}
		if err := h.kafkaProducer.PublishFoodfolioShoppinglistCreated("foodfolio.shoppinglist.created", event); err != nil {
			log.Printf("Warning: Failed to publish shoppinglist created event: %v", err)
		}
	}

	return nil
}

type UpdateShoppinglistHandler struct {
	shoppinglistRepo interfaces.ShoppinglistRepository
	kafkaProducer    *kafka.Producer
}

func NewUpdateShoppinglistHandler(
	shoppinglistRepo interfaces.ShoppinglistRepository,
	kafkaProducer *kafka.Producer,
) *UpdateShoppinglistHandler {
	return &UpdateShoppinglistHandler{
		shoppinglistRepo: shoppinglistRepo,
		kafkaProducer:    kafkaProducer,
	}
}

func (h *UpdateShoppinglistHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateShoppinglistCommand)

	shoppinglist, err := h.shoppinglistRepo.GetByID(ctx, updateCmd.AggregateID)
	if err != nil {
		return err
	}
	if shoppinglist == nil {
		return errors.New("shoppinglist not found")
	}

	shoppinglist.Name = updateCmd.Name

	if err := h.shoppinglistRepo.Update(ctx, shoppinglist); err != nil {
		return err
	}

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.FoodfolioShoppinglistUpdatedEvent{
			ShoppinglistID: shoppinglist.ID,
			Name:           shoppinglist.Name,
			UpdatedAt:      time.Now(),
		}
		if err := h.kafkaProducer.PublishFoodfolioShoppinglistUpdated("foodfolio.shoppinglist.updated", event); err != nil {
			log.Printf("Warning: Failed to publish shoppinglist updated event: %v", err)
		}
	}

	return nil
}

type DeleteShoppinglistHandler struct {
	shoppinglistRepo interfaces.ShoppinglistRepository
	kafkaProducer    *kafka.Producer
}

func NewDeleteShoppinglistHandler(
	shoppinglistRepo interfaces.ShoppinglistRepository,
	kafkaProducer *kafka.Producer,
) *DeleteShoppinglistHandler {
	return &DeleteShoppinglistHandler{
		shoppinglistRepo: shoppinglistRepo,
		kafkaProducer:    kafkaProducer,
	}
}

func (h *DeleteShoppinglistHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteShoppinglistCommand)

	_, err := h.shoppinglistRepo.GetByID(ctx, deleteCmd.AggregateID)
	if err != nil {
		return err
	}

	if err := h.shoppinglistRepo.Delete(ctx, deleteCmd.AggregateID); err != nil {
		return err
	}

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.FoodfolioShoppinglistDeletedEvent{
			ShoppinglistID: deleteCmd.AggregateID,
			DeletedAt:      time.Now(),
		}
		if err := h.kafkaProducer.PublishFoodfolioShoppinglistDeleted("foodfolio.shoppinglist.deleted", event); err != nil {
			log.Printf("Warning: Failed to publish shoppinglist deleted event: %v", err)
		}
	}

	return nil
}

type GenerateFromLowStockHandler struct {
	shoppinglistRepo interfaces.ShoppinglistRepository
	variantRepo      interfaces.ItemVariantRepository
}

func NewGenerateFromLowStockHandler(
	shoppinglistRepo interfaces.ShoppinglistRepository,
	variantRepo interfaces.ItemVariantRepository,
) *GenerateFromLowStockHandler {
	return &GenerateFromLowStockHandler{
		shoppinglistRepo: shoppinglistRepo,
		variantRepo:      variantRepo,
	}
}

func (h *GenerateFromLowStockHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	genCmd := cmd.(*GenerateFromLowStockCommand)

	// Create shoppinglist
	shoppinglist := &domain.Shoppinglist{
		Name: genCmd.Name,
	}

	if err := h.shoppinglistRepo.Create(ctx, shoppinglist); err != nil {
		return err
	}

	genCmd.AggregateID = shoppinglist.ID

	// Get low stock variants
	lowStockVariants, _, err := h.variantRepo.GetLowStockVariants(ctx, 0, 1000) // Get all low stock items
	if err != nil {
		return err
	}

	// Add items to shoppinglist
	itemsAdded := 0
	for _, variant := range lowStockVariants {
		// Calculate how many to buy (bring to MinSKU)
		currentStock, err := h.variantRepo.GetCurrentStock(ctx, variant.ID)
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

		if err := h.shoppinglistRepo.AddItem(ctx, item); err != nil {
			continue // Skip if error
		}

		itemsAdded++
	}

	genCmd.ItemsAdded = itemsAdded

	return nil
}

// ============================================================================
// Shoppinglist Item Command Handlers
// ============================================================================

type AddItemToShoppinglistHandler struct {
	shoppinglistRepo interfaces.ShoppinglistRepository
	variantRepo      interfaces.ItemVariantRepository
	kafkaProducer    *kafka.Producer
}

func NewAddItemToShoppinglistHandler(
	shoppinglistRepo interfaces.ShoppinglistRepository,
	variantRepo interfaces.ItemVariantRepository,
	kafkaProducer *kafka.Producer,
) *AddItemToShoppinglistHandler {
	return &AddItemToShoppinglistHandler{
		shoppinglistRepo: shoppinglistRepo,
		variantRepo:      variantRepo,
		kafkaProducer:    kafkaProducer,
	}
}

func (h *AddItemToShoppinglistHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	addCmd := cmd.(*AddItemToShoppinglistCommand)

	// Validate shoppinglist exists
	_, err := h.shoppinglistRepo.GetByID(ctx, addCmd.ShoppinglistID)
	if err != nil {
		return err
	}

	// Validate variant exists
	variant, err := h.variantRepo.GetByID(ctx, addCmd.VariantID)
	if err != nil {
		return err
	}
	if variant == nil {
		return errors.New("item variant not found")
	}

	item := &domain.ShoppinglistItem{
		ShoppinglistID: addCmd.ShoppinglistID,
		ItemVariantID:  addCmd.VariantID,
		Quantity:       addCmd.Quantity,
		IsPurchased:    false,
	}

	if err := h.shoppinglistRepo.AddItem(ctx, item); err != nil {
		return err
	}

	addCmd.ItemID = item.ID

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.FoodfolioShoppinglistItemAddedEvent{
			ShoppinglistID: addCmd.ShoppinglistID,
			ItemID:         item.ID,
			VariantID:      addCmd.VariantID,
			Quantity:       addCmd.Quantity,
			AddedAt:        time.Now(),
		}
		if err := h.kafkaProducer.PublishFoodfolioShoppinglistItemAdded("foodfolio.shoppinglist.item.added", event); err != nil {
			log.Printf("Warning: Failed to publish shoppinglist item added event: %v", err)
		}
	}

	return nil
}

type RemoveItemFromShoppinglistHandler struct {
	shoppinglistRepo interfaces.ShoppinglistRepository
	kafkaProducer    *kafka.Producer
}

func NewRemoveItemFromShoppinglistHandler(
	shoppinglistRepo interfaces.ShoppinglistRepository,
	kafkaProducer *kafka.Producer,
) *RemoveItemFromShoppinglistHandler {
	return &RemoveItemFromShoppinglistHandler{
		shoppinglistRepo: shoppinglistRepo,
		kafkaProducer:    kafkaProducer,
	}
}

func (h *RemoveItemFromShoppinglistHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	removeCmd := cmd.(*RemoveItemFromShoppinglistCommand)

	// Validate shoppinglist exists
	_, err := h.shoppinglistRepo.GetByID(ctx, removeCmd.ShoppinglistID)
	if err != nil {
		return err
	}

	if err := h.shoppinglistRepo.RemoveItem(ctx, removeCmd.ShoppinglistID, removeCmd.ItemID); err != nil {
		return err
	}

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.FoodfolioShoppinglistItemRemovedEvent{
			ShoppinglistID: removeCmd.ShoppinglistID,
			ItemID:         removeCmd.ItemID,
			RemovedAt:      time.Now(),
		}
		if err := h.kafkaProducer.PublishFoodfolioShoppinglistItemRemoved("foodfolio.shoppinglist.item.removed", event); err != nil {
			log.Printf("Warning: Failed to publish shoppinglist item removed event: %v", err)
		}
	}

	return nil
}

type UpdateShoppinglistItemHandler struct {
	shoppinglistRepo interfaces.ShoppinglistRepository
}

func NewUpdateShoppinglistItemHandler(
	shoppinglistRepo interfaces.ShoppinglistRepository,
) *UpdateShoppinglistItemHandler {
	return &UpdateShoppinglistItemHandler{
		shoppinglistRepo: shoppinglistRepo,
	}
}

func (h *UpdateShoppinglistItemHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateShoppinglistItemCommand)

	// Get item
	item, err := h.shoppinglistRepo.GetItem(ctx, updateCmd.AggregateID)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.New("shoppinglist item not found")
	}

	item.Quantity = updateCmd.Quantity
	item.IsPurchased = updateCmd.IsPurchased

	return h.shoppinglistRepo.UpdateItem(ctx, item)
}

type MarkItemPurchasedHandler struct {
	shoppinglistRepo interfaces.ShoppinglistRepository
	kafkaProducer    *kafka.Producer
}

func NewMarkItemPurchasedHandler(
	shoppinglistRepo interfaces.ShoppinglistRepository,
	kafkaProducer *kafka.Producer,
) *MarkItemPurchasedHandler {
	return &MarkItemPurchasedHandler{
		shoppinglistRepo: shoppinglistRepo,
		kafkaProducer:    kafkaProducer,
	}
}

func (h *MarkItemPurchasedHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	markCmd := cmd.(*MarkItemPurchasedCommand)

	// Get item before marking purchased to get details for event
	item, err := h.shoppinglistRepo.GetItem(ctx, markCmd.AggregateID)
	if err != nil {
		return err
	}

	if err := h.shoppinglistRepo.MarkItemPurchased(ctx, markCmd.AggregateID); err != nil {
		return err
	}

	// Publish Kafka event
	if h.kafkaProducer != nil && item != nil {
		event := kafka.FoodfolioShoppinglistItemPurchasedEvent{
			ShoppinglistID: item.ShoppinglistID,
			ItemID:         markCmd.AggregateID,
			VariantID:      item.ItemVariantID,
			Quantity:       item.Quantity,
			PurchasedAt:    time.Now(),
		}
		if err := h.kafkaProducer.PublishFoodfolioShoppinglistItemPurchased("foodfolio.shoppinglist.item.purchased", event); err != nil {
			log.Printf("Warning: Failed to publish shoppinglist item purchased event: %v", err)
		}
	}

	return nil
}

type MarkAllItemsPurchasedHandler struct {
	shoppinglistRepo interfaces.ShoppinglistRepository
}

func NewMarkAllItemsPurchasedHandler(
	shoppinglistRepo interfaces.ShoppinglistRepository,
) *MarkAllItemsPurchasedHandler {
	return &MarkAllItemsPurchasedHandler{
		shoppinglistRepo: shoppinglistRepo,
	}
}

func (h *MarkAllItemsPurchasedHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	markCmd := cmd.(*MarkAllItemsPurchasedCommand)

	// Validate shoppinglist exists
	_, err := h.shoppinglistRepo.GetByID(ctx, markCmd.ShoppinglistID)
	if err != nil {
		return err
	}

	itemsMarked, err := h.shoppinglistRepo.MarkAllItemsPurchased(ctx, markCmd.ShoppinglistID)
	if err != nil {
		return err
	}

	markCmd.ItemsMarked = itemsMarked

	return nil
}

type ClearPurchasedItemsHandler struct {
	shoppinglistRepo interfaces.ShoppinglistRepository
}

func NewClearPurchasedItemsHandler(
	shoppinglistRepo interfaces.ShoppinglistRepository,
) *ClearPurchasedItemsHandler {
	return &ClearPurchasedItemsHandler{
		shoppinglistRepo: shoppinglistRepo,
	}
}

func (h *ClearPurchasedItemsHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	clearCmd := cmd.(*ClearPurchasedItemsCommand)

	// Validate shoppinglist exists
	_, err := h.shoppinglistRepo.GetByID(ctx, clearCmd.ShoppinglistID)
	if err != nil {
		return err
	}

	itemsCleared, err := h.shoppinglistRepo.ClearPurchasedItems(ctx, clearCmd.ShoppinglistID)
	if err != nil {
		return err
	}

	clearCmd.ItemsCleared = itemsCleared

	return nil
}
