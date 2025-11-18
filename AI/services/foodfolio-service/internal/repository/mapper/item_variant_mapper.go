package mapper

import (
	"gorm.io/gorm"
	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/entity"
)

// ItemVariantToEntity converts domain model to database entity
func ItemVariantToEntity(variant *domain.ItemVariant) *entity.ItemVariantEntity {
	if variant == nil {
		return nil
	}

	e := &entity.ItemVariantEntity{
		ID:               variant.ID,
		ItemID:           variant.ItemID,
		SizeID:           variant.SizeID,
		VariantName:      variant.VariantName,
		Slug:             variant.Slug,
		Barcode:          variant.Barcode,
		MinSKU:           variant.MinSKU,
		MaxSKU:           variant.MaxSKU,
		IsNormallyFrozen: variant.IsNormallyFrozen,
		CreatedAt:        variant.CreatedAt,
		UpdatedAt:        variant.UpdatedAt,
	}

	// Convert DeletedAt
	if variant.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *variant.DeletedAt,
			Valid: true,
		}
	}

	return e
}

// ItemVariantToDomain converts database entity to domain model
func ItemVariantToDomain(e *entity.ItemVariantEntity) *domain.ItemVariant {
	if e == nil {
		return nil
	}

	variant := &domain.ItemVariant{
		ID:               e.ID,
		ItemID:           e.ItemID,
		SizeID:           e.SizeID,
		VariantName:      e.VariantName,
		Slug:             e.Slug,
		Barcode:          e.Barcode,
		MinSKU:           e.MinSKU,
		MaxSKU:           e.MaxSKU,
		IsNormallyFrozen: e.IsNormallyFrozen,
		CreatedAt:        e.CreatedAt,
		UpdatedAt:        e.UpdatedAt,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		variant.DeletedAt = &deletedAt
	}

	// Map preloaded relations
	if e.Item != nil {
		variant.Item = ItemToDomain(e.Item)
	}

	if e.Size != nil {
		variant.Size = SizeToDomain(e.Size)
	}

	return variant
}

// ItemVariantsToDomain converts slice of entities to domain models
func ItemVariantsToDomain(entities []*entity.ItemVariantEntity) []*domain.ItemVariant {
	variants := make([]*domain.ItemVariant, 0, len(entities))
	for _, e := range entities {
		variants = append(variants, ItemVariantToDomain(e))
	}
	return variants
}
