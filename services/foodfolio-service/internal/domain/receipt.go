package domain

import (
	"time"

	"gorm.io/gorm"
)

// Receipt represents a scanned receipt (Kassenbon)
type Receipt struct {
	ID          string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	WarehouseID string         `gorm:"type:uuid;not null;index" json:"warehouse_id"`            // Where purchased
	ScanDate    time.Time      `gorm:"not null" json:"scan_date"`                               // When scanned
	TotalPrice  float64        `gorm:"type:decimal(10,2);not null" json:"total_price"`          // Total from receipt
	ImagePath   *string        `gorm:"type:varchar(500)" json:"image_path"`                     // Path to scanned image
	OCRText     *string        `gorm:"type:text" json:"ocr_text"`                               // Raw OCR output
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relations
	Warehouse    *Warehouse     `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	ReceiptItems []ReceiptItem  `gorm:"foreignKey:ReceiptID" json:"items,omitempty"`
}

func (Receipt) TableName() string {
	return "receipts"
}

// ReceiptItem represents an item on a receipt
type ReceiptItem struct {
	ID            string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	ReceiptID     string         `gorm:"type:uuid;not null;index" json:"receipt_id"`
	ItemVariantID *string        `gorm:"type:uuid;index" json:"item_variant_id"`                 // Matched item (nullable if not matched)
	ItemName      string         `gorm:"type:varchar(255);not null" json:"item_name"`            // Name from receipt
	Quantity      int            `gorm:"not null;default:1" json:"quantity"`
	UnitPrice     float64        `gorm:"type:decimal(10,2);not null" json:"unit_price"`
	TotalPrice    float64        `gorm:"type:decimal(10,2);not null" json:"total_price"`
	ArticleNumber *string        `gorm:"type:varchar(255)" json:"article_number"`                // From receipt
	IsMatched     bool           `gorm:"not null;default:false" json:"is_matched"`               // Successfully matched to item variant
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relations
	Receipt     *Receipt     `gorm:"foreignKey:ReceiptID" json:"receipt,omitempty"`
	ItemVariant *ItemVariant `gorm:"foreignKey:ItemVariantID" json:"item_variant,omitempty"`
}

func (ReceiptItem) TableName() string {
	return "receipt_items"
}
