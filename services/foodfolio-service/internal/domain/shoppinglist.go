package domain

import (
	"time"

)

// Shoppinglist represents a shopping list
type Shoppinglist struct {
	ID        string         
	Name      string                            // e.g., "Weekly Shopping", "Party Supplies"
	CreatedAt time.Time      
	UpdatedAt time.Time      
	DeletedAt *time.Time 

	// Relations
	Items []*ShoppinglistItem 
}


// ShoppinglistItem represents an item on a shopping list
type ShoppinglistItem struct {
	ID               string         
	ShoppinglistID   string         
	ItemVariantID    string         
	Quantity         int                                  // How many to buy
	IsPurchased      bool                         // Already bought
	CreatedAt        time.Time      
	UpdatedAt        time.Time      
	DeletedAt        *time.Time 

	// Relations
	Shoppinglist *Shoppinglist 
	ItemVariant  *ItemVariant  
}

