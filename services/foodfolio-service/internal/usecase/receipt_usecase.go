package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/interfaces"
)

var (
	ErrReceiptNotFound     = errors.New("receipt not found")
	ErrReceiptItemNotFound = errors.New("receipt item not found")
	ErrInvalidReceiptData  = errors.New("invalid receipt data")
)

type ReceiptUseCase interface {
	CreateReceipt(ctx context.Context, warehouseID string, scanDate time.Time, totalPrice float64, imagePath, ocrText *string) (*domain.Receipt, error)
	GetReceiptByID(ctx context.Context, id string) (*domain.Receipt, error)
	ListReceipts(ctx context.Context, page, pageSize int, warehouseID *string, startDate, endDate *time.Time, includeDeleted bool) ([]*domain.Receipt, int64, error)
	UpdateReceipt(ctx context.Context, id, warehouseID string, scanDate time.Time, totalPrice float64) (*domain.Receipt, error)
	DeleteReceipt(ctx context.Context, id string) error
	AddItemToReceipt(ctx context.Context, receiptID, itemName string, quantity int, unitPrice, totalPrice float64, articleNumber *string, itemVariantID *string) (*domain.ReceiptItem, error)
	UpdateReceiptItem(ctx context.Context, itemID, itemName string, quantity int, unitPrice, totalPrice float64, articleNumber *string) (*domain.ReceiptItem, error)
	MatchReceiptItem(ctx context.Context, receiptItemID, itemVariantID string) (*domain.ReceiptItem, error)
	AutoMatchReceiptItems(ctx context.Context, receiptID string, similarityThreshold float64) (int, int, error)
	GetUnmatchedItems(ctx context.Context, receiptID string) ([]*domain.ReceiptItem, error)
	GetStatistics(ctx context.Context, warehouseID *string, startDate, endDate *time.Time) (map[string]interface{}, error)
	CreateInventoryFromReceipt(ctx context.Context, receiptID, locationID string, defaultExpiryDate *time.Time, onlyMatched bool) (int, error)
}

type receiptUseCase struct {
	receiptRepo   interfaces.ReceiptRepository
	warehouseRepo interfaces.WarehouseRepository
	variantRepo   interfaces.ItemVariantRepository
	detailRepo    interfaces.ItemDetailRepository
	locationRepo  interfaces.LocationRepository
}

func NewReceiptUseCase(
	receiptRepo interfaces.ReceiptRepository,
	warehouseRepo interfaces.WarehouseRepository,
	variantRepo interfaces.ItemVariantRepository,
	detailRepo interfaces.ItemDetailRepository,
	locationRepo interfaces.LocationRepository,
) ReceiptUseCase {
	return &receiptUseCase{
		receiptRepo:   receiptRepo,
		warehouseRepo: warehouseRepo,
		variantRepo:   variantRepo,
		detailRepo:    detailRepo,
		locationRepo:  locationRepo,
	}
}

func (uc *receiptUseCase) CreateReceipt(ctx context.Context, warehouseID string, scanDate time.Time, totalPrice float64, imagePath, ocrText *string) (*domain.Receipt, error) {
	// Validate
	if warehouseID == "" {
		return nil, ErrInvalidReceiptData
	}

	if totalPrice < 0 {
		return nil, errors.New("total price cannot be negative")
	}

	// Validate warehouse exists
	warehouse, err := uc.warehouseRepo.GetByID(ctx, warehouseID)
	if err != nil {
		return nil, err
	}
	if warehouse == nil {
		return nil, errors.New("warehouse not found")
	}

	receipt := &domain.Receipt{
		WarehouseID: warehouseID,
		ScanDate:    scanDate,
		TotalPrice:  totalPrice,
		ImagePath:   imagePath,
		OCRText:     ocrText,
	}

	if err := uc.receiptRepo.Create(ctx, receipt); err != nil {
		return nil, err
	}

	return receipt, nil
}

func (uc *receiptUseCase) GetReceiptByID(ctx context.Context, id string) (*domain.Receipt, error) {
	receipt, err := uc.receiptRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if receipt == nil {
		return nil, ErrReceiptNotFound
	}

	return receipt, nil
}

func (uc *receiptUseCase) ListReceipts(ctx context.Context, page, pageSize int, warehouseID *string, startDate, endDate *time.Time, includeDeleted bool) ([]*domain.Receipt, int64, error) {
	offset := (page - 1) * pageSize

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return uc.receiptRepo.List(ctx, offset, pageSize, warehouseID, startDate, endDate, includeDeleted)
}

func (uc *receiptUseCase) UpdateReceipt(ctx context.Context, id, warehouseID string, scanDate time.Time, totalPrice float64) (*domain.Receipt, error) {
	receipt, err := uc.GetReceiptByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate
	if warehouseID == "" {
		return nil, ErrInvalidReceiptData
	}

	if totalPrice < 0 {
		return nil, errors.New("total price cannot be negative")
	}

	// Validate warehouse exists
	warehouse, err := uc.warehouseRepo.GetByID(ctx, warehouseID)
	if err != nil {
		return nil, err
	}
	if warehouse == nil {
		return nil, errors.New("warehouse not found")
	}

	receipt.WarehouseID = warehouseID
	receipt.ScanDate = scanDate
	receipt.TotalPrice = totalPrice

	if err := uc.receiptRepo.Update(ctx, receipt); err != nil {
		return nil, err
	}

	return receipt, nil
}

func (uc *receiptUseCase) DeleteReceipt(ctx context.Context, id string) error {
	_, err := uc.GetReceiptByID(ctx, id)
	if err != nil {
		return err
	}

	return uc.receiptRepo.Delete(ctx, id)
}

func (uc *receiptUseCase) AddItemToReceipt(ctx context.Context, receiptID, itemName string, quantity int, unitPrice, totalPrice float64, articleNumber *string, itemVariantID *string) (*domain.ReceiptItem, error) {
	// Validate
	if receiptID == "" || itemName == "" {
		return nil, ErrInvalidReceiptData
	}

	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than 0")
	}

	if unitPrice < 0 || totalPrice < 0 {
		return nil, errors.New("prices cannot be negative")
	}

	// Validate receipt exists
	_, err := uc.GetReceiptByID(ctx, receiptID)
	if err != nil {
		return nil, err
	}

	// Validate variant if provided
	isMatched := false
	if itemVariantID != nil && *itemVariantID != "" {
		variant, err := uc.variantRepo.GetByID(ctx, *itemVariantID)
		if err != nil {
			return nil, err
		}
		if variant == nil {
			return nil, errors.New("item variant not found")
		}
		isMatched = true
	}

	item := &domain.ReceiptItem{
		ReceiptID:     receiptID,
		ItemVariantID: itemVariantID,
		ItemName:      itemName,
		Quantity:      quantity,
		UnitPrice:     unitPrice,
		TotalPrice:    totalPrice,
		ArticleNumber: articleNumber,
		IsMatched:     isMatched,
	}

	if err := uc.receiptRepo.AddItem(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (uc *receiptUseCase) UpdateReceiptItem(ctx context.Context, itemID, itemName string, quantity int, unitPrice, totalPrice float64, articleNumber *string) (*domain.ReceiptItem, error) {
	// Get item
	item, err := uc.receiptRepo.GetItem(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrReceiptItemNotFound
	}

	// Validate
	if itemName == "" {
		return nil, ErrInvalidReceiptData
	}

	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than 0")
	}

	if unitPrice < 0 || totalPrice < 0 {
		return nil, errors.New("prices cannot be negative")
	}

	item.ItemName = itemName
	item.Quantity = quantity
	item.UnitPrice = unitPrice
	item.TotalPrice = totalPrice
	item.ArticleNumber = articleNumber

	if err := uc.receiptRepo.UpdateItem(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (uc *receiptUseCase) MatchReceiptItem(ctx context.Context, receiptItemID, itemVariantID string) (*domain.ReceiptItem, error) {
	// Validate variant exists
	variant, err := uc.variantRepo.GetByID(ctx, itemVariantID)
	if err != nil {
		return nil, err
	}
	if variant == nil {
		return nil, errors.New("item variant not found")
	}

	if err := uc.receiptRepo.MatchItem(ctx, receiptItemID, itemVariantID); err != nil {
		return nil, err
	}

	// Reload item
	return uc.receiptRepo.GetItem(ctx, receiptItemID)
}

func (uc *receiptUseCase) AutoMatchReceiptItems(ctx context.Context, receiptID string, similarityThreshold float64) (int, int, error) {
	// Get receipt items
	items, err := uc.receiptRepo.GetItems(ctx, receiptID)
	if err != nil {
		return 0, 0, err
	}

	matchedCount := 0
	unmatchedCount := 0

	for _, item := range items {
		if item.IsMatched {
			continue // Skip already matched items
		}

		// Try to match by article number first
		if item.ArticleNumber != nil && *item.ArticleNumber != "" {
			// Search variants by article number (simplified - in production use proper search)
			// For now, skip article number matching
		}

		// Try to match by name similarity (simplified implementation)
		// In production, use proper fuzzy matching algorithm
		variants, _, err := uc.variantRepo.List(ctx, 0, 100, nil, nil, nil, false)
		if err != nil {
			continue
		}

		var bestMatch *domain.ItemVariant
		var bestSimilarity float64 = 0

		for _, variant := range variants {
			if variant.Item == nil {
				continue
			}

			// Calculate simple similarity (in production, use Levenshtein or similar)
			similarity := calculateSimpleSimilarity(strings.ToLower(item.ItemName), strings.ToLower(variant.Item.Name+" "+variant.VariantName))

			if similarity > bestSimilarity && similarity >= similarityThreshold {
				bestSimilarity = similarity
				bestMatch = variant
			}
		}

		if bestMatch != nil {
			// Match found
			if err := uc.receiptRepo.MatchItem(ctx, item.ID, bestMatch.ID); err == nil {
				matchedCount++
			}
		} else {
			unmatchedCount++
		}
	}

	return matchedCount, unmatchedCount, nil
}

func (uc *receiptUseCase) GetUnmatchedItems(ctx context.Context, receiptID string) ([]*domain.ReceiptItem, error) {
	// Validate receipt exists
	_, err := uc.GetReceiptByID(ctx, receiptID)
	if err != nil {
		return nil, err
	}

	return uc.receiptRepo.GetUnmatchedItems(ctx, receiptID)
}

func (uc *receiptUseCase) GetStatistics(ctx context.Context, warehouseID *string, startDate, endDate *time.Time) (map[string]interface{}, error) {
	return uc.receiptRepo.GetStatistics(ctx, warehouseID, startDate, endDate)
}

func (uc *receiptUseCase) CreateInventoryFromReceipt(ctx context.Context, receiptID, locationID string, defaultExpiryDate *time.Time, onlyMatched bool) (int, error) {
	// Validate receipt exists
	receipt, err := uc.GetReceiptByID(ctx, receiptID)
	if err != nil {
		return 0, err
	}

	// Validate location exists
	location, err := uc.locationRepo.GetByID(ctx, locationID)
	if err != nil {
		return 0, err
	}
	if location == nil {
		return 0, errors.New("location not found")
	}

	// Get receipt items
	items, err := uc.receiptRepo.GetItems(ctx, receiptID)
	if err != nil {
		return 0, err
	}

	createdCount := 0

	for _, item := range items {
		// Skip if not matched and onlyMatched is true
		if onlyMatched && !item.IsMatched {
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
				LocationID:    locationID,
				ArticleNumber: item.ArticleNumber,
				PurchasePrice: item.UnitPrice,
				PurchaseDate:  receipt.ScanDate,
				ExpiryDate:    defaultExpiryDate,
				HasDeposit:    false, // Could be enhanced
				IsFrozen:      false, // Could be enhanced
				IsOpened:      false,
			}

			if err := uc.detailRepo.Create(ctx, detail); err != nil {
				continue // Skip on error
			}

			createdCount++
		}
	}

	return createdCount, nil
}

// Helper function for simple similarity calculation
func calculateSimpleSimilarity(s1, s2 string) float64 {
	// Very simplified similarity check - in production use proper algorithm
	if strings.Contains(s2, s1) || strings.Contains(s1, s2) {
		return 0.8
	}

	// Count matching words
	words1 := strings.Fields(s1)
	words2 := strings.Fields(s2)

	matchingWords := 0
	for _, w1 := range words1 {
		for _, w2 := range words2 {
			if w1 == w2 {
				matchingWords++
				break
			}
		}
	}

	if len(words1) == 0 {
		return 0
	}

	return float64(matchingWords) / float64(len(words1))
}
