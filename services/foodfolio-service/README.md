# Foodfolio Service

A comprehensive food inventory management system with receipt scanning, expiry tracking, and intelligent stock alerts.

## Overview

The Foodfolio Service manages your household food inventory with advanced features including:
- Individual item tracking with expiry dates (MHD)
- Receipt scanning with OCR
- Smart inventory alerts (min/max stock levels)
- Shopping list management
- Hierarchical categories and storage locations
- Purchase tracking by store

## Architecture

Built using **Clean Architecture** principles with:
- **Domain Layer**: Core business entities and logic
- **Repository Layer**: Data access abstraction
- **Use Case Layer**: Business operations
- **Handler Layer**: gRPC API endpoints

## Domain Model

### Hierarchical Item Structure

```
Item (e.g., "Monster Energy")
└── ItemVariant (e.g., "Monster Energy Original 500ml")
    ├── MinSKU: 3 (alert when stock below this)
    ├── MaxSKU: 10 (alert when stock above this)
    ├── Barcode: EAN code
    └── ItemDetail (individual physical can/bottle)
        ├── PurchasePrice: €1.99 (actual paid price)
        ├── PurchaseDate: 2025-01-15
        ├── ExpiryDate: 2025-06-01 (MHD)
        ├── OpenedDate: null
        ├── IsOpened: false
        ├── Warehouse: "Rewe" (where purchased)
        └── Location: "Kühlschrank" (where stored)
```

### Core Entities

#### 1. **Item** (Base Product)
The master product entry (e.g., "Coca-Cola", "Monster Energy").

**Fields:**
- Name, Slug
- Category (hierarchical)
- Company (brand)
- Type (packaging type)

#### 2. **ItemVariant** (Flavor + Size Combination)
Specific variants of an item with different flavors or sizes.

**Fields:**
- VariantName (e.g., "Original", "Ultra White")
- Size (e.g., "500ml", "1L")
- Barcode (EAN/UPC)
- **MinSKU** - Minimum stock level (alert when below)
- **MaxSKU** - Maximum stock level (alert when above)
- IsNormallyFrozen - Whether this variant is typically frozen

**Helper Methods:**
- `CurrentStock()` - Calculates current inventory from ItemDetails
- `NeedsRestock()` - Checks if stock is below MinSKU
- `IsOverstocked()` - Checks if stock is above MaxSKU

#### 3. **ItemDetail** (Physical Item)
Each entry represents **one physical item** (1 can, 1 bottle, 1 package).

**Fields:**
- ArticleNumber (store's product number)
- PurchasePrice (actual price paid)
- PurchaseDate
- **ExpiryDate** (MHD - Mindesthaltbarkeitsdatum)
- OpenedDate (when opened)
- IsOpened
- HasDeposit (Pfand)
- IsFrozen (current state)
- Warehouse (where purchased)
- Location (where stored)

**Helper Methods:**
- `IsExpired()` - Checks if past expiry date
- `IsExpiringSoon(days)` - Checks if expiring within N days
- `IsConsumed()` - Checks if opened and past expiry

#### 4. **Category** (Hierarchical)
Product categories with parent-child relationships.

**Examples:**
- Beverages → Energy Drinks → Sugar-Free Energy Drinks
- Food → Dairy → Milk

**Fields:**
- Name, Slug
- ParentID (optional)

#### 5. **Company** (Brand/Manufacturer)
Product brands and manufacturers.

**Examples:**
- Monster
- Coca-Cola
- Red Bull

#### 6. **Type** (Packaging Type)
Type of packaging or container.

**Examples:**
- Can (Dose)
- PET Bottle
- Glass Bottle
- Box (Packung)
- Bag (Tüte)

#### 7. **Size**
Product size with numeric value and unit.

**Fields:**
- Name (e.g., "500ml", "1L")
- Value (numeric: 500, 1)
- Unit (e.g., "ml", "L", "g", "kg")

#### 8. **Warehouse** (Purchase Location)
Stores where items are purchased.

**Examples:**
- Rewe
- Lidl
- Aldi
- Amazon

**Note:** This is NOT a storage location - it tracks where items were bought.

#### 9. **Location** (Storage Location)
Home storage locations (hierarchical).

**Examples:**
- Kühlschrank (Fridge)
  - Top Shelf
  - Middle Shelf
  - Vegetable Drawer
- Vorratsschrank (Pantry)
- Gefrierschrank (Freezer)

#### 10. **Shoppinglist**
Shopping lists with items to purchase.

**Fields:**
- Name (e.g., "Weekly Shopping", "Party Supplies")
- ShoppinglistItems (join table)

#### 11. **ShoppinglistItem**
Items on a shopping list.

**Fields:**
- ItemVariant (what to buy)
- Quantity (how many)
- IsPurchased (already bought)

#### 12. **Receipt** (with OCR)
Scanned receipts from purchases.

**Fields:**
- Warehouse (store)
- ScanDate
- TotalPrice
- ImagePath (path to scanned image)
- OCRText (raw OCR output)
- ReceiptItems

#### 13. **ReceiptItem**
Individual items from a receipt.

**Fields:**
- ItemName (from receipt)
- Quantity
- UnitPrice, TotalPrice
- ArticleNumber (from receipt)
- ItemVariantID (nullable - matched product)
- IsMatched (successfully matched to catalog)

## Key Features

### 1. Stock Level Alerts
- **Min/Max SKU tracking** per ItemVariant
- Alerts when stock falls below minimum
- Alerts when stock exceeds maximum (prevent over-purchasing)

### 2. Expiry Date Tracking (MHD)
- Individual expiry dates for each physical item
- `IsExpired()` - Check if already expired
- `IsExpiringSoon(days)` - Get items expiring within N days
- Event publishing for expiring items

### 3. Receipt Scanning with OCR
- Upload receipt images (JPEG, PNG, PDF)
- OCR text extraction (Tesseract)
- Automatic item matching to catalog
- Manual matching for unrecognized items

### 4. Purchase Tracking
- Track where each item was purchased (Warehouse)
- Track actual purchase prices
- Receipt history

### 5. Storage Management
- Hierarchical storage locations
- Track where each item is stored
- Move items between locations

### 6. Shopping List
- Create multiple shopping lists
- Add items with quantities
- Mark items as purchased
- Auto-generate from low stock alerts

### 7. Deposit Tracking (Pfand)
- Track which items have deposits
- Filter items with deposits

### 8. Frozen Item Management
- Track if items are normally frozen
- Track current frozen state
- Track opened dates for frozen items

## Configuration

### Environment Variables

```bash
# Server Configuration
PORT=8081                    # HTTP server port
GRPC_PORT=9091              # gRPC server port
ENVIRONMENT=development
LOG_LEVEL=info

# Authentication
AUTH_ENABLED=false          # Set to true for Keycloak auth

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=foodfolio
DB_PASSWORD=your_password_here
DB_NAME=foodfolio
DB_SSL_MODE=disable

# Kafka/Redpanda
KAFKA_BROKERS=localhost:19092
KAFKA_GROUP_ID=foodfolio-service
KAFKA_TOPIC_PREFIX=foodfolio
KAFKA_TOPIC_ITEM_EVENTS=foodfolio.events.item
KAFKA_TOPIC_INVENTORY_EVENTS=foodfolio.events.inventory
KAFKA_TOPIC_RECEIPT_EVENTS=foodfolio.events.receipt
KAFKA_TOPIC_ALERT_EVENTS=foodfolio.events.alert

# Keycloak (if AUTH_ENABLED=true)
KEYCLOAK_URL=http://localhost:8080
KEYCLOAK_REALM=your-realm
KEYCLOAK_CLIENT_ID=foodfolio-service

# OCR / Receipt Scanning
OCR_STORAGE_PATH=./receipts
OCR_MAX_SIZE=10485760       # 10MB
OCR_ALLOWED_TYPES=image/jpeg,image/png,application/pdf
OCR_PROVIDER=tesseract
```

## Event Publishing

The service publishes domain events to Kafka:

### Item Events (`foodfolio.events.item`)
- ItemCreated
- ItemUpdated
- ItemDeleted
- ItemVariantCreated
- ItemVariantUpdated
- ItemVariantDeleted

### Inventory Events (`foodfolio.events.inventory`)
- ItemDetailCreated
- ItemDetailOpened
- ItemDetailConsumed
- ItemDetailMoved (location change)

### Receipt Events (`foodfolio.events.receipt`)
- ReceiptScanned
- ReceiptItemMatched
- ReceiptProcessed

### Alert Events (`foodfolio.events.alert`)
- StockBelowMinimum
- StockAboveMaximum
- ItemExpiringSoon
- ItemExpired

## API Endpoints

### gRPC Services

**CategoryService**
- CreateCategory
- GetCategory
- ListCategories
- UpdateCategory
- DeleteCategory

**CompanyService**
- CreateCompany
- GetCompany
- ListCompanies
- UpdateCompany
- DeleteCompany

**ItemService**
- CreateItem
- GetItem
- ListItems
- UpdateItem
- DeleteItem
- GetItemWithVariants
- SearchItems

**ItemVariantService**
- CreateItemVariant
- GetItemVariant
- ListItemVariants
- UpdateItemVariant
- DeleteItemVariant
- GetCurrentStock
- GetLowStockVariants
- GetOverstockedVariants

**ItemDetailService**
- CreateItemDetail
- GetItemDetail
- ListItemDetails
- UpdateItemDetail
- DeleteItemDetail
- OpenItem
- GetExpiringItems
- GetExpiredItems

**ReceiptService**
- UploadReceipt
- ProcessReceipt
- GetReceipt
- ListReceipts
- MatchReceiptItem

**ShoppinglistService**
- CreateShoppinglist
- GetShoppinglist
- ListShoppinglists
- AddItem
- RemoveItem
- MarkItemPurchased
- DeleteShoppinglist

**LocationService**
- CreateLocation
- GetLocation
- ListLocations
- UpdateLocation
- DeleteLocation

**WarehouseService**
- CreateWarehouse
- GetWarehouse
- ListWarehouses
- UpdateWarehouse
- DeleteWarehouse

### HTTP Endpoints

- `GET /health` - Health check
- `GET /health/ready` - Readiness probe (checks database)
- `GET /health/live` - Liveness probe

## Database Schema

PostgreSQL with GORM auto-migrations.

### Tables
- `categories` (with self-referencing parent_id)
- `companies`
- `types`
- `sizes`
- `warehouses`
- `locations` (with self-referencing parent_id)
- `items`
- `item_variants`
- `item_details`
- `shoppinglists`
- `shoppinglist_items` (join table)
- `receipts`
- `receipt_items`

All tables include:
- UUID primary keys (auto-generated)
- Timestamps (created_at, updated_at)
- Soft deletes (deleted_at)

## Development

### Prerequisites
- Go 1.24+
- PostgreSQL 14+
- Kafka/Redpanda
- Tesseract OCR (for receipt scanning)

### Setup

1. Copy environment file:
```bash
cp .env.example .env
```

2. Update database credentials in `.env`

3. Install dependencies:
```bash
go mod download
```

4. Run database migrations (automatic on startup):
```bash
go run cmd/server/main.go
```

### Running the Service

```bash
# Development mode
go run cmd/server/main.go

# Build binary
go build -o bin/foodfolio-service cmd/server/main.go

# Run binary
./bin/foodfolio-service
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/domain
```

## Use Cases

### Example 1: Add New Product Purchase

1. **Create or find Item** (e.g., "Monster Energy")
2. **Create or find ItemVariant** (e.g., "Monster Original 500ml")
3. **Create ItemDetails** for each physical can purchased:
   ```go
   // Bought 6 cans at Rewe for €1.99 each, expires 2025-06-01
   for i := 0; i < 6; i++ {
       CreateItemDetail({
           ItemVariantID: variantID,
           WarehouseID: reweID,
           LocationID: fridgeID,
           PurchasePrice: 1.99,
           PurchaseDate: time.Now(),
           ExpiryDate: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
       })
   }
   ```

### Example 2: Scan Receipt and Add Inventory

1. **Upload receipt image**
2. **OCR processing** extracts text
3. **Create Receipt** with total price
4. **Create ReceiptItems** from OCR data
5. **Match items** to existing ItemVariants (manual or automatic)
6. **Create ItemDetails** for matched items

### Example 3: Check Expiring Items

```go
// Get items expiring in next 7 days
expiringItems := GetExpiringItems(7)

// Send alert events
for _, item := range expiringItems {
    PublishEvent("foodfolio.events.alert", ItemExpiringSoon{
        ItemDetailID: item.ID,
        ItemName: item.ItemVariant.Item.Name,
        ExpiryDate: item.ExpiryDate,
        Location: item.Location.Name,
    })
}
```

### Example 4: Stock Alert System

```go
// Check all variants for stock levels
variants := GetAllItemVariants()

for _, variant := range variants {
    currentStock := variant.CurrentStock()

    if currentStock < variant.MinSKU {
        PublishEvent("foodfolio.events.alert", StockBelowMinimum{
            ItemVariantID: variant.ID,
            ItemName: variant.Item.Name,
            VariantName: variant.VariantName,
            CurrentStock: currentStock,
            MinSKU: variant.MinSKU,
        })
    }

    if currentStock > variant.MaxSKU {
        PublishEvent("foodfolio.events.alert", StockAboveMaximum{
            ItemVariantID: variant.ID,
            CurrentStock: currentStock,
            MaxSKU: variant.MaxSKU,
        })
    }
}
```

## Technology Stack

- **Go 1.24** - Programming language
- **gRPC** - API protocol
- **Protocol Buffers** - API contracts
- **PostgreSQL** - Database
- **GORM** - ORM
- **Kafka/Redpanda** - Event streaming
- **Tesseract** - OCR engine
- **Keycloak** - Authentication (optional)
- **Gorilla Mux** - HTTP routing

## License

See root LICENSE file. This is private software for personal use only.

## Related Services

- **Blog Service** - Blog and content management
- **Warcraft Service** - Blizzard API integration
- **Twitchbot Service** - Twitch stream tracking
- **Notification Service** - Cross-service notifications
- **SSE Service** - Real-time frontend updates
- **Gateway Service** - API gateway
