package domain

import (
	"time"

)

// Location represents a storage location at home (e.g., "Fridge", "Pantry", "Freezer")
type Location struct {
	ID        string         
	Name      string         
	Slug      string         
	ParentID  *string                                 // e.g., "Fridge" -> "Top Shelf"
	CreatedAt time.Time      
	UpdatedAt time.Time      
	DeletedAt *time.Time 

	// Relations
	Parent      *Location    
	Children    []Location   
	ItemDetails []ItemDetail 
}

