package domain

import (
	"time"

)

// Warehouse represents a store where items are purchased (e.g., "Rewe", "Lidl", "Edeka")
type Warehouse struct {
	ID        string         
	Name      string         
	Slug      string         
	CreatedAt time.Time      
	UpdatedAt time.Time      
	DeletedAt *time.Time 

	// Relations
	ItemDetails []ItemDetail 
	Receipts    []Receipt    
}

