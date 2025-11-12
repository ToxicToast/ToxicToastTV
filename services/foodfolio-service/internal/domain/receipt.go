package domain

import (
	"time"

)

// Receipt represents a scanned receipt (Kassenbon)
type Receipt struct {
	ID         string         
	WarehouseID string                    // Where purchased
	ScanDate    time.Time                                    // When scanned
	TotalPrice  float64                 // Total from receipt
	ImagePath   *string                            // Path to scanned image
	OCRText     *string                                      // Raw OCR output
	CreatedAt   time.Time     
	UpdatedAt   time.Time     
	DeletedAt   *time.Time 

	// Relations
	Warehouse *Warehouse     
	Items     []*ReceiptItem 
}


// ReceiptItem represents an item on a receipt
type ReceiptItem struct {
	ID            string         
	ReceiptID     string         
	ItemVariantID *string                         // Matched item (nullable if not matched)
	ItemName      string                     // Name from receipt
	Quantity      int            
	UnitPrice     float64        
	TotalPrice    float64        
	ArticleNumber *string                        // From receipt
	IsMatched     bool                          // Successfully matched to item variant
	CreatedAt     time.Time      
	UpdatedAt     time.Time      
	DeletedAt     *time.Time 

	// Relations
	Receipt     *Receipt     
	ItemVariant *ItemVariant 
}

