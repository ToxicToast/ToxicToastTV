package command

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"github.com/toxictoast/toxictoastgo/shared/kafka"

	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/interfaces"
)

// ============================================================================
// Receipt Commands
// ============================================================================

type CreateReceiptCommand struct {
	cqrs.BaseCommand
	WarehouseID string     `json:"warehouse_id"`
	ScanDate    time.Time  `json:"scan_date"`
	TotalPrice  float64    `json:"total_price"`
	ImagePath   *string    `json:"image_path"`
	OCRText     *string    `json:"ocr_text"`
}

func (c *CreateReceiptCommand) CommandName() string {
	return "create_receipt"
}

func (c *CreateReceiptCommand) Validate() error {
	if c.WarehouseID == "" {
		return errors.New("warehouse_id is required")
	}
	if c.TotalPrice < 0 {
		return errors.New("total_price cannot be negative")
	}
	return nil
}

type UpdateReceiptCommand struct {
	cqrs.BaseCommand
	WarehouseID string    `json:"warehouse_id"`
	ScanDate    time.Time `json:"scan_date"`
	TotalPrice  float64   `json:"total_price"`
}

func (c *UpdateReceiptCommand) CommandName() string {
	return "update_receipt"
}

func (c *UpdateReceiptCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("id is required")
	}
	if c.WarehouseID == "" {
		return errors.New("warehouse_id is required")
	}
	if c.TotalPrice < 0 {
		return errors.New("total_price cannot be negative")
	}
	return nil
}

type UpdateReceiptOCRDataCommand struct {
	cqrs.BaseCommand
	ImagePath  string  `json:"image_path"`
	OCRText    string  `json:"ocr_text"`
	TotalPrice float64 `json:"total_price"`
}

func (c *UpdateReceiptOCRDataCommand) CommandName() string {
	return "update_receipt_ocr_data"
}

func (c *UpdateReceiptOCRDataCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("id is required")
	}
	return nil
}

type DeleteReceiptCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteReceiptCommand) CommandName() string {
	return "delete_receipt"
}

func (c *DeleteReceiptCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("id is required")
	}
	return nil
}

// ============================================================================
// Receipt Item Commands
// ============================================================================

type AddItemToReceiptCommand struct {
	cqrs.BaseCommand
	ReceiptID     string  `json:"receipt_id"`
	ItemName      string  `json:"item_name"`
	Quantity      int     `json:"quantity"`
	UnitPrice     float64 `json:"unit_price"`
	TotalPrice    float64 `json:"total_price"`
	ArticleNumber *string `json:"article_number"`
	ItemVariantID *string `json:"item_variant_id"`
	ItemID        string  `json:"item_id"`
}

func (c *AddItemToReceiptCommand) CommandName() string {
	return "add_item_to_receipt"
}

func (c *AddItemToReceiptCommand) Validate() error {
	if c.ReceiptID == "" || c.ItemName == "" {
		return errors.New("receipt_id and item_name are required")
	}
	if c.Quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}
	if c.UnitPrice < 0 || c.TotalPrice < 0 {
		return errors.New("prices cannot be negative")
	}
	return nil
}

type UpdateReceiptItemCommand struct {
	cqrs.BaseCommand
	ItemName      string  `json:"item_name"`
	Quantity      int     `json:"quantity"`
	UnitPrice     float64 `json:"unit_price"`
	TotalPrice    float64 `json:"total_price"`
	ArticleNumber *string `json:"article_number"`
}

func (c *UpdateReceiptItemCommand) CommandName() string {
	return "update_receipt_item"
}

func (c *UpdateReceiptItemCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("item_id is required")
	}
	if c.ItemName == "" {
		return errors.New("item_name is required")
	}
	if c.Quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}
	if c.UnitPrice < 0 || c.TotalPrice < 0 {
		return errors.New("prices cannot be negative")
	}
	return nil
}

type MatchReceiptItemCommand struct {
	cqrs.BaseCommand
	ItemVariantID string `json:"item_variant_id"`
}

func (c *MatchReceiptItemCommand) CommandName() string {
	return "match_receipt_item"
}

func (c *MatchReceiptItemCommand) Validate() error {
	if c.AggregateID == "" || c.ItemVariantID == "" {
		return errors.New("item_id and item_variant_id are required")
	}
	return nil
}

type AutoMatchReceiptItemsCommand struct {
	cqrs.BaseCommand
	ReceiptID           string  `json:"receipt_id"`
	SimilarityThreshold float64 `json:"similarity_threshold"`
	MatchedCount        int     `json:"matched_count"`
	UnmatchedCount      int     `json:"unmatched_count"`
}

func (c *AutoMatchReceiptItemsCommand) CommandName() string {
	return "auto_match_receipt_items"
}

func (c *AutoMatchReceiptItemsCommand) Validate() error {
	if c.ReceiptID == "" {
		return errors.New("receipt_id is required")
	}
	if c.SimilarityThreshold <= 0 || c.SimilarityThreshold > 1 {
		return errors.New("similarity_threshold must be between 0 and 1")
	}
	return nil
}

type CreateInventoryFromReceiptCommand struct {
	cqrs.BaseCommand
	ReceiptID         string     `json:"receipt_id"`
	LocationID        string     `json:"location_id"`
	DefaultExpiryDate *time.Time `json:"default_expiry_date"`
	OnlyMatched       bool       `json:"only_matched"`
	CreatedCount      int        `json:"created_count"`
}

func (c *CreateInventoryFromReceiptCommand) CommandName() string {
	return "create_inventory_from_receipt"
}

func (c *CreateInventoryFromReceiptCommand) Validate() error {
	if c.ReceiptID == "" || c.LocationID == "" {
		return errors.New("receipt_id and location_id are required")
	}
	return nil
}

// ============================================================================
// Receipt Command Handlers
// ============================================================================

type CreateReceiptHandler struct {
	receiptRepo   interfaces.ReceiptRepository
	warehouseRepo interfaces.WarehouseRepository
	kafkaProducer *kafka.Producer
}

func NewCreateReceiptHandler(
	receiptRepo interfaces.ReceiptRepository,
	warehouseRepo interfaces.WarehouseRepository,
	kafkaProducer *kafka.Producer,
) *CreateReceiptHandler {
	return &CreateReceiptHandler{
		receiptRepo:   receiptRepo,
		warehouseRepo: warehouseRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *CreateReceiptHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	createCmd := cmd.(*CreateReceiptCommand)

	// Validate warehouse exists
	warehouse, err := h.warehouseRepo.GetByID(ctx, createCmd.WarehouseID)
	if err != nil {
		return err
	}
	if warehouse == nil {
		return errors.New("warehouse not found")
	}

	receipt := &domain.Receipt{
		WarehouseID: createCmd.WarehouseID,
		ScanDate:    createCmd.ScanDate,
		TotalPrice:  createCmd.TotalPrice,
		ImagePath:   createCmd.ImagePath,
		OCRText:     createCmd.OCRText,
	}

	if err := h.receiptRepo.Create(ctx, receipt); err != nil {
		return err
	}

	createCmd.AggregateID = receipt.ID

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.FoodfolioReceiptCreatedEvent{
			ReceiptID:   receipt.ID,
			WarehouseID: receipt.WarehouseID,
			ScanDate:    receipt.ScanDate,
			TotalPrice:  receipt.TotalPrice,
			ImagePath:   receipt.ImagePath,
			OCRText:     receipt.OCRText,
			CreatedAt:   time.Now(),
		}
		if err := h.kafkaProducer.PublishFoodfolioReceiptCreated("foodfolio.receipt.created", event); err != nil {
			log.Printf("Warning: Failed to publish receipt created event: %v", err)
		}

		// If it has OCR text or image, also publish scanned event
		if receipt.ImagePath != nil || receipt.OCRText != nil {
			scannedEvent := kafka.FoodfolioReceiptScannedEvent{
				ReceiptID:   receipt.ID,
				WarehouseID: receipt.WarehouseID,
				ImagePath:   receipt.ImagePath,
				OCRText:     receipt.OCRText,
				ScannedAt:   time.Now(),
			}
			if err := h.kafkaProducer.PublishFoodfolioReceiptScanned("foodfolio.receipt.scanned", scannedEvent); err != nil {
				log.Printf("Warning: Failed to publish receipt scanned event: %v", err)
			}
		}
	}

	return nil
}

type UpdateReceiptHandler struct {
	receiptRepo   interfaces.ReceiptRepository
	warehouseRepo interfaces.WarehouseRepository
}

func NewUpdateReceiptHandler(
	receiptRepo interfaces.ReceiptRepository,
	warehouseRepo interfaces.WarehouseRepository,
) *UpdateReceiptHandler {
	return &UpdateReceiptHandler{
		receiptRepo:   receiptRepo,
		warehouseRepo: warehouseRepo,
	}
}

func (h *UpdateReceiptHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateReceiptCommand)

	receipt, err := h.receiptRepo.GetByID(ctx, updateCmd.AggregateID)
	if err != nil {
		return err
	}
	if receipt == nil {
		return errors.New("receipt not found")
	}

	// Validate warehouse exists
	warehouse, err := h.warehouseRepo.GetByID(ctx, updateCmd.WarehouseID)
	if err != nil {
		return err
	}
	if warehouse == nil {
		return errors.New("warehouse not found")
	}

	receipt.WarehouseID = updateCmd.WarehouseID
	receipt.ScanDate = updateCmd.ScanDate
	receipt.TotalPrice = updateCmd.TotalPrice

	return h.receiptRepo.Update(ctx, receipt)
}

type UpdateReceiptOCRDataHandler struct {
	receiptRepo interfaces.ReceiptRepository
}

func NewUpdateReceiptOCRDataHandler(
	receiptRepo interfaces.ReceiptRepository,
) *UpdateReceiptOCRDataHandler {
	return &UpdateReceiptOCRDataHandler{
		receiptRepo: receiptRepo,
	}
}

func (h *UpdateReceiptOCRDataHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateReceiptOCRDataCommand)

	receipt, err := h.receiptRepo.GetByID(ctx, updateCmd.AggregateID)
	if err != nil {
		return err
	}
	if receipt == nil {
		return errors.New("receipt not found")
	}

	// Update OCR-related fields
	receipt.ImagePath = &updateCmd.ImagePath
	receipt.OCRText = &updateCmd.OCRText
	receipt.TotalPrice = updateCmd.TotalPrice

	return h.receiptRepo.Update(ctx, receipt)
}

type DeleteReceiptHandler struct {
	receiptRepo   interfaces.ReceiptRepository
	kafkaProducer *kafka.Producer
}

func NewDeleteReceiptHandler(
	receiptRepo interfaces.ReceiptRepository,
	kafkaProducer *kafka.Producer,
) *DeleteReceiptHandler {
	return &DeleteReceiptHandler{
		receiptRepo:   receiptRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *DeleteReceiptHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteReceiptCommand)

	_, err := h.receiptRepo.GetByID(ctx, deleteCmd.AggregateID)
	if err != nil {
		return err
	}

	if err := h.receiptRepo.Delete(ctx, deleteCmd.AggregateID); err != nil {
		return err
	}

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.FoodfolioReceiptDeletedEvent{
			ReceiptID: deleteCmd.AggregateID,
			DeletedAt: time.Now(),
		}
		if err := h.kafkaProducer.PublishFoodfolioReceiptDeleted("foodfolio.receipt.deleted", event); err != nil {
			log.Printf("Warning: Failed to publish receipt deleted event: %v", err)
		}
	}

	return nil
}

// ============================================================================
// Receipt Item Command Handlers
// ============================================================================

type AddItemToReceiptHandler struct {
	receiptRepo interfaces.ReceiptRepository
	variantRepo interfaces.ItemVariantRepository
}

func NewAddItemToReceiptHandler(
	receiptRepo interfaces.ReceiptRepository,
	variantRepo interfaces.ItemVariantRepository,
) *AddItemToReceiptHandler {
	return &AddItemToReceiptHandler{
		receiptRepo: receiptRepo,
		variantRepo: variantRepo,
	}
}

func (h *AddItemToReceiptHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	addCmd := cmd.(*AddItemToReceiptCommand)

	// Validate receipt exists
	_, err := h.receiptRepo.GetByID(ctx, addCmd.ReceiptID)
	if err != nil {
		return err
	}

	// Validate variant if provided
	isMatched := false
	if addCmd.ItemVariantID != nil && *addCmd.ItemVariantID != "" {
		variant, err := h.variantRepo.GetByID(ctx, *addCmd.ItemVariantID)
		if err != nil {
			return err
		}
		if variant == nil {
			return errors.New("item variant not found")
		}
		isMatched = true
	}

	item := &domain.ReceiptItem{
		ReceiptID:     addCmd.ReceiptID,
		ItemVariantID: addCmd.ItemVariantID,
		ItemName:      addCmd.ItemName,
		Quantity:      addCmd.Quantity,
		UnitPrice:     addCmd.UnitPrice,
		TotalPrice:    addCmd.TotalPrice,
		ArticleNumber: addCmd.ArticleNumber,
		IsMatched:     isMatched,
	}

	if err := h.receiptRepo.AddItem(ctx, item); err != nil {
		return err
	}

	addCmd.ItemID = item.ID

	return nil
}

type UpdateReceiptItemHandler struct {
	receiptRepo interfaces.ReceiptRepository
}

func NewUpdateReceiptItemHandler(
	receiptRepo interfaces.ReceiptRepository,
) *UpdateReceiptItemHandler {
	return &UpdateReceiptItemHandler{
		receiptRepo: receiptRepo,
	}
}

func (h *UpdateReceiptItemHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateReceiptItemCommand)

	// Get item
	item, err := h.receiptRepo.GetItem(ctx, updateCmd.AggregateID)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.New("receipt item not found")
	}

	item.ItemName = updateCmd.ItemName
	item.Quantity = updateCmd.Quantity
	item.UnitPrice = updateCmd.UnitPrice
	item.TotalPrice = updateCmd.TotalPrice
	item.ArticleNumber = updateCmd.ArticleNumber

	return h.receiptRepo.UpdateItem(ctx, item)
}

type MatchReceiptItemHandler struct {
	receiptRepo interfaces.ReceiptRepository
	variantRepo interfaces.ItemVariantRepository
}

func NewMatchReceiptItemHandler(
	receiptRepo interfaces.ReceiptRepository,
	variantRepo interfaces.ItemVariantRepository,
) *MatchReceiptItemHandler {
	return &MatchReceiptItemHandler{
		receiptRepo: receiptRepo,
		variantRepo: variantRepo,
	}
}

func (h *MatchReceiptItemHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	matchCmd := cmd.(*MatchReceiptItemCommand)

	// Validate variant exists
	variant, err := h.variantRepo.GetByID(ctx, matchCmd.ItemVariantID)
	if err != nil {
		return err
	}
	if variant == nil {
		return errors.New("item variant not found")
	}

	return h.receiptRepo.MatchItem(ctx, matchCmd.AggregateID, matchCmd.ItemVariantID)
}

type AutoMatchReceiptItemsHandler struct {
	receiptRepo interfaces.ReceiptRepository
	variantRepo interfaces.ItemVariantRepository
}

func NewAutoMatchReceiptItemsHandler(
	receiptRepo interfaces.ReceiptRepository,
	variantRepo interfaces.ItemVariantRepository,
) *AutoMatchReceiptItemsHandler {
	return &AutoMatchReceiptItemsHandler{
		receiptRepo: receiptRepo,
		variantRepo: variantRepo,
	}
}

func (h *AutoMatchReceiptItemsHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	autoCmd := cmd.(*AutoMatchReceiptItemsCommand)

	// Get receipt items
	items, err := h.receiptRepo.GetItems(ctx, autoCmd.ReceiptID)
	if err != nil {
		return err
	}

	log.Printf("Auto-matching: Processing %d items for receipt %s", len(items), autoCmd.ReceiptID)

	matchedCount := 0
	unmatchedCount := 0

	for _, item := range items {
		if item.IsMatched {
			continue // Skip already matched items
		}

		// Try to match by name similarity (simplified implementation)
		variants, _, err := h.variantRepo.List(ctx, 0, 100, nil, nil, nil, false)
		if err != nil {
			log.Printf("Failed to load variants: %v", err)
			continue
		}

		log.Printf("Auto-matching item '%s' against %d variants", item.ItemName, len(variants))

		var bestMatch *domain.ItemVariant
		var bestSimilarity float64 = 0

		for _, variant := range variants {
			if variant.Item == nil {
				log.Printf("  Skipping variant %s: no item loaded", variant.ID)
				continue
			}

			// Calculate simple similarity
			fullName := variant.Item.Name + " " + variant.VariantName
			similarity := calculateSimpleSimilarity(strings.ToLower(item.ItemName), strings.ToLower(fullName))

			if similarity >= 0.5 { // Log all candidates above 50%
				log.Printf("  Candidate: '%s' = %.2f similarity", fullName, similarity)
			}

			if similarity > bestSimilarity && similarity >= autoCmd.SimilarityThreshold {
				bestSimilarity = similarity
				bestMatch = variant
			}
		}

		if bestMatch != nil {
			// Match found
			log.Printf("  MATCH FOUND: '%s' -> '%s %s' (%.2f)", item.ItemName, bestMatch.Item.Name, bestMatch.VariantName, bestSimilarity)
			if err := h.receiptRepo.MatchItem(ctx, item.ID, bestMatch.ID); err == nil {
				matchedCount++
			} else {
				log.Printf("  Failed to save match: %v", err)
			}
		} else {
			log.Printf("  NO MATCH for '%s' (best similarity: %.2f, threshold: %.2f)", item.ItemName, bestSimilarity, autoCmd.SimilarityThreshold)
			unmatchedCount++
		}
	}

	autoCmd.MatchedCount = matchedCount
	autoCmd.UnmatchedCount = unmatchedCount

	return nil
}

type CreateInventoryFromReceiptHandler struct {
	receiptRepo  interfaces.ReceiptRepository
	detailRepo   interfaces.ItemDetailRepository
	locationRepo interfaces.LocationRepository
}

func NewCreateInventoryFromReceiptHandler(
	receiptRepo interfaces.ReceiptRepository,
	detailRepo interfaces.ItemDetailRepository,
	locationRepo interfaces.LocationRepository,
) *CreateInventoryFromReceiptHandler {
	return &CreateInventoryFromReceiptHandler{
		receiptRepo:  receiptRepo,
		detailRepo:   detailRepo,
		locationRepo: locationRepo,
	}
}

func (h *CreateInventoryFromReceiptHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	invCmd := cmd.(*CreateInventoryFromReceiptCommand)

	// Validate receipt exists
	receipt, err := h.receiptRepo.GetByID(ctx, invCmd.ReceiptID)
	if err != nil {
		return err
	}
	if receipt == nil {
		return errors.New("receipt not found")
	}

	// Validate location exists
	location, err := h.locationRepo.GetByID(ctx, invCmd.LocationID)
	if err != nil {
		return err
	}
	if location == nil {
		return errors.New("location not found")
	}

	// Get receipt items
	items, err := h.receiptRepo.GetItems(ctx, invCmd.ReceiptID)
	if err != nil {
		return err
	}

	createdCount := 0

	for _, item := range items {
		// Skip if not matched and onlyMatched is true
		if invCmd.OnlyMatched && !item.IsMatched {
			continue
		}

		// Skip if no variant matched
		if item.ItemVariantID == nil {
			continue
		}

		// Create ItemDetails for this receipt item
		for i := 0; i < item.Quantity; i++ {
			detail := &domain.ItemDetail{
				ItemVariantID: *item.ItemVariantID,
				WarehouseID:   receipt.WarehouseID,
				LocationID:    invCmd.LocationID,
				ArticleNumber: item.ArticleNumber,
				PurchasePrice: item.UnitPrice,
				PurchaseDate:  receipt.ScanDate,
				ExpiryDate:    invCmd.DefaultExpiryDate,
				HasDeposit:    false, // Could be enhanced
				IsFrozen:      false, // Could be enhanced
				IsOpened:      false,
			}

			if err := h.detailRepo.Create(ctx, detail); err != nil {
				continue // Skip on error
			}

			createdCount++
		}
	}

	invCmd.CreatedCount = createdCount

	return nil
}

// Helper function for simple similarity calculation
func calculateSimpleSimilarity(s1, s2 string) float64 {
	// Very simplified similarity check
	if strings.Contains(s2, s1) || strings.Contains(s1, s2) {
		return 0.9
	}

	// Normalize strings
	normalize := func(s string) string {
		s = strings.ReplaceAll(s, ".", " ")
		s = strings.ReplaceAll(s, "_", " ")
		return strings.ToLower(s)
	}

	norm1 := normalize(s1)
	norm2 := normalize(s2)

	// Check normalized contains
	if strings.Contains(norm2, norm1) || strings.Contains(norm1, norm2) {
		return 0.85
	}

	// Count matching words
	words1 := strings.Fields(norm1)
	words2 := strings.Fields(norm2)

	matchingWords := 0
	for _, w1 := range words1 {
		for _, w2 := range words2 {
			if w1 == w2 {
				matchingWords++
				break
			}
			if len(w1) >= 3 && len(w2) >= 3 {
				if strings.HasPrefix(w2, w1) || strings.HasPrefix(w1, w2) {
					matchingWords++
					break
				}
				if strings.Contains(w2, w1) || strings.Contains(w1, w2) {
					matchingWords++
					break
				}
			}
		}
	}

	if len(words1) == 0 {
		return 0
	}

	return float64(matchingWords) / float64(len(words1))
}
