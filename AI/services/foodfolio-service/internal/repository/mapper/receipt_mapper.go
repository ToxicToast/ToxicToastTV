package mapper

import (
	"gorm.io/gorm"
	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/entity"
)

// ReceiptToEntity converts domain model to database entity
func ReceiptToEntity(receipt *domain.Receipt) *entity.ReceiptEntity {
	if receipt == nil {
		return nil
	}

	e := &entity.ReceiptEntity{
		ID:          receipt.ID,
		WarehouseID: receipt.WarehouseID,
		ScanDate:    receipt.ScanDate,
		TotalPrice:  receipt.TotalPrice,
		ImagePath:   receipt.ImagePath,
		OCRText:     receipt.OCRText,
		CreatedAt:   receipt.CreatedAt,
		UpdatedAt:   receipt.UpdatedAt,
	}

	// Convert DeletedAt
	if receipt.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *receipt.DeletedAt,
			Valid: true,
		}
	}

	return e
}

// ReceiptToDomain converts database entity to domain model
func ReceiptToDomain(e *entity.ReceiptEntity) *domain.Receipt {
	if e == nil {
		return nil
	}

	receipt := &domain.Receipt{
		ID:          e.ID,
		WarehouseID: e.WarehouseID,
		ScanDate:    e.ScanDate,
		TotalPrice:  e.TotalPrice,
		ImagePath:   e.ImagePath,
		OCRText:     e.OCRText,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		receipt.DeletedAt = &deletedAt
	}

	// Convert Items
	if e.Items != nil {
		receipt.Items = ReceiptItemsToDomain(e.Items)
	}

	return receipt
}

// ReceiptsToDomain converts slice of entities to domain models
func ReceiptsToDomain(entities []*entity.ReceiptEntity) []*domain.Receipt {
	receipts := make([]*domain.Receipt, 0, len(entities))
	for _, e := range entities {
		receipts = append(receipts, ReceiptToDomain(e))
	}
	return receipts
}
