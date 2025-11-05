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
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/gorm"

	"github.com/toxictoast/toxictoastgo/shared/auth"
	"github.com/toxictoast/toxictoastgo/shared/database"
	"github.com/toxictoast/toxictoastgo/shared/kafka"
	"github.com/toxictoast/toxictoastgo/shared/logger"

	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/pkg/config"

	// Repository layer
	repoImpl "toxictoast/services/foodfolio-service/internal/repository/impl"

	// Use case layer
	"toxictoast/services/foodfolio-service/internal/usecase"

	// Handler layer
	grpcHandler "toxictoast/services/foodfolio-service/internal/handler/grpc"

	// Proto definitions
	pb "toxictoast/services/foodfolio-service/api/proto/foodfolio/v1"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	db        *gorm.DB
)

func main() {
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
	entities := []interface{}{
		&domain.Category{},
		&domain.Company{},
		&domain.Type{},
		&domain.Size{},
		&domain.Warehouse{},
		&domain.Location{},
		&domain.Item{},
		&domain.ItemVariant{},
		&domain.ItemDetail{},
		&domain.Shoppinglist{},
		&domain.ShoppinglistItem{},
		&domain.Receipt{},
		&domain.ReceiptItem{},
	}
	if err := database.AutoMigrate(db, entities...); err != nil {
		log.Fatalf("Auto-migration failed: %v", err)
	}
	log.Printf("Database schema is up to date")

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

	// Initialize use cases
	log.Println("Initializing use cases...")
	categoryUC := usecase.NewCategoryUseCase(categoryRepo)
	companyUC := usecase.NewCompanyUseCase(companyRepo)
	typeUC := usecase.NewTypeUseCase(typeRepo)
	sizeUC := usecase.NewSizeUseCase(sizeRepo)
	warehouseUC := usecase.NewWarehouseUseCase(warehouseRepo)
	locationUC := usecase.NewLocationUseCase(locationRepo)
	itemUC := usecase.NewItemUseCase(itemRepo, categoryRepo, companyRepo, typeRepo)
	itemVariantUC := usecase.NewItemVariantUseCase(itemVariantRepo, itemRepo, sizeRepo)
	itemDetailUC := usecase.NewItemDetailUseCase(itemDetailRepo, itemVariantRepo, warehouseRepo, locationRepo)
	shoppinglistUC := usecase.NewShoppinglistUseCase(shoppinglistRepo, itemVariantRepo)
	receiptUC := usecase.NewReceiptUseCase(receiptRepo, warehouseRepo, itemVariantRepo, itemDetailRepo, locationRepo)
	log.Println("Use cases initialized")

	// Initialize gRPC handlers
	log.Println("Initializing gRPC handlers...")
	categoryHandler := grpcHandler.NewCategoryHandler(categoryUC)
	companyHandler := grpcHandler.NewCompanyHandler(companyUC)
	typeHandler := grpcHandler.NewTypeHandler(typeUC)
	sizeHandler := grpcHandler.NewSizeHandler(sizeUC)
	warehouseHandler := grpcHandler.NewWarehouseHandler(warehouseUC)
	locationHandler := grpcHandler.NewLocationHandler(locationUC)
	itemHandler := grpcHandler.NewItemHandler(itemUC)
	itemVariantHandler := grpcHandler.NewItemVariantHandler(itemVariantUC)
	itemDetailHandler := grpcHandler.NewItemDetailHandler(itemDetailUC)
	shoppinglistHandler := grpcHandler.NewShoppinglistHandler(shoppinglistUC)
	receiptHandler := grpcHandler.NewReceiptHandler(receiptUC)
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
	// Setup interceptors
	var opts []grpc.ServerOption

	if keycloakAuth != nil {
		opts = append(opts,
			grpc.UnaryInterceptor(keycloakAuth.UnaryInterceptor()),
			grpc.StreamInterceptor(keycloakAuth.StreamInterceptor()),
		)
	}

	// Create gRPC server
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
