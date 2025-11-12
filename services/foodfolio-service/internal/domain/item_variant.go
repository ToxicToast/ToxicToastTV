package domain

import (
	"time"

)

// ItemVariant represents a specific variant (flavor + size combination)
// Example: "Monster Energy Original 500ml" vs "Monster Energy Ultra White 500ml"
type ItemVariant struct {
	ID                 string         
	ItemID             string         
	SizeID             string         
	VariantName        string                 // e.g., "Original", "Ultra White", "Zero"
	Slug               string         
	Barcode            *string                  // EAN/UPC
	MinSKU             int                                 // Alert when below
	MaxSKU             int                                 // Alert when above
	IsNormallyFrozen   bool                 // Should this be frozen
	CreatedAt          time.Time      
	UpdatedAt          time.Time      
	DeletedAt          *time.Time 

	// Relations
	Item        *Item        
	Size        *Size        
	ItemDetails []ItemDetail 
}


// CurrentStock returns the count of active (not opened/consumed) ItemDetails
func (iv *ItemVariant) CurrentStock() int {
	count := 0
	for _, detail := range iv.ItemDetails {
		if !detail.IsOpened && !detail.IsConsumed() {
			count++
		}
	}
	return count
}
