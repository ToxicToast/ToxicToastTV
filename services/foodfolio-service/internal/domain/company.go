package domain

import (
	"time"

)

// Company represents a brand/manufacturer (e.g., "Monster", "Coca-Cola", "Müller")
type Company struct {
	ID        string         
	Name      string         
	Slug      string         
	CreatedAt time.Time      
	UpdatedAt time.Time      
	DeletedAt *time.Time 

	// Relations
	Items []Item 
}

