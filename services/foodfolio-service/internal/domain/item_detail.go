package domain

import (
	"time"

)

// ItemDetail represents one physical item/unit
// Each can/bottle/package is tracked individually
type ItemDetail struct {
	ID            string         
	ItemVariantID string         
	WarehouseID   string                      // Where purchased
	LocationID    string                       // Where stored
	ArticleNumber *string                          // Shop's article number
	PurchasePrice float64                // Actual paid price
	PurchaseDate  time.Time                                  // When bought
	ExpiryDate    *time.Time     `json:"expiry_date"`                                              // MHD (best before date)
	OpenedDate    *time.Time     `json:"opened_date"`                                              // When opened
	IsOpened      bool                             // Currently opened
	HasDeposit    bool                           // Pfand
	IsFrozen      bool                             // Currently frozen
	CreatedAt     time.Time      
	UpdatedAt     time.Time      
	DeletedAt     *time.Time 

	// Relations
	ItemVariant *ItemVariant 
	Warehouse   *Warehouse   
	Location    *Location    
}


// IsExpired checks if the item has expired
func (id *ItemDetail) IsExpired() bool {
	if id.ExpiryDate == nil {
		return false
	}
	return time.Now().After(*id.ExpiryDate)
}

// IsExpiringSoon checks if item expires within given days
func (id *ItemDetail) IsExpiringSoon(days int) bool {
	if id.ExpiryDate == nil {
		return false
	}
	threshold := time.Now().AddDate(0, 0, days)
	return id.ExpiryDate.Before(threshold)
}

// IsConsumed checks if item is likely consumed (soft-deleted or opened long ago)
func (id *ItemDetail) IsConsumed() bool {
	// If soft-deleted, it's consumed
	if id.DeletedAt != nil {
		return true
	}
	// If opened more than 30 days ago, consider consumed
	if id.OpenedDate != nil && time.Since(*id.OpenedDate) > 30*24*time.Hour {
		return true
	}
	return false
}
