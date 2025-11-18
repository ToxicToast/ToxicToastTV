package domain

import (
	"time"

)

// Type represents packaging type (e.g., "Can", "PET Bottle", "Box", "Bag", "Glass")
type Type struct {
	ID        string         
	Name      string         
	Slug      string         
	CreatedAt time.Time      
	UpdatedAt time.Time      
	DeletedAt *time.Time 

	// Relations
	Items []Item 
}

