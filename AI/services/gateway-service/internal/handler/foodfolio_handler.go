package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	pb "toxictoast/services/foodfolio-service/api/proto"
	"google.golang.org/grpc"
)

// FoodFolioHandler handles HTTP-to-gRPC translation for foodfolio service
type FoodFolioHandler struct {
	categoryClient     pb.CategoryServiceClient
	companyClient      pb.CompanyServiceClient
	typeClient         pb.TypeServiceClient
	sizeClient         pb.SizeServiceClient
	warehouseClient    pb.WarehouseServiceClient
	locationClient     pb.LocationServiceClient
	itemClient         pb.ItemServiceClient
	itemVariantClient  pb.ItemVariantServiceClient
	itemDetailClient   pb.ItemDetailServiceClient
	shoppinglistClient pb.ShoppinglistServiceClient
	receiptClient      pb.ReceiptServiceClient
}

// NewFoodFolioHandler creates a new foodfolio handler
func NewFoodFolioHandler(conn *grpc.ClientConn) *FoodFolioHandler {
	return &FoodFolioHandler{
		categoryClient:     pb.NewCategoryServiceClient(conn),
		companyClient:      pb.NewCompanyServiceClient(conn),
		typeClient:         pb.NewTypeServiceClient(conn),
		sizeClient:         pb.NewSizeServiceClient(conn),
		warehouseClient:    pb.NewWarehouseServiceClient(conn),
		locationClient:     pb.NewLocationServiceClient(conn),
		itemClient:         pb.NewItemServiceClient(conn),
		itemVariantClient:  pb.NewItemVariantServiceClient(conn),
		itemDetailClient:   pb.NewItemDetailServiceClient(conn),
		shoppinglistClient: pb.NewShoppinglistServiceClient(conn),
		receiptClient:      pb.NewReceiptServiceClient(conn),
	}
}

// RegisterRoutes registers all foodfolio routes
func (h *FoodFolioHandler) RegisterRoutes(router *mux.Router) {
	// Category routes
	router.HandleFunc("/categories", h.ListCategories).Methods("GET")
	router.HandleFunc("/categories", h.CreateCategory).Methods("POST")
	router.HandleFunc("/categories/tree", h.GetCategoryTree).Methods("GET")
	router.HandleFunc("/categories/{id}", h.GetCategory).Methods("GET")
	router.HandleFunc("/categories/{id}", h.UpdateCategory).Methods("PUT")
	router.HandleFunc("/categories/{id}", h.DeleteCategory).Methods("DELETE")

	// Company routes
	router.HandleFunc("/companies", h.ListCompanies).Methods("GET")
	router.HandleFunc("/companies", h.CreateCompany).Methods("POST")
	router.HandleFunc("/companies/{id}", h.GetCompany).Methods("GET")
	router.HandleFunc("/companies/{id}", h.UpdateCompany).Methods("PUT")
	router.HandleFunc("/companies/{id}", h.DeleteCompany).Methods("DELETE")

	// Type routes
	router.HandleFunc("/types", h.ListTypes).Methods("GET")
	router.HandleFunc("/types", h.CreateType).Methods("POST")
	router.HandleFunc("/types/{id}", h.GetType).Methods("GET")
	router.HandleFunc("/types/{id}", h.UpdateType).Methods("PUT")
	router.HandleFunc("/types/{id}", h.DeleteType).Methods("DELETE")

	// Size routes
	router.HandleFunc("/sizes", h.ListSizes).Methods("GET")
	router.HandleFunc("/sizes", h.CreateSize).Methods("POST")
	router.HandleFunc("/sizes/{id}", h.GetSize).Methods("GET")
	router.HandleFunc("/sizes/{id}", h.UpdateSize).Methods("PUT")
	router.HandleFunc("/sizes/{id}", h.DeleteSize).Methods("DELETE")

	// Warehouse routes
	router.HandleFunc("/warehouses", h.ListWarehouses).Methods("GET")
	router.HandleFunc("/warehouses", h.CreateWarehouse).Methods("POST")
	router.HandleFunc("/warehouses/{id}", h.GetWarehouse).Methods("GET")
	router.HandleFunc("/warehouses/{id}", h.UpdateWarehouse).Methods("PUT")
	router.HandleFunc("/warehouses/{id}", h.DeleteWarehouse).Methods("DELETE")

	// Location routes
	router.HandleFunc("/locations", h.ListLocations).Methods("GET")
	router.HandleFunc("/locations", h.CreateLocation).Methods("POST")
	router.HandleFunc("/locations/tree", h.GetLocationTree).Methods("GET")
	router.HandleFunc("/locations/{id}", h.GetLocation).Methods("GET")
	router.HandleFunc("/locations/{id}", h.UpdateLocation).Methods("PUT")
	router.HandleFunc("/locations/{id}", h.DeleteLocation).Methods("DELETE")

	// Item routes
	router.HandleFunc("/items", h.ListItems).Methods("GET")
	router.HandleFunc("/items", h.CreateItem).Methods("POST")
	router.HandleFunc("/items/search", h.SearchItems).Methods("GET")
	router.HandleFunc("/items/{id}", h.GetItem).Methods("GET")
	router.HandleFunc("/items/{id}", h.UpdateItem).Methods("PUT")
	router.HandleFunc("/items/{id}", h.DeleteItem).Methods("DELETE")

	// ItemVariant routes
	router.HandleFunc("/item-variants", h.ListItemVariants).Methods("GET")
	router.HandleFunc("/item-variants", h.CreateItemVariant).Methods("POST")
	router.HandleFunc("/item-variants/low-stock", h.GetLowStockVariants).Methods("GET")
	router.HandleFunc("/item-variants/overstocked", h.GetOverstockedVariants).Methods("GET")
	router.HandleFunc("/item-variants/barcode", h.ScanBarcode).Methods("GET")
	router.HandleFunc("/item-variants/{id}", h.GetItemVariant).Methods("GET")
	router.HandleFunc("/item-variants/{id}", h.UpdateItemVariant).Methods("PUT")
	router.HandleFunc("/item-variants/{id}", h.DeleteItemVariant).Methods("DELETE")
	router.HandleFunc("/item-variants/{id}/stock", h.GetCurrentStock).Methods("GET")
	router.HandleFunc("/item-variants/{id}/with-details", h.GetItemWithVariants).Methods("GET")

	// ItemDetail routes
	router.HandleFunc("/item-details", h.ListItemDetails).Methods("GET")
	router.HandleFunc("/item-details", h.CreateItemDetail).Methods("POST")
	router.HandleFunc("/item-details/batch", h.BatchCreateItemDetails).Methods("POST")
	router.HandleFunc("/item-details/expiring", h.GetExpiringItems).Methods("GET")
	router.HandleFunc("/item-details/expired", h.GetExpiredItems).Methods("GET")
	router.HandleFunc("/item-details/deposit", h.GetItemsWithDeposit).Methods("GET")
	router.HandleFunc("/item-details/by-location/{location_id}", h.GetItemsByLocation).Methods("GET")
	router.HandleFunc("/item-details/{id}", h.GetItemDetail).Methods("GET")
	router.HandleFunc("/item-details/{id}", h.UpdateItemDetail).Methods("PUT")
	router.HandleFunc("/item-details/{id}", h.DeleteItemDetail).Methods("DELETE")
	router.HandleFunc("/item-details/{id}/open", h.OpenItem).Methods("POST")
	router.HandleFunc("/item-details/move", h.MoveItems).Methods("POST")

	// Shoppinglist routes
	router.HandleFunc("/shoppinglists", h.ListShoppinglists).Methods("GET")
	router.HandleFunc("/shoppinglists", h.CreateShoppinglist).Methods("POST")
	router.HandleFunc("/shoppinglists/generate-low-stock", h.GenerateFromLowStock).Methods("POST")
	router.HandleFunc("/shoppinglists/{id}", h.GetShoppinglist).Methods("GET")
	router.HandleFunc("/shoppinglists/{id}", h.UpdateShoppinglist).Methods("PUT")
	router.HandleFunc("/shoppinglists/{id}", h.DeleteShoppinglist).Methods("DELETE")
	router.HandleFunc("/shoppinglists/{id}/items", h.AddItemToShoppinglist).Methods("POST")
	router.HandleFunc("/shoppinglists/{id}/items/{item_id}", h.UpdateShoppinglistItem).Methods("PUT")
	router.HandleFunc("/shoppinglists/{id}/items/{item_id}", h.RemoveItemFromShoppinglist).Methods("DELETE")
	router.HandleFunc("/shoppinglists/{id}/items/{item_id}/purchase", h.MarkItemPurchased).Methods("POST")
	router.HandleFunc("/shoppinglists/{id}/purchase-all", h.MarkAllItemsPurchased).Methods("POST")
	router.HandleFunc("/shoppinglists/{id}/clear-purchased", h.ClearPurchasedItems).Methods("POST")

	// Receipt routes
	router.HandleFunc("/receipts", h.ListReceipts).Methods("GET")
	router.HandleFunc("/receipts", h.CreateReceipt).Methods("POST")
	router.HandleFunc("/receipts/upload", h.UploadReceipt).Methods("POST")
	router.HandleFunc("/receipts/statistics", h.GetReceiptStatistics).Methods("GET")
	router.HandleFunc("/receipts/{id}", h.GetReceipt).Methods("GET")
	router.HandleFunc("/receipts/{id}", h.UpdateReceipt).Methods("PUT")
	router.HandleFunc("/receipts/{id}", h.DeleteReceipt).Methods("DELETE")
	router.HandleFunc("/receipts/{id}/process", h.ProcessReceipt).Methods("POST")
	router.HandleFunc("/receipts/{id}/items", h.AddItemToReceipt).Methods("POST")
	router.HandleFunc("/receipts/{id}/items/{item_id}", h.UpdateReceiptItem).Methods("PUT")
	router.HandleFunc("/receipts/{id}/items/{item_id}/match", h.MatchReceiptItem).Methods("POST")
	router.HandleFunc("/receipts/{id}/auto-match", h.AutoMatchReceiptItems).Methods("POST")
	router.HandleFunc("/receipts/{id}/create-inventory", h.CreateInventoryFromReceipt).Methods("POST")
}

// Category handlers
func (h *FoodFolioHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 50
	}

	req := &pb.ListCategoriesRequest{
		Page:            int32(page),
		PageSize:        int32(pageSize),
		IncludeChildren: r.URL.Query().Get("include_children") == "true",
	}

	if parentID := r.URL.Query().Get("parent_id"); parentID != "" {
		req.ParentId = &parentID
	}

	resp, err := h.categoryClient.ListCategories(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to list categories: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.categoryClient.CreateCategory(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to create category: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := &pb.IdRequest{Id: id}
	resp, err := h.categoryClient.GetCategory(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get category: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req pb.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.Id = id

	resp, err := h.categoryClient.UpdateCategory(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to update category: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := &pb.IdRequest{Id: id}
	resp, err := h.categoryClient.DeleteCategory(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to delete category: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) GetCategoryTree(w http.ResponseWriter, r *http.Request) {
	maxDepth, _ := strconv.ParseInt(r.URL.Query().Get("max_depth"), 10, 32)

	req := &pb.GetCategoryTreeRequest{
		MaxDepth: int32(maxDepth),
	}

	if rootID := r.URL.Query().Get("root_id"); rootID != "" {
		req.RootId = &rootID
	}

	resp, err := h.categoryClient.GetCategoryTree(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get category tree: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Company handlers
func (h *FoodFolioHandler) ListCompanies(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 50
	}

	req := &pb.ListCompaniesRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	if search := r.URL.Query().Get("search"); search != "" {
		req.Search = &search
	}

	resp, err := h.companyClient.ListCompanies(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to list companies: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) CreateCompany(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateCompanyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.companyClient.CreateCompany(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to create company: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) GetCompany(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := &pb.IdRequest{Id: id}
	resp, err := h.companyClient.GetCompany(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get company: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) UpdateCompany(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req pb.UpdateCompanyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.Id = id

	resp, err := h.companyClient.UpdateCompany(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to update company: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) DeleteCompany(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := &pb.IdRequest{Id: id}
	resp, err := h.companyClient.DeleteCompany(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to delete company: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Due to file size, implementing remaining handlers in similar pattern...
// Types, Sizes, Warehouses, Locations follow same CRUD pattern

// Simplified implementations for remaining services (following same pattern)
func (h *FoodFolioHandler) ListTypes(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 50
	}

	req := &pb.ListTypesRequest{Page: int32(page), PageSize: int32(pageSize)}
	if search := r.URL.Query().Get("search"); search != "" {
		req.Search = &search
	}

	resp, err := h.typeClient.ListTypes(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to list types: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) CreateType(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.typeClient.CreateType(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to create type: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) GetType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.typeClient.GetType(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get type: "+err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) UpdateType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var req pb.UpdateTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.Id = vars["id"]

	resp, err := h.typeClient.UpdateType(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to update type: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) DeleteType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.typeClient.DeleteType(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to delete type: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Size handlers
func (h *FoodFolioHandler) ListSizes(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 50
	}

	req := &pb.ListSizesRequest{Page: int32(page), PageSize: int32(pageSize)}

	resp, err := h.sizeClient.ListSizes(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to list sizes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) CreateSize(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateSizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.sizeClient.CreateSize(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to create size: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) GetSize(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.sizeClient.GetSize(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get size: "+err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) UpdateSize(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var req pb.UpdateSizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.Id = vars["id"]

	resp, err := h.sizeClient.UpdateSize(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to update size: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) DeleteSize(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.sizeClient.DeleteSize(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to delete size: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
// Warehouse handlers
func (h *FoodFolioHandler) ListWarehouses(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 50
	}

	req := &pb.ListWarehousesRequest{Page: int32(page), PageSize: int32(pageSize)}
	if search := r.URL.Query().Get("search"); search != "" {
		req.Search = &search
	}

	resp, err := h.warehouseClient.ListWarehouses(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to list warehouses: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) CreateWarehouse(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateWarehouseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.warehouseClient.CreateWarehouse(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to create warehouse: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) GetWarehouse(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.warehouseClient.GetWarehouse(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get warehouse: "+err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) UpdateWarehouse(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var req pb.UpdateWarehouseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.Id = vars["id"]

	resp, err := h.warehouseClient.UpdateWarehouse(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to update warehouse: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) DeleteWarehouse(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.warehouseClient.DeleteWarehouse(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to delete warehouse: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
// Location handlers
func (h *FoodFolioHandler) ListLocations(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 50
	}

	req := &pb.ListLocationsRequest{
		Page:            int32(page),
		PageSize:        int32(pageSize),
		IncludeChildren: r.URL.Query().Get("include_children") == "true",
	}

	if parentID := r.URL.Query().Get("parent_id"); parentID != "" {
		req.ParentId = &parentID
	}

	resp, err := h.locationClient.ListLocations(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to list locations: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.locationClient.CreateLocation(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to create location: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) GetLocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.locationClient.GetLocation(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get location: "+err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var req pb.UpdateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.Id = vars["id"]

	resp, err := h.locationClient.UpdateLocation(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to update location: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) DeleteLocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.locationClient.DeleteLocation(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to delete location: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) GetLocationTree(w http.ResponseWriter, r *http.Request) {
	maxDepth, _ := strconv.ParseInt(r.URL.Query().Get("max_depth"), 10, 32)

	req := &pb.GetLocationTreeRequest{
		MaxDepth: int32(maxDepth),
	}

	if rootID := r.URL.Query().Get("root_id"); rootID != "" {
		req.RootId = &rootID
	}

	resp, err := h.locationClient.GetLocationTree(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get location tree: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
// Item handlers
func (h *FoodFolioHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 50
	}

	req := &pb.ListItemsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	if categoryID := r.URL.Query().Get("category_id"); categoryID != "" {
		req.CategoryId = &categoryID
	}
	if companyID := r.URL.Query().Get("company_id"); companyID != "" {
		req.CompanyId = &companyID
	}
	if search := r.URL.Query().Get("search"); search != "" {
		req.Search = &search
	}

	resp, err := h.itemClient.ListItems(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to list items: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.itemClient.CreateItem(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to create item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) GetItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.itemClient.GetItem(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get item: "+err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var req pb.UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.Id = vars["id"]

	resp, err := h.itemClient.UpdateItem(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to update item: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.itemClient.DeleteItem(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to delete item: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) SearchItems(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Missing search query parameter 'q'", http.StatusBadRequest)
		return
	}

	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 50
	}

	req := &pb.SearchItemsRequest{
		Query:    query,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	resp, err := h.itemClient.SearchItems(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to search items: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
// ItemVariant handlers
func (h *FoodFolioHandler) ListItemVariants(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 50
	}

	req := &pb.ListItemVariantsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	if itemID := r.URL.Query().Get("item_id"); itemID != "" {
		req.ItemId = &itemID
	}

	resp, err := h.itemVariantClient.ListItemVariants(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to list item variants: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) CreateItemVariant(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateItemVariantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.itemVariantClient.CreateItemVariant(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to create item variant: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) GetItemVariant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.itemVariantClient.GetItemVariant(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get item variant: "+err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) UpdateItemVariant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var req pb.UpdateItemVariantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.Id = vars["id"]

	resp, err := h.itemVariantClient.UpdateItemVariant(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to update item variant: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) DeleteItemVariant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.itemVariantClient.DeleteItemVariant(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to delete item variant: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) GetCurrentStock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.GetCurrentStockRequest{Id: vars["id"]}
	resp, err := h.itemVariantClient.GetCurrentStock(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get current stock: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) GetLowStockVariants(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 50
	}

	req := &pb.GetLowStockVariantsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	}
	resp, err := h.itemVariantClient.GetLowStockVariants(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get low stock variants: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) GetOverstockedVariants(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 50
	}

	req := &pb.GetOverstockedVariantsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	}
	resp, err := h.itemVariantClient.GetOverstockedVariants(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get overstocked variants: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) GetItemWithVariants(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.GetItemWithVariantsRequest{
		Id:             vars["id"],
		IncludeDetails: r.URL.Query().Get("include_details") == "true",
	}
	resp, err := h.itemVariantClient.GetItemWithVariants(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get item with variants: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) ScanBarcode(w http.ResponseWriter, r *http.Request) {
	barcode := r.URL.Query().Get("barcode")
	if barcode == "" {
		http.Error(w, "Missing barcode parameter", http.StatusBadRequest)
		return
	}

	req := &pb.ScanBarcodeRequest{Barcode: barcode}
	resp, err := h.itemVariantClient.ScanBarcode(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to scan barcode: "+err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
// ItemDetail handlers
func (h *FoodFolioHandler) ListItemDetails(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 50
	}

	req := &pb.ListItemDetailsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	if variantID := r.URL.Query().Get("variant_id"); variantID != "" {
		req.ItemVariantId = &variantID
	}
	if locationID := r.URL.Query().Get("location_id"); locationID != "" {
		req.LocationId = &locationID
	}

	resp, err := h.itemDetailClient.ListItemDetails(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to list item details: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) CreateItemDetail(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateItemDetailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.itemDetailClient.CreateItemDetail(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to create item detail: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) BatchCreateItemDetails(w http.ResponseWriter, r *http.Request) {
	var req pb.BatchCreateItemDetailsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.itemDetailClient.BatchCreateItemDetails(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to batch create item details: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) GetItemDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.itemDetailClient.GetItemDetail(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get item detail: "+err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) UpdateItemDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var req pb.UpdateItemDetailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.Id = vars["id"]

	resp, err := h.itemDetailClient.UpdateItemDetail(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to update item detail: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) DeleteItemDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.itemDetailClient.DeleteItemDetail(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to delete item detail: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) OpenItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.OpenItemRequest{Id: vars["id"]}
	resp, err := h.itemDetailClient.OpenItem(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to open item: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) MoveItems(w http.ResponseWriter, r *http.Request) {
	var req pb.MoveItemsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.itemDetailClient.MoveItems(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to move items: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) GetExpiringItems(w http.ResponseWriter, r *http.Request) {
	days, _ := strconv.ParseInt(r.URL.Query().Get("days"), 10, 32)
	if days == 0 {
		days = 7
	}
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 50
	}

	req := &pb.GetExpiringItemsRequest{
		Days:     int32(days),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}
	resp, err := h.itemDetailClient.GetExpiringItems(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get expiring items: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) GetExpiredItems(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 50
	}

	req := &pb.GetExpiredItemsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	}
	resp, err := h.itemDetailClient.GetExpiredItems(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get expired items: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) GetItemsWithDeposit(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 50
	}

	req := &pb.GetItemsWithDepositRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	}
	resp, err := h.itemDetailClient.GetItemsWithDeposit(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get items with deposit: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) GetItemsByLocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 50
	}

	req := &pb.GetItemsByLocationRequest{
		LocationId:      vars["location_id"],
		IncludeChildren: r.URL.Query().Get("include_children") == "true",
		Page:            int32(page),
		PageSize:        int32(pageSize),
	}
	resp, err := h.itemDetailClient.GetItemsByLocation(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get items by location: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
// Shoppinglist handlers
func (h *FoodFolioHandler) ListShoppinglists(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 50
	}

	req := &pb.ListShoppinglistsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	resp, err := h.shoppinglistClient.ListShoppinglists(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to list shoppinglists: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) CreateShoppinglist(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateShoppinglistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.shoppinglistClient.CreateShoppinglist(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to create shoppinglist: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) GetShoppinglist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.shoppinglistClient.GetShoppinglist(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get shoppinglist: "+err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) UpdateShoppinglist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var req pb.UpdateShoppinglistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.Id = vars["id"]

	resp, err := h.shoppinglistClient.UpdateShoppinglist(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to update shoppinglist: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) DeleteShoppinglist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.shoppinglistClient.DeleteShoppinglist(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to delete shoppinglist: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) AddItemToShoppinglist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var req pb.AddItemToShoppinglistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.ShoppinglistId = vars["id"]

	resp, err := h.shoppinglistClient.AddItemToShoppinglist(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to add item to shoppinglist: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) RemoveItemFromShoppinglist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.RemoveItemFromShoppinglistRequest{
		ShoppinglistId: vars["id"],
		ItemId:         vars["item_id"],
	}

	resp, err := h.shoppinglistClient.RemoveItemFromShoppinglist(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to remove item from shoppinglist: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) UpdateShoppinglistItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var req pb.UpdateShoppinglistItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.Id = vars["item_id"]

	resp, err := h.shoppinglistClient.UpdateShoppinglistItem(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to update shoppinglist item: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) MarkItemPurchased(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.MarkItemPurchasedRequest{
		Id: vars["item_id"],
	}

	resp, err := h.shoppinglistClient.MarkItemPurchased(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to mark item purchased: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) MarkAllItemsPurchased(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.MarkAllItemsPurchasedRequest{ShoppinglistId: vars["id"]}

	resp, err := h.shoppinglistClient.MarkAllItemsPurchased(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to mark all items purchased: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) ClearPurchasedItems(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.ClearPurchasedItemsRequest{ShoppinglistId: vars["id"]}

	resp, err := h.shoppinglistClient.ClearPurchasedItems(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to clear purchased items: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) GenerateFromLowStock(w http.ResponseWriter, r *http.Request) {
	var req pb.GenerateFromLowStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.shoppinglistClient.GenerateFromLowStock(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to generate from low stock: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
// Receipt handlers
func (h *FoodFolioHandler) ListReceipts(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 50
	}

	req := &pb.ListReceiptsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	resp, err := h.receiptClient.ListReceipts(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to list receipts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) CreateReceipt(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateReceiptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.receiptClient.CreateReceipt(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to create receipt: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) GetReceipt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.receiptClient.GetReceipt(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get receipt: "+err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) UpdateReceipt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var req pb.UpdateReceiptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.Id = vars["id"]

	resp, err := h.receiptClient.UpdateReceipt(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to update receipt: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) DeleteReceipt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.receiptClient.DeleteReceipt(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to delete receipt: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) UploadReceipt(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (max 32MB)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get warehouse_id from form
	warehouseID := r.FormValue("warehouse_id")
	if warehouseID == "" {
		http.Error(w, "warehouse_id is required", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create gRPC request
	req := &pb.UploadReceiptRequest{
		WarehouseId: warehouseID,
		ImageData:   fileData,
		ContentType: header.Header.Get("Content-Type"),
	}

	resp, err := h.receiptClient.UploadReceipt(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to upload receipt: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) ProcessReceipt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.ProcessReceiptRequest{
		ReceiptId: vars["id"],
		AutoMatch: r.URL.Query().Get("auto_match") == "true",
	}

	resp, err := h.receiptClient.ProcessReceipt(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to process receipt: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) AddItemToReceipt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var req pb.AddItemToReceiptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.ReceiptId = vars["id"]

	resp, err := h.receiptClient.AddItemToReceipt(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to add item to receipt: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) UpdateReceiptItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var req pb.UpdateReceiptItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.Id = vars["item_id"]

	resp, err := h.receiptClient.UpdateReceiptItem(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to update receipt item: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) MatchReceiptItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var reqBody struct {
		ItemVariantId string `json:"item_variant_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	req := &pb.MatchReceiptItemRequest{
		ReceiptItemId: vars["item_id"],
		ItemVariantId: reqBody.ItemVariantId,
	}

	resp, err := h.receiptClient.MatchReceiptItem(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to match receipt item: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) AutoMatchReceiptItems(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.AutoMatchReceiptItemsRequest{
		ReceiptId: vars["id"],
	}

	// Optional similarity threshold from query parameter
	if threshold := r.URL.Query().Get("similarity_threshold"); threshold != "" {
		if val, err := strconv.ParseFloat(threshold, 64); err == nil {
			req.SimilarityThreshold = val
		}
	}

	resp, err := h.receiptClient.AutoMatchReceiptItems(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to auto-match receipt items: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) CreateInventoryFromReceipt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var req pb.CreateInventoryFromReceiptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.ReceiptId = vars["id"]

	resp, err := h.receiptClient.CreateInventoryFromReceipt(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to create inventory from receipt: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *FoodFolioHandler) GetReceiptStatistics(w http.ResponseWriter, r *http.Request) {
	req := &pb.GetReceiptStatisticsRequest{}

	// Optional filters from query parameters
	if warehouseID := r.URL.Query().Get("warehouse_id"); warehouseID != "" {
		req.WarehouseId = &warehouseID
	}

	resp, err := h.receiptClient.GetReceiptStatistics(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get receipt statistics: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
