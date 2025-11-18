package impl

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/entity"
	"toxictoast/services/foodfolio-service/internal/repository/interfaces"
	"toxictoast/services/foodfolio-service/internal/repository/mapper"
)

type receiptRepository struct {
	db *gorm.DB
}

// NewReceiptRepository creates a new receipt repository instance
func NewReceiptRepository(db *gorm.DB) interfaces.ReceiptRepository {
	return &receiptRepository{db: db}
}

func (r *receiptRepository) Create(ctx context.Context, receipt *domain.Receipt) error {
	e := mapper.ReceiptToEntity(receipt)
	if err := r.db.WithContext(ctx).Create(e).Error; err != nil {
		return err
	}
	// Copy back the generated ID and timestamps
	receipt.ID = e.ID
	receipt.CreatedAt = e.CreatedAt
	receipt.UpdatedAt = e.UpdatedAt
	return nil
}

func (r *receiptRepository) GetByID(ctx context.Context, id string) (*domain.Receipt, error) {
	var e entity.ReceiptEntity
	err := r.db.WithContext(ctx).
		Preload("Warehouse").
		Preload("Items").
		Preload("Items.ItemVariant").
		Preload("Items.ItemVariant.Item").
		Preload("Items.ItemVariant.Size").
		First(&e, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ReceiptToDomain(&e), nil
}

func (r *receiptRepository) List(ctx context.Context, offset, limit int, warehouseID *string, startDate, endDate *time.Time, includeDeleted bool) ([]*domain.Receipt, int64, error) {
	var entities []*entity.ReceiptEntity
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.ReceiptEntity{})

	if includeDeleted {
		query = query.Unscoped()
	}

	// Filters
	if warehouseID != nil && *warehouseID != "" {
		query = query.Where("warehouse_id = ?", *warehouseID)
	}

	if startDate != nil {
		query = query.Where("scan_date >= ?", *startDate)
	}

	if endDate != nil {
		query = query.Where("scan_date <= ?", *endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Preload("Warehouse").
		Preload("Items").
		Order("scan_date DESC").
		Offset(offset).
		Limit(limit).
		Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return mapper.ReceiptsToDomain(entities), total, nil
}

func (r *receiptRepository) Update(ctx context.Context, receipt *domain.Receipt) error {
	e := mapper.ReceiptToEntity(receipt)
	// Use Updates to only update the receipt fields, not associations
	return r.db.WithContext(ctx).Model(&entity.ReceiptEntity{}).
		Where("id = ?", e.ID).
		Updates(map[string]interface{}{
			"warehouse_id": e.WarehouseID,
			"scan_date":    e.ScanDate,
			"total_price":  e.TotalPrice,
			"image_path":   e.ImagePath,
			"ocr_text":     e.OCRText,
		}).Error
}

func (r *receiptRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.ReceiptEntity{}, "id = ?", id).Error
}

func (r *receiptRepository) HardDelete(ctx context.Context, id string) error {
	// Also delete associated items
	r.db.WithContext(ctx).Unscoped().Where("receipt_id = ?", id).Delete(&entity.ReceiptItemEntity{})
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.ReceiptEntity{}, "id = ?", id).Error
}

func (r *receiptRepository) AddItem(ctx context.Context, item *domain.ReceiptItem) error {
	e := mapper.ReceiptItemToEntity(item)
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *receiptRepository) UpdateItem(ctx context.Context, item *domain.ReceiptItem) error {
	e := mapper.ReceiptItemToEntity(item)
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *receiptRepository) MatchItem(ctx context.Context, receiptItemID, itemVariantID string) error {
	return r.db.WithContext(ctx).
		Model(&entity.ReceiptItemEntity{}).
		Where("id = ?", receiptItemID).
		Updates(map[string]interface{}{
			"item_variant_id": itemVariantID,
			"is_matched":      true,
		}).Error
}

func (r *receiptRepository) GetItem(ctx context.Context, itemID string) (*domain.ReceiptItem, error) {
	var e entity.ReceiptItemEntity
	err := r.db.WithContext(ctx).
		Preload("ItemVariant").
		Preload("ItemVariant.Item").
		Preload("ItemVariant.Size").
		First(&e, "id = ?", itemID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ReceiptItemToDomain(&e), nil
}

func (r *receiptRepository) GetItems(ctx context.Context, receiptID string) ([]*domain.ReceiptItem, error) {
	var entities []*entity.ReceiptItemEntity

	err := r.db.WithContext(ctx).
		Preload("ItemVariant").
		Preload("ItemVariant.Item").
		Preload("ItemVariant.Size").
		Where("receipt_id = ?", receiptID).
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	return mapper.ReceiptItemsToDomain(entities), nil
}

func (r *receiptRepository) GetUnmatchedItems(ctx context.Context, receiptID string) ([]*domain.ReceiptItem, error) {
	var entities []*entity.ReceiptItemEntity

	err := r.db.WithContext(ctx).
		Where("receipt_id = ? AND is_matched = ?", receiptID, false).
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	return mapper.ReceiptItemsToDomain(entities), nil
}

func (r *receiptRepository) GetTotalAmount(ctx context.Context, warehouseID *string, startDate, endDate *time.Time) (float64, error) {
	var total float64

	query := r.db.WithContext(ctx).Model(&entity.ReceiptEntity{})

	if warehouseID != nil && *warehouseID != "" {
		query = query.Where("warehouse_id = ?", *warehouseID)
	}

	if startDate != nil {
		query = query.Where("scan_date >= ?", *startDate)
	}

	if endDate != nil {
		query = query.Where("scan_date <= ?", *endDate)
	}

	err := query.Select("COALESCE(SUM(total_price), 0)").Scan(&total).Error
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (r *receiptRepository) GetStatistics(ctx context.Context, warehouseID *string, startDate, endDate *time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	query := r.db.WithContext(ctx).Model(&entity.ReceiptEntity{})

	if warehouseID != nil && *warehouseID != "" {
		query = query.Where("warehouse_id = ?", *warehouseID)
	}

	if startDate != nil {
		query = query.Where("scan_date >= ?", *startDate)
	}

	if endDate != nil {
		query = query.Where("scan_date <= ?", *endDate)
	}

	// Total receipts
	var totalReceipts int64
	if err := query.Count(&totalReceipts).Error; err != nil {
		return nil, err
	}
	stats["total_receipts"] = totalReceipts

	// Total amount
	var totalAmount float64
	if err := query.Select("COALESCE(SUM(total_price), 0)").Scan(&totalAmount).Error; err != nil {
		return nil, err
	}
	stats["total_amount"] = totalAmount

	// Average amount
	var avgAmount float64
	if totalReceipts > 0 {
		avgAmount = totalAmount / float64(totalReceipts)
	}
	stats["average_amount"] = avgAmount

	// Total items
	var totalItems int64
	itemQuery := r.db.WithContext(ctx).Model(&entity.ReceiptItemEntity{})

	if warehouseID != nil || startDate != nil || endDate != nil {
		itemQuery = itemQuery.Joins("JOIN receipts ON receipts.id = receipt_items.receipt_id")

		if warehouseID != nil && *warehouseID != "" {
			itemQuery = itemQuery.Where("receipts.warehouse_id = ?", *warehouseID)
		}

		if startDate != nil {
			itemQuery = itemQuery.Where("receipts.scan_date >= ?", *startDate)
		}

		if endDate != nil {
			itemQuery = itemQuery.Where("receipts.scan_date <= ?", *endDate)
		}
	}

	if err := itemQuery.Count(&totalItems).Error; err != nil {
		return nil, err
	}
	stats["total_items"] = totalItems

	// Average items per receipt
	var avgItemsPerReceipt float64
	if totalReceipts > 0 {
		avgItemsPerReceipt = float64(totalItems) / float64(totalReceipts)
	}
	stats["average_items_per_receipt"] = avgItemsPerReceipt

	// Items by warehouse
	type WarehouseCount struct {
		WarehouseID string
		Count       int64
	}
	var warehouseCounts []WarehouseCount

	warehouseQuery := r.db.WithContext(ctx).
		Model(&entity.ReceiptEntity{}).
		Select("warehouse_id, COUNT(*) as count").
		Group("warehouse_id")

	if startDate != nil {
		warehouseQuery = warehouseQuery.Where("scan_date >= ?", *startDate)
	}

	if endDate != nil {
		warehouseQuery = warehouseQuery.Where("scan_date <= ?", *endDate)
	}

	if err := warehouseQuery.Scan(&warehouseCounts).Error; err != nil {
		return nil, err
	}

	itemsByWarehouse := make(map[string]int64)
	for _, wc := range warehouseCounts {
		itemsByWarehouse[wc.WarehouseID] = wc.Count
	}
	stats["items_by_warehouse"] = itemsByWarehouse

	return stats, nil
}
