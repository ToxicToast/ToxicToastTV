package domain

import (
	"time"

)

// Item represents a base product (e.g., "Monster Energy", "Coca Cola")
type Item struct {
	ID         string         
	Name       string         
	Slug       string         
	CategoryID string         
	CompanyID  string         
	TypeID     string         
	CreatedAt  time.Time      
	UpdatedAt  time.Time      
	DeletedAt  *time.Time 

	// Relations
	Category     *Category     
	Company      *Company      
	Type         *Type         
	ItemVariants []ItemVariant 
}

