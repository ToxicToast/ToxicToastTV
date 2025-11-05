package impl

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/interfaces"
)

type receiptRepository struct {
	db *gorm.DB
}

// NewReceiptRepository creates a new receipt repository instance
func NewReceiptRepository(db *gorm.DB) interfaces.ReceiptRepository {
	return &receiptRepository{db: db}
}

func (r *receiptRepository) Create(ctx context.Context, receipt *domain.Receipt) error {
	return r.db.WithContext(ctx).Create(receipt).Error
}

func (r *receiptRepository) GetByID(ctx context.Context, id string) (*domain.Receipt, error) {
	var receipt domain.Receipt
	err := r.db.WithContext(ctx).
		Preload("Warehouse").
		Preload("ReceiptItems").
		Preload("ReceiptItems.ItemVariant").
		Preload("ReceiptItems.ItemVariant.Item").
		Preload("ReceiptItems.ItemVariant.Size").
		First(&receipt, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &receipt, nil
}

func (r *receiptRepository) List(ctx context.Context, offset, limit int, warehouseID *string, startDate, endDate *time.Time, includeDeleted bool) ([]*domain.Receipt, int64, error) {
	var receipts []*domain.Receipt
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Receipt{})

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
		Preload("ReceiptItems").
		Order("scan_date DESC").
		Offset(offset).
		Limit(limit).
		Find(&receipts).Error; err != nil {
		return nil, 0, err
	}

	return receipts, total, nil
}

func (r *receiptRepository) Update(ctx context.Context, receipt *domain.Receipt) error {
	return r.db.WithContext(ctx).Save(receipt).Error
}

func (r *receiptRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Receipt{}, "id = ?", id).Error
}

func (r *receiptRepository) HardDelete(ctx context.Context, id string) error {
	// Also delete associated items
	r.db.WithContext(ctx).Unscoped().Where("receipt_id = ?", id).Delete(&domain.ReceiptItem{})
	return r.db.WithContext(ctx).Unscoped().Delete(&domain.Receipt{}, "id = ?", id).Error
}

func (r *receiptRepository) AddItem(ctx context.Context, item *domain.ReceiptItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *receiptRepository) UpdateItem(ctx context.Context, item *domain.ReceiptItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *receiptRepository) MatchItem(ctx context.Context, receiptItemID, itemVariantID string) error {
	return r.db.WithContext(ctx).
		Model(&domain.ReceiptItem{}).
		Where("id = ?", receiptItemID).
		Updates(map[string]interface{}{
			"item_variant_id": itemVariantID,
			"is_matched":      true,
		}).Error
}

func (r *receiptRepository) GetItem(ctx context.Context, itemID string) (*domain.ReceiptItem, error) {
	var item domain.ReceiptItem
	err := r.db.WithContext(ctx).
		Preload("ItemVariant").
		Preload("ItemVariant.Item").
		Preload("ItemVariant.Size").
		First(&item, "id = ?", itemID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *receiptRepository) GetItems(ctx context.Context, receiptID string) ([]*domain.ReceiptItem, error) {
	var items []*domain.ReceiptItem

	err := r.db.WithContext(ctx).
		Preload("ItemVariant").
		Preload("ItemVariant.Item").
		Preload("ItemVariant.Size").
		Where("receipt_id = ?", receiptID).
		Find(&items).Error

	if err != nil {
		return nil, err
	}

	return items, nil
}

func (r *receiptRepository) GetUnmatchedItems(ctx context.Context, receiptID string) ([]*domain.ReceiptItem, error) {
	var items []*domain.ReceiptItem

	err := r.db.WithContext(ctx).
		Where("receipt_id = ? AND is_matched = ?", receiptID, false).
		Find(&items).Error

	if err != nil {
		return nil, err
	}

	return items, nil
}

func (r *receiptRepository) GetTotalAmount(ctx context.Context, warehouseID *string, startDate, endDate *time.Time) (float64, error) {
	var total float64

	query := r.db.WithContext(ctx).Model(&domain.Receipt{})

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

	query := r.db.WithContext(ctx).Model(&domain.Receipt{})

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
	itemQuery := r.db.WithContext(ctx).Model(&domain.ReceiptItem{})

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
		Model(&domain.Receipt{}).
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
