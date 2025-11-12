package kafka

import "time"

// Category Events
type CategoryCreatedEvent struct {
	CategoryID  string    `json:"category_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	ParentID    *string   `json:"parent_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type CategoryUpdatedEvent struct {
	CategoryID  string    `json:"category_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	ParentID    *string   `json:"parent_id,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CategoryDeletedEvent struct {
	CategoryID string    `json:"category_id"`
	DeletedAt  time.Time `json:"deleted_at"`
}

// Post Events
type PostCreatedEvent struct {
	PostID    string    `json:"post_id"`
	Title     string    `json:"title"`
	Slug      string    `json:"slug"`
	AuthorID  string    `json:"author_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type PostUpdatedEvent struct {
	PostID    string    `json:"post_id"`
	Title     string    `json:"title"`
	Slug      string    `json:"slug"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PostPublishedEvent struct {
	PostID      string    `json:"post_id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	AuthorID    string    `json:"author_id"`
	PublishedAt time.Time `json:"published_at"`
}

type PostDeletedEvent struct {
	PostID    string    `json:"post_id"`
	DeletedAt time.Time `json:"deleted_at"`
}

// Tag Events
type TagCreatedEvent struct {
	TagID     string    `json:"tag_id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

type TagUpdatedEvent struct {
	TagID     string    `json:"tag_id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TagDeletedEvent struct {
	TagID     string    `json:"tag_id"`
	DeletedAt time.Time `json:"deleted_at"`
}

// Comment Events
type CommentCreatedEvent struct {
	CommentID   string    `json:"comment_id"`
	PostID      string    `json:"post_id"`
	AuthorName  string    `json:"author_name"`
	AuthorEmail string    `json:"author_email"`
	Content     string    `json:"content"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type CommentModeratedEvent struct {
	CommentID   string    `json:"comment_id"`
	PostID      string    `json:"post_id"`
	OldStatus   string    `json:"old_status"`
	NewStatus   string    `json:"new_status"`
	ModeratedAt time.Time `json:"moderated_at"`
}

type CommentDeletedEvent struct {
	CommentID string    `json:"comment_id"`
	PostID    string    `json:"post_id"`
	DeletedAt time.Time `json:"deleted_at"`
}

type CommentApprovedEvent struct {
	CommentID  string    `json:"comment_id"`
	PostID     string    `json:"post_id"`
	ApprovedAt time.Time `json:"approved_at"`
}

type CommentRejectedEvent struct {
	CommentID  string    `json:"comment_id"`
	PostID     string    `json:"post_id"`
	Reason     string    `json:"reason"`
	RejectedAt time.Time `json:"rejected_at"`
}

// Media Events
type MediaUploadedEvent struct {
	MediaID          string    `json:"media_id"`
	Filename         string    `json:"filename"`
	OriginalFilename string    `json:"original_filename"`
	MimeType         string    `json:"mime_type"`
	Size             int64     `json:"size"`
	URL              string    `json:"url"`
	UploadedBy       string    `json:"uploaded_by"`
	UploadedAt       time.Time `json:"uploaded_at"`
}

type MediaDeletedEvent struct {
	MediaID   string    `json:"media_id"`
	Filename  string    `json:"filename"`
	DeletedAt time.Time `json:"deleted_at"`
}

type MediaThumbnailGeneratedEvent struct {
	MediaID      string    `json:"media_id"`
	ThumbnailURL string    `json:"thumbnail_url"`
	GeneratedAt  time.Time `json:"generated_at"`
}

// Link Events
type LinkCreatedEvent struct {
	LinkID      string     `json:"link_id"`
	OriginalURL string     `json:"original_url"`
	ShortCode   string     `json:"short_code"`
	CustomAlias *string    `json:"custom_alias,omitempty"`
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
}

type LinkUpdatedEvent struct {
	LinkID      string     `json:"link_id"`
	OriginalURL string     `json:"original_url"`
	ShortCode   string     `json:"short_code"`
	CustomAlias *string    `json:"custom_alias,omitempty"`
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	IsActive    bool       `json:"is_active"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type LinkDeletedEvent struct {
	LinkID    string    `json:"link_id"`
	ShortCode string    `json:"short_code"`
	DeletedAt time.Time `json:"deleted_at"`
}

type LinkActivatedEvent struct {
	LinkID      string    `json:"link_id"`
	ShortCode   string    `json:"short_code"`
	ActivatedAt time.Time `json:"activated_at"`
}

type LinkDeactivatedEvent struct {
	LinkID        string    `json:"link_id"`
	ShortCode     string    `json:"short_code"`
	DeactivatedAt time.Time `json:"deactivated_at"`
}

type LinkExpiredEvent struct {
	LinkID    string    `json:"link_id"`
	ShortCode string    `json:"short_code"`
	ExpiresAt time.Time `json:"expires_at"`
}

type LinkClickedEvent struct {
	ClickID    string     `json:"click_id"`
	LinkID     string     `json:"link_id"`
	ShortCode  string     `json:"short_code"`
	IPAddress  string     `json:"ip_address"`
	UserAgent  string     `json:"user_agent"`
	Referer    *string    `json:"referer,omitempty"`
	Country    *string    `json:"country,omitempty"`
	City       *string    `json:"city,omitempty"`
	DeviceType *string    `json:"device_type,omitempty"`
	ClickedAt  time.Time  `json:"clicked_at"`
}

type LinkClickFraudDetectedEvent struct {
	ClickID   string    `json:"click_id"`
	LinkID    string    `json:"link_id"`
	ShortCode string    `json:"short_code"`
	IPAddress string    `json:"ip_address"`
	Reason    string    `json:"reason"`
	DetectedAt time.Time `json:"detected_at"`
}

// Foodfolio Category Events
type FoodfolioCategoryCreatedEvent struct {
	CategoryID string    `json:"category_id"`
	Name       string    `json:"name"`
	Slug       string    `json:"slug"`
	ParentID   *string   `json:"parent_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type FoodfolioCategoryUpdatedEvent struct {
	CategoryID string    `json:"category_id"`
	Name       string    `json:"name"`
	Slug       string    `json:"slug"`
	ParentID   *string   `json:"parent_id,omitempty"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type FoodfolioCategoryDeletedEvent struct {
	CategoryID string    `json:"category_id"`
	DeletedAt  time.Time `json:"deleted_at"`
}

// Foodfolio Company Events
type FoodfolioCompanyCreatedEvent struct {
	CompanyID string    `json:"company_id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

type FoodfolioCompanyUpdatedEvent struct {
	CompanyID string    `json:"company_id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FoodfolioCompanyDeletedEvent struct {
	CompanyID string    `json:"company_id"`
	DeletedAt time.Time `json:"deleted_at"`
}

// Foodfolio Item Events
type FoodfolioItemCreatedEvent struct {
	ItemID     string    `json:"item_id"`
	Name       string    `json:"name"`
	Slug       string    `json:"slug"`
	CategoryID string    `json:"category_id"`
	CompanyID  string    `json:"company_id"`
	TypeID     string    `json:"type_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type FoodfolioItemUpdatedEvent struct {
	ItemID     string    `json:"item_id"`
	Name       string    `json:"name"`
	Slug       string    `json:"slug"`
	CategoryID string    `json:"category_id"`
	CompanyID  string    `json:"company_id"`
	TypeID     string    `json:"type_id"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type FoodfolioItemDeletedEvent struct {
	ItemID    string    `json:"item_id"`
	DeletedAt time.Time `json:"deleted_at"`
}

// Foodfolio Item Variant Events
type FoodfolioVariantCreatedEvent struct {
	VariantID        string    `json:"variant_id"`
	ItemID           string    `json:"item_id"`
	SizeID           string    `json:"size_id"`
	VariantName      string    `json:"variant_name"`
	Barcode          *string   `json:"barcode,omitempty"`
	MinSKU           int       `json:"min_sku"`
	MaxSKU           int       `json:"max_sku"`
	IsNormallyFrozen bool      `json:"is_normally_frozen"`
	CreatedAt        time.Time `json:"created_at"`
}

type FoodfolioVariantUpdatedEvent struct {
	VariantID        string    `json:"variant_id"`
	VariantName      string    `json:"variant_name"`
	Barcode          *string   `json:"barcode,omitempty"`
	MinSKU           int       `json:"min_sku"`
	MaxSKU           int       `json:"max_sku"`
	IsNormallyFrozen bool      `json:"is_normally_frozen"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type FoodfolioVariantDeletedEvent struct {
	VariantID string    `json:"variant_id"`
	DeletedAt time.Time `json:"deleted_at"`
}

type FoodfolioVariantStockLowEvent struct {
	VariantID    string    `json:"variant_id"`
	ItemID       string    `json:"item_id"`
	VariantName  string    `json:"variant_name"`
	CurrentStock int       `json:"current_stock"`
	MinSKU       int       `json:"min_sku"`
	DetectedAt   time.Time `json:"detected_at"`
}

type FoodfolioVariantStockEmptyEvent struct {
	VariantID   string    `json:"variant_id"`
	ItemID      string    `json:"item_id"`
	VariantName string    `json:"variant_name"`
	DetectedAt  time.Time `json:"detected_at"`
}

// Foodfolio Item Detail Events
type FoodfolioDetailCreatedEvent struct {
	DetailID      string     `json:"detail_id"`
	VariantID     string     `json:"variant_id"`
	WarehouseID   string     `json:"warehouse_id"`
	LocationID    string     `json:"location_id"`
	ArticleNumber *string    `json:"article_number,omitempty"`
	PurchasePrice float64    `json:"purchase_price"`
	PurchaseDate  time.Time  `json:"purchase_date"`
	ExpiryDate    *time.Time `json:"expiry_date,omitempty"`
	HasDeposit    bool       `json:"has_deposit"`
	IsFrozen      bool       `json:"is_frozen"`
	CreatedAt     time.Time  `json:"created_at"`
}

type FoodfolioDetailOpenedEvent struct {
	DetailID  string    `json:"detail_id"`
	VariantID string    `json:"variant_id"`
	OpenedAt  time.Time `json:"opened_at"`
}

type FoodfolioDetailExpiredEvent struct {
	DetailID   string     `json:"detail_id"`
	VariantID  string     `json:"variant_id"`
	ExpiryDate *time.Time `json:"expiry_date,omitempty"`
	DetectedAt time.Time  `json:"detected_at"`
}

type FoodfolioDetailExpiringSoonEvent struct {
	DetailID   string     `json:"detail_id"`
	VariantID  string     `json:"variant_id"`
	ExpiryDate *time.Time `json:"expiry_date,omitempty"`
	DaysLeft   int        `json:"days_left"`
	DetectedAt time.Time  `json:"detected_at"`
}

type FoodfolioDetailConsumedEvent struct {
	DetailID   string    `json:"detail_id"`
	VariantID  string    `json:"variant_id"`
	ConsumedAt time.Time `json:"consumed_at"`
}

type FoodfolioDetailMovedEvent struct {
	DetailID      string    `json:"detail_id"`
	VariantID     string    `json:"variant_id"`
	OldLocationID string    `json:"old_location_id"`
	NewLocationID string    `json:"new_location_id"`
	MovedAt       time.Time `json:"moved_at"`
}

type FoodfolioDetailFrozenEvent struct {
	DetailID  string    `json:"detail_id"`
	VariantID string    `json:"variant_id"`
	FrozenAt  time.Time `json:"frozen_at"`
}

type FoodfolioDetailThawedEvent struct {
	DetailID  string    `json:"detail_id"`
	VariantID string    `json:"variant_id"`
	ThawedAt  time.Time `json:"thawed_at"`
}

// Foodfolio Location Events
type FoodfolioLocationCreatedEvent struct {
	LocationID string    `json:"location_id"`
	Name       string    `json:"name"`
	Slug       string    `json:"slug"`
	ParentID   *string   `json:"parent_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type FoodfolioLocationUpdatedEvent struct {
	LocationID string    `json:"location_id"`
	Name       string    `json:"name"`
	Slug       string    `json:"slug"`
	ParentID   *string   `json:"parent_id,omitempty"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type FoodfolioLocationDeletedEvent struct {
	LocationID string    `json:"location_id"`
	DeletedAt  time.Time `json:"deleted_at"`
}

// Foodfolio Warehouse Events
type FoodfolioWarehouseCreatedEvent struct {
	WarehouseID string    `json:"warehouse_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	CreatedAt   time.Time `json:"created_at"`
}

type FoodfolioWarehouseUpdatedEvent struct {
	WarehouseID string    `json:"warehouse_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type FoodfolioWarehouseDeletedEvent struct {
	WarehouseID string    `json:"warehouse_id"`
	DeletedAt   time.Time `json:"deleted_at"`
}

// Foodfolio Receipt Events
type FoodfolioReceiptCreatedEvent struct {
	ReceiptID   string     `json:"receipt_id"`
	WarehouseID string     `json:"warehouse_id"`
	ScanDate    time.Time  `json:"scan_date"`
	TotalPrice  float64    `json:"total_price"`
	ImagePath   *string    `json:"image_path,omitempty"`
	OCRText     *string    `json:"ocr_text,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type FoodfolioReceiptScannedEvent struct {
	ReceiptID   string     `json:"receipt_id"`
	WarehouseID string     `json:"warehouse_id"`
	ImagePath   *string    `json:"image_path,omitempty"`
	OCRText     *string    `json:"ocr_text,omitempty"`
	ScannedAt   time.Time  `json:"scanned_at"`
}

type FoodfolioReceiptDeletedEvent struct {
	ReceiptID string    `json:"receipt_id"`
	DeletedAt time.Time `json:"deleted_at"`
}

// Foodfolio Shopping List Events
type FoodfolioShoppinglistCreatedEvent struct {
	ShoppinglistID string    `json:"shoppinglist_id"`
	Name           string    `json:"name"`
	CreatedAt      time.Time `json:"created_at"`
}

type FoodfolioShoppinglistUpdatedEvent struct {
	ShoppinglistID string    `json:"shoppinglist_id"`
	Name           string    `json:"name"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type FoodfolioShoppinglistDeletedEvent struct {
	ShoppinglistID string    `json:"shoppinglist_id"`
	DeletedAt      time.Time `json:"deleted_at"`
}

type FoodfolioShoppinglistItemAddedEvent struct {
	ShoppinglistID string    `json:"shoppinglist_id"`
	ItemID         string    `json:"item_id"`
	VariantID      string    `json:"variant_id"`
	Quantity       int       `json:"quantity"`
	AddedAt        time.Time `json:"added_at"`
}

type FoodfolioShoppinglistItemRemovedEvent struct {
	ShoppinglistID string    `json:"shoppinglist_id"`
	ItemID         string    `json:"item_id"`
	RemovedAt      time.Time `json:"removed_at"`
}

type FoodfolioShoppinglistItemPurchasedEvent struct {
	ShoppinglistID string    `json:"shoppinglist_id"`
	ItemID         string    `json:"item_id"`
	VariantID      string    `json:"variant_id"`
	Quantity       int       `json:"quantity"`
	PurchasedAt    time.Time `json:"purchased_at"`
}
