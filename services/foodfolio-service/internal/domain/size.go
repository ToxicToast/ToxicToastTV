package domain

import (
	"time"

)

// Size represents product size (e.g., "500ml", "1L", "250g", "1kg")
type Size struct {
	ID        string         
	Name      string         
	Value     float64                                      // Numeric value (e.g., 500, 1, 250)
	Unit      string                       // Unit (e.g., "ml", "L", "g", "kg")
	CreatedAt time.Time      
	UpdatedAt time.Time      
	DeletedAt *time.Time 

	// Relations
	ItemVariants []ItemVariant 
}

