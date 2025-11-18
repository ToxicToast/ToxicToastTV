package domain

import (
	"time"

)

// Category represents a product category (e.g., "Beverages", "Dairy", "Snacks")
type Category struct {
	ID        string         
	Name      string         
	Slug      string         
	ParentID  *string        
	CreatedAt time.Time      
	UpdatedAt time.Time      
	DeletedAt *time.Time 

	// Relations
	Parent   *Category  
	Children []Category 
	Items    []Item     
}

