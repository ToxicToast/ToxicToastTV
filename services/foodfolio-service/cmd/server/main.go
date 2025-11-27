package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/gorm"

	"github.com/toxictoast/toxictoastgo/shared/auth"
	"github.com/toxictoast/toxictoastgo/shared/database"
	sharedgrpc "github.com/toxictoast/toxictoastgo/shared/grpc"
	"github.com/toxictoast/toxictoastgo/shared/kafka"
	"github.com/toxictoast/toxictoastgo/shared/logger"

	"toxictoast/services/foodfolio-service/pkg/config"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"

	// Repository layer
	"toxictoast/services/foodfolio-service/internal/repository/entity"
	repoImpl "toxictoast/services/foodfolio-service/internal/repository/impl"

	// CQRS layer
	"toxictoast/services/foodfolio-service/internal/command"
	"toxictoast/services/foodfolio-service/internal/query"

	// Scheduler layer
	"toxictoast/services/foodfolio-service/internal/scheduler"

	// Handler layer
	grpcHandler "toxictoast/services/foodfolio-service/internal/handler/grpc"

	// Proto definitions
	pb "toxictoast/services/foodfolio-service/api/proto"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	db        *gorm.DB
)

func main() {
	// Load .env file (ignore error in production where env vars are set directly)
	_ = godotenv.Load()

	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger.Init()

	log.Printf("Starting Foodfolio Service v%s (built: %s)", Version, BuildTime)
	log.Printf("Environment: %s", cfg.Environment)

	// Connect to database with retry
	var err error
	db, err = database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Printf("Database connected successfully")

	// Run auto-migrations
	dbEntities := []interface{}{
		&entity.CategoryEntity{},
		&entity.CompanyEntity{},
		&entity.TypeEntity{},
		&entity.SizeEntity{},
		&entity.WarehouseEntity{},
		&entity.LocationEntity{},
		&entity.ItemEntity{},
		&entity.ItemVariantEntity{},
		&entity.ItemDetailEntity{},
		&entity.ShoppinglistEntity{},
		&entity.ShoppinglistItemEntity{},
		&entity.ReceiptEntity{},
		&entity.ReceiptItemEntity{},
	}
	if err := database.AutoMigrate(db, dbEntities...); err != nil {
		log.Fatalf("Auto-migration failed: %v", err)
	}
	log.Printf("Database schema is up to date (using foodfolio_ table prefix)")

	// Initialize Kafka producer
	kafkaProducer, err := kafka.NewProducer(cfg.Kafka.Brokers)
	if err != nil {
		log.Printf("Warning: Failed to initialize Kafka producer: %v", err)
		log.Printf("Service will continue without event publishing")
		kafkaProducer = nil
	} else {
		log.Printf("Kafka producer connected successfully")
		defer kafkaProducer.Close()
	}

	// Initialize Keycloak auth
	var keycloakAuth *auth.KeycloakAuth
	if cfg.AuthEnabled {
		keycloakAuth, err = auth.NewKeycloakAuth(&cfg.Keycloak)
		if err != nil {
			log.Printf("Warning: Failed to initialize Keycloak auth: %v", err)
			log.Printf("Service will continue without authentication")
			keycloakAuth = nil
		} else {
			log.Printf("Keycloak authentication initialized")
		}
	} else {
		log.Printf("Authentication is DISABLED (AUTH_ENABLED=false)")
	}

	// Initialize repositories
	log.Println("Initializing repositories...")
	categoryRepo := repoImpl.NewCategoryRepository(db)
	companyRepo := repoImpl.NewCompanyRepository(db)
	typeRepo := repoImpl.NewTypeRepository(db)
	sizeRepo := repoImpl.NewSizeRepository(db)
	warehouseRepo := repoImpl.NewWarehouseRepository(db)
	locationRepo := repoImpl.NewLocationRepository(db)
	itemRepo := repoImpl.NewItemRepository(db)
	itemVariantRepo := repoImpl.NewItemVariantRepository(db)
	itemDetailRepo := repoImpl.NewItemDetailRepository(db)
	shoppinglistRepo := repoImpl.NewShoppinglistRepository(db)
	receiptRepo := repoImpl.NewReceiptRepository(db)
	log.Println("Repositories initialized")

	// Initialize Command Bus
	log.Println("Initializing Command Bus...")
	commandBus := cqrs.NewCommandBus()

	// Register Category Command Handlers
	commandBus.RegisterHandler("create_category", command.NewCreateCategoryHandler(categoryRepo))
	commandBus.RegisterHandler("update_category", command.NewUpdateCategoryHandler(categoryRepo))
	commandBus.RegisterHandler("delete_category", command.NewDeleteCategoryHandler(categoryRepo))

	// Register Company Command Handlers
	commandBus.RegisterHandler("create_company", command.NewCreateCompanyHandler(companyRepo))
	commandBus.RegisterHandler("update_company", command.NewUpdateCompanyHandler(companyRepo))
	commandBus.RegisterHandler("delete_company", command.NewDeleteCompanyHandler(companyRepo))

	// Register Type Command Handlers
	commandBus.RegisterHandler("create_type", command.NewCreateTypeHandler(typeRepo))
	commandBus.RegisterHandler("update_type", command.NewUpdateTypeHandler(typeRepo))
	commandBus.RegisterHandler("delete_type", command.NewDeleteTypeHandler(typeRepo))

	// Register Size Command Handlers
	commandBus.RegisterHandler("create_size", command.NewCreateSizeHandler(sizeRepo))
	commandBus.RegisterHandler("update_size", command.NewUpdateSizeHandler(sizeRepo))
	commandBus.RegisterHandler("delete_size", command.NewDeleteSizeHandler(sizeRepo))

	// Register Warehouse Command Handlers
	commandBus.RegisterHandler("create_warehouse", command.NewCreateWarehouseHandler(warehouseRepo))
	commandBus.RegisterHandler("update_warehouse", command.NewUpdateWarehouseHandler(warehouseRepo))
	commandBus.RegisterHandler("delete_warehouse", command.NewDeleteWarehouseHandler(warehouseRepo))

	// Register Location Command Handlers
	commandBus.RegisterHandler("create_location", command.NewCreateLocationHandler(locationRepo))
	commandBus.RegisterHandler("update_location", command.NewUpdateLocationHandler(locationRepo))
	commandBus.RegisterHandler("delete_location", command.NewDeleteLocationHandler(locationRepo))

	// Register Item Command Handlers
	commandBus.RegisterHandler("create_item", command.NewCreateItemHandler(itemRepo, categoryRepo, companyRepo, typeRepo, kafkaProducer))
	commandBus.RegisterHandler("update_item", command.NewUpdateItemHandler(itemRepo, categoryRepo, companyRepo, typeRepo, kafkaProducer))
	commandBus.RegisterHandler("delete_item", command.NewDeleteItemHandler(itemRepo, kafkaProducer))

	// Register ItemVariant Command Handlers
	commandBus.RegisterHandler("create_item_variant", command.NewCreateItemVariantHandler(itemVariantRepo, itemRepo, sizeRepo, kafkaProducer))
	commandBus.RegisterHandler("update_item_variant", command.NewUpdateItemVariantHandler(itemVariantRepo, kafkaProducer))
	commandBus.RegisterHandler("delete_item_variant", command.NewDeleteItemVariantHandler(itemVariantRepo, kafkaProducer))

	// Register ItemDetail Command Handlers
	commandBus.RegisterHandler("create_item_detail", command.NewCreateItemDetailHandler(itemDetailRepo, itemVariantRepo, warehouseRepo, locationRepo, kafkaProducer))
	commandBus.RegisterHandler("batch_create_item_details", command.NewBatchCreateItemDetailsHandler(itemDetailRepo, itemVariantRepo, warehouseRepo, locationRepo, kafkaProducer))
	commandBus.RegisterHandler("open_item", command.NewOpenItemHandler(itemDetailRepo, kafkaProducer))
	commandBus.RegisterHandler("move_items", command.NewMoveItemsHandler(itemDetailRepo, locationRepo, kafkaProducer))
	commandBus.RegisterHandler("update_item_detail", command.NewUpdateItemDetailHandler(itemDetailRepo, locationRepo, kafkaProducer))
	commandBus.RegisterHandler("delete_item_detail", command.NewDeleteItemDetailHandler(itemDetailRepo, kafkaProducer))

	// Register Shoppinglist Command Handlers
	commandBus.RegisterHandler("create_shoppinglist", command.NewCreateShoppinglistHandler(shoppinglistRepo, kafkaProducer))
	commandBus.RegisterHandler("update_shoppinglist", command.NewUpdateShoppinglistHandler(shoppinglistRepo, kafkaProducer))
	commandBus.RegisterHandler("delete_shoppinglist", command.NewDeleteShoppinglistHandler(shoppinglistRepo, kafkaProducer))
	commandBus.RegisterHandler("generate_from_low_stock", command.NewGenerateFromLowStockHandler(shoppinglistRepo, itemVariantRepo))
	commandBus.RegisterHandler("add_item_to_shoppinglist", command.NewAddItemToShoppinglistHandler(shoppinglistRepo, itemVariantRepo, kafkaProducer))
	commandBus.RegisterHandler("remove_item_from_shoppinglist", command.NewRemoveItemFromShoppinglistHandler(shoppinglistRepo, kafkaProducer))
	commandBus.RegisterHandler("update_shoppinglist_item", command.NewUpdateShoppinglistItemHandler(shoppinglistRepo))
	commandBus.RegisterHandler("mark_item_purchased", command.NewMarkItemPurchasedHandler(shoppinglistRepo, kafkaProducer))
	commandBus.RegisterHandler("mark_all_items_purchased", command.NewMarkAllItemsPurchasedHandler(shoppinglistRepo))
	commandBus.RegisterHandler("clear_purchased_items", command.NewClearPurchasedItemsHandler(shoppinglistRepo))

	// Register Receipt Command Handlers
	commandBus.RegisterHandler("create_receipt", command.NewCreateReceiptHandler(receiptRepo, warehouseRepo, kafkaProducer))
	commandBus.RegisterHandler("update_receipt", command.NewUpdateReceiptHandler(receiptRepo, warehouseRepo))
	commandBus.RegisterHandler("update_receipt_ocr_data", command.NewUpdateReceiptOCRDataHandler(receiptRepo))
	commandBus.RegisterHandler("delete_receipt", command.NewDeleteReceiptHandler(receiptRepo, kafkaProducer))
	commandBus.RegisterHandler("add_item_to_receipt", command.NewAddItemToReceiptHandler(receiptRepo, itemVariantRepo))
	commandBus.RegisterHandler("update_receipt_item", command.NewUpdateReceiptItemHandler(receiptRepo))
	commandBus.RegisterHandler("match_receipt_item", command.NewMatchReceiptItemHandler(receiptRepo, itemVariantRepo))
	commandBus.RegisterHandler("auto_match_receipt_items", command.NewAutoMatchReceiptItemsHandler(receiptRepo, itemVariantRepo))
	commandBus.RegisterHandler("create_inventory_from_receipt", command.NewCreateInventoryFromReceiptHandler(receiptRepo, itemDetailRepo, locationRepo))

	log.Printf("Command Bus initialized with 62 command handlers")

	// Initialize Query Bus
	log.Println("Initializing Query Bus...")
	queryBus := cqrs.NewQueryBus()

	// Register Category Query Handlers
	queryBus.RegisterHandler("get_category_by_id", query.NewGetCategoryByIDHandler(categoryRepo))
	queryBus.RegisterHandler("get_category_by_slug", query.NewGetCategoryBySlugHandler(categoryRepo))
	queryBus.RegisterHandler("list_categories", query.NewListCategoriesHandler(categoryRepo))
	queryBus.RegisterHandler("get_category_tree", query.NewGetCategoryTreeHandler(categoryRepo))

	// Register Company Query Handlers
	queryBus.RegisterHandler("get_company_by_id", query.NewGetCompanyByIDHandler(companyRepo))
	queryBus.RegisterHandler("get_company_by_slug", query.NewGetCompanyBySlugHandler(companyRepo))
	queryBus.RegisterHandler("list_companies", query.NewListCompaniesHandler(companyRepo))

	// Register Type Query Handlers
	queryBus.RegisterHandler("get_type_by_id", query.NewGetTypeByIDHandler(typeRepo))
	queryBus.RegisterHandler("get_type_by_slug", query.NewGetTypeBySlugHandler(typeRepo))
	queryBus.RegisterHandler("list_types", query.NewListTypesHandler(typeRepo))

	// Register Size Query Handlers
	queryBus.RegisterHandler("get_size_by_id", query.NewGetSizeByIDHandler(sizeRepo))
	queryBus.RegisterHandler("list_sizes", query.NewListSizesHandler(sizeRepo))

	// Register Warehouse Query Handlers
	queryBus.RegisterHandler("get_warehouse_by_id", query.NewGetWarehouseByIDHandler(warehouseRepo))
	queryBus.RegisterHandler("get_warehouse_by_slug", query.NewGetWarehouseBySlugHandler(warehouseRepo))
	queryBus.RegisterHandler("list_warehouses", query.NewListWarehousesHandler(warehouseRepo))

	// Register Location Query Handlers
	queryBus.RegisterHandler("get_location_by_id", query.NewGetLocationByIDHandler(locationRepo))
	queryBus.RegisterHandler("get_location_by_slug", query.NewGetLocationBySlugHandler(locationRepo))
	queryBus.RegisterHandler("list_locations", query.NewListLocationsHandler(locationRepo))
	queryBus.RegisterHandler("get_location_tree", query.NewGetLocationTreeHandler(locationRepo))

	// Register Item Query Handlers
	queryBus.RegisterHandler("get_item_by_id", query.NewGetItemByIDHandler(itemRepo))
	queryBus.RegisterHandler("get_item_by_slug", query.NewGetItemBySlugHandler(itemRepo))
	queryBus.RegisterHandler("get_item_with_variants", query.NewGetItemWithVariantsHandler(itemRepo))
	queryBus.RegisterHandler("list_items", query.NewListItemsHandler(itemRepo))
	queryBus.RegisterHandler("search_items", query.NewSearchItemsHandler(itemRepo))

	// Register ItemVariant Query Handlers
	queryBus.RegisterHandler("get_item_variant_by_id", query.NewGetItemVariantByIDHandler(itemVariantRepo))
	queryBus.RegisterHandler("get_item_variant_by_barcode", query.NewGetItemVariantByBarcodeHandler(itemVariantRepo))
	queryBus.RegisterHandler("list_item_variants", query.NewListItemVariantsHandler(itemVariantRepo))
	queryBus.RegisterHandler("get_item_variants_by_item", query.NewGetItemVariantsByItemHandler(itemVariantRepo, itemRepo))
	queryBus.RegisterHandler("get_low_stock_variants", query.NewGetLowStockVariantsHandler(itemVariantRepo))
	queryBus.RegisterHandler("get_overstocked_variants", query.NewGetOverstockedVariantsHandler(itemVariantRepo))
	queryBus.RegisterHandler("get_current_stock", query.NewGetCurrentStockHandler(itemVariantRepo))

	// Register ItemDetail Query Handlers
	queryBus.RegisterHandler("get_item_detail_by_id", query.NewGetItemDetailByIDHandler(itemDetailRepo))
	queryBus.RegisterHandler("list_item_details", query.NewListItemDetailsHandler(itemDetailRepo))
	queryBus.RegisterHandler("get_item_details_by_variant", query.NewGetItemDetailsByVariantHandler(itemDetailRepo, itemVariantRepo))
	queryBus.RegisterHandler("get_item_details_by_location", query.NewGetItemDetailsByLocationHandler(itemDetailRepo, locationRepo))
	queryBus.RegisterHandler("get_expiring_items", query.NewGetExpiringItemsHandler(itemDetailRepo))
	queryBus.RegisterHandler("get_expired_items", query.NewGetExpiredItemsHandler(itemDetailRepo))
	queryBus.RegisterHandler("get_items_with_deposit", query.NewGetItemsWithDepositHandler(itemDetailRepo))

	// Register Shoppinglist Query Handlers
	queryBus.RegisterHandler("get_shoppinglist_by_id", query.NewGetShoppinglistByIDHandler(shoppinglistRepo))
	queryBus.RegisterHandler("list_shoppinglists", query.NewListShoppinglistsHandler(shoppinglistRepo))

	// Register Receipt Query Handlers
	queryBus.RegisterHandler("get_receipt_by_id", query.NewGetReceiptByIDHandler(receiptRepo))
	queryBus.RegisterHandler("list_receipts", query.NewListReceiptsHandler(receiptRepo))
	queryBus.RegisterHandler("get_unmatched_items", query.NewGetUnmatchedItemsHandler(receiptRepo))
	queryBus.RegisterHandler("get_statistics", query.NewGetStatisticsHandler(receiptRepo))

	log.Printf("Query Bus initialized with 50 query handlers")
	log.Printf("CQRS initialization complete: 112 total handlers")

	// Initialize background job schedulers
	itemExpirationScheduler := scheduler.NewItemExpirationScheduler(
		kafkaProducer,
		itemDetailRepo,
		cfg.ItemExpirationInterval,
		cfg.ItemExpirationEnabled,
	)
	stockLevelScheduler := scheduler.NewStockLevelScheduler(
		kafkaProducer,
		itemVariantRepo,
		cfg.StockLevelInterval,
		cfg.StockLevelEnabled,
	)

	// Start background jobs
	itemExpirationScheduler.Start()
	stockLevelScheduler.Start()
	log.Printf("Background jobs initialized")

	// Initialize gRPC handlers
	log.Println("Initializing gRPC handlers...")
	categoryHandler := grpcHandler.NewCategoryHandler(commandBus, queryBus)
	companyHandler := grpcHandler.NewCompanyHandler(commandBus, queryBus)
	typeHandler := grpcHandler.NewTypeHandler(commandBus, queryBus)
	sizeHandler := grpcHandler.NewSizeHandler(commandBus, queryBus)
	warehouseHandler := grpcHandler.NewWarehouseHandler(commandBus, queryBus)
	locationHandler := grpcHandler.NewLocationHandler(commandBus, queryBus)
	itemHandler := grpcHandler.NewItemHandler(commandBus, queryBus)
	itemVariantHandler := grpcHandler.NewItemVariantHandler(commandBus, queryBus)
	itemDetailHandler := grpcHandler.NewItemDetailHandler(commandBus, queryBus)
	shoppinglistHandler := grpcHandler.NewShoppinglistHandler(commandBus, queryBus)
	receiptHandler := grpcHandler.NewReceiptHandler(commandBus, queryBus)
	log.Println("gRPC handlers initialized")

	// Setup gRPC server
	grpcServer := setupGRPCServer(
		cfg,
		keycloakAuth,
		categoryHandler,
		companyHandler,
		typeHandler,
		sizeHandler,
		warehouseHandler,
		locationHandler,
		itemHandler,
		itemVariantHandler,
		itemDetailHandler,
		shoppinglistHandler,
		receiptHandler,
	)

	// Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
		if err != nil {
			log.Fatalf("Failed to listen on gRPC port: %v", err)
		}
		log.Printf("gRPC server starting on port %s", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	// Setup HTTP server for health checks
	httpServer := setupHTTPServer(cfg)

	// Start HTTP server
	go func() {
		log.Printf("HTTP server starting on port %s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop background jobs
	itemExpirationScheduler.Stop()
	stockLevelScheduler.Stop()
	log.Println("Background jobs stopped")

	// Shutdown HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Shutdown gRPC server
	grpcServer.GracefulStop()

	log.Println("Servers stopped")
}

func setupGRPCServer(
	cfg *config.Config,
	keycloakAuth *auth.KeycloakAuth,
	categoryHandler *grpcHandler.CategoryHandler,
	companyHandler *grpcHandler.CompanyHandler,
	typeHandler *grpcHandler.TypeHandler,
	sizeHandler *grpcHandler.SizeHandler,
	warehouseHandler *grpcHandler.WarehouseHandler,
	locationHandler *grpcHandler.LocationHandler,
	itemHandler *grpcHandler.ItemHandler,
	itemVariantHandler *grpcHandler.ItemVariantHandler,
	itemDetailHandler *grpcHandler.ItemDetailHandler,
	shoppinglistHandler *grpcHandler.ShoppinglistHandler,
	receiptHandler *grpcHandler.ReceiptHandler,
) *grpc.Server {
	// Setup interceptors - always add auth interceptor to extract user from metadata
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		sharedgrpc.AuthInterceptor,
	}
	streamInterceptors := []grpc.StreamServerInterceptor{
		sharedgrpc.StreamAuthInterceptor,
	}

	// Add Keycloak interceptor if enabled
	if keycloakAuth != nil {
		unaryInterceptors = append(unaryInterceptors, keycloakAuth.UnaryInterceptor())
		streamInterceptors = append(streamInterceptors, keycloakAuth.StreamInterceptor())
	}

	// Create server options with chained interceptors
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	}

	// Create gRPC server with interceptors
	server := grpc.NewServer(opts...)

	// Register all services
	log.Println("Registering gRPC services...")
	pb.RegisterCategoryServiceServer(server, categoryHandler)
	pb.RegisterCompanyServiceServer(server, companyHandler)
	pb.RegisterTypeServiceServer(server, typeHandler)
	pb.RegisterSizeServiceServer(server, sizeHandler)
	pb.RegisterWarehouseServiceServer(server, warehouseHandler)
	pb.RegisterLocationServiceServer(server, locationHandler)
	pb.RegisterItemServiceServer(server, itemHandler)
	pb.RegisterItemVariantServiceServer(server, itemVariantHandler)
	pb.RegisterItemDetailServiceServer(server, itemDetailHandler)
	pb.RegisterShoppinglistServiceServer(server, shoppinglistHandler)
	pb.RegisterReceiptServiceServer(server, receiptHandler)
	log.Println("All gRPC services registered")

	// Enable reflection for tools like grpcurl
	reflection.Register(server)

	return server
}

func setupHTTPServer(cfg *config.Config) *http.Server {
	router := mux.NewRouter()

	// Health check endpoints
	router.HandleFunc("/health", healthCheckHandler).Methods("GET")
	router.HandleFunc("/health/ready", readinessHandler).Methods("GET")
	router.HandleFunc("/health/live", livenessHandler).Methods("GET")

	return &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Healthy"))
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	if err := database.CheckHealth(db); err != nil {
		http.Error(w, "Database not ready", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ready"))
}

func livenessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Alive"))
}
