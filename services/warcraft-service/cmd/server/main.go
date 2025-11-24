package main

import (
	"context"
	"encoding/json"
	"fmt"
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

	"github.com/toxictoast/toxictoastgo/shared/database"
	sharedgrpc "github.com/toxictoast/toxictoastgo/shared/grpc"
	"github.com/toxictoast/toxictoastgo/shared/kafka"
	"github.com/toxictoast/toxictoastgo/shared/logger"

	pb "toxictoast/services/warcraft-service/api/proto"
	grpcHandler "toxictoast/services/warcraft-service/internal/handler/grpc"
	"toxictoast/services/warcraft-service/internal/repository/entity"
	"toxictoast/services/warcraft-service/internal/repository/impl"
	"toxictoast/services/warcraft-service/internal/scheduler"
	"toxictoast/services/warcraft-service/internal/usecase"
	"toxictoast/services/warcraft-service/pkg/blizzard"
	"toxictoast/services/warcraft-service/pkg/config"
)

type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
	Version   string            `json:"version"`
}

var (
	// Build information (set by build flags)
	Version   = "dev"
	BuildTime = "unknown"
	db        *gorm.DB
)

func main() {
	// Load .env file (ignore error in production where env vars are set directly)
	_ = godotenv.Load()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger.Init()

	log.Printf("Starting Warcraft Service v%s (built: %s)", Version, BuildTime)
	log.Printf("Environment: %s", cfg.Environment)

	// Connect to database with retry
	db, err = database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Printf("Database connected successfully")

	// Run auto-migrations
	dbEntities := []interface{}{
		&entity.Faction{},
		&entity.Race{},
		&entity.Class{},
		&entity.Character{},
		&entity.CharacterDetails{},
		&entity.CharacterEquipment{},
		&entity.CharacterStats{},
		&entity.Guild{},
	}
	if err := database.AutoMigrate(db, dbEntities...); err != nil {
		log.Fatalf("Auto-migration failed: %v", err)
	}
	log.Printf("Database schema is up to date")

	// Initialize Blizzard API client
	blizzardClient := blizzard.NewClient(
		cfg.BlizzardClientID,
		cfg.BlizzardClientSecret,
		cfg.BlizzardRegion,
	)
	if cfg.BlizzardClientID != "" {
		log.Printf("Blizzard API client initialized (region: %s)", cfg.BlizzardRegion)
	} else {
		log.Printf("Blizzard API credentials not configured - API calls will fail")
	}

	// Initialize Kafka producer
	var kafkaProducer *kafka.Producer
	if len(cfg.KafkaBrokers) > 0 && cfg.KafkaBrokers[0] != "" {
		var err error
		kafkaProducer, err = kafka.NewProducer(cfg.KafkaBrokers)
		if err != nil {
			log.Printf("Warning: Failed to initialize Kafka producer: %v", err)
			log.Printf("Service will continue without event publishing")
		} else {
			log.Printf("Kafka producer initialized with brokers: %v", cfg.KafkaBrokers)
			defer kafkaProducer.Close()
		}
	} else {
		log.Printf("Kafka brokers not configured - event publishing disabled")
	}

	// Initialize repositories
	characterRepo := impl.NewCharacterRepository(db)
	characterDetailsRepo := impl.NewCharacterDetailsRepository(db)
	characterEquipmentRepo := impl.NewCharacterEquipmentRepository(db)
	characterStatsRepo := impl.NewCharacterStatsRepository(db)
	guildRepo := impl.NewGuildRepository(db)
	raceRepo := impl.NewRaceRepository(db)
	classRepo := impl.NewClassRepository(db)
	factionRepo := impl.NewFactionRepository(db)

	// Initialize use cases
	characterUseCase := usecase.NewCharacterUseCase(
		characterRepo,
		characterDetailsRepo,
		characterEquipmentRepo,
		characterStatsRepo,
		raceRepo,
		classRepo,
		factionRepo,
		guildRepo,
		blizzardClient,
		kafkaProducer,
	)
	guildUseCase := usecase.NewGuildUseCase(guildRepo, factionRepo, blizzardClient, kafkaProducer)

	// Initialize gRPC handlers
	characterHandler := grpcHandler.NewCharacterHandler(characterUseCase)
	guildHandler := grpcHandler.NewGuildHandler(guildUseCase)

	// Initialize background job schedulers
	characterSyncScheduler := scheduler.NewCharacterSyncScheduler(
		characterUseCase,
		cfg.CharacterSyncInterval,
		cfg.CharacterSyncEnabled,
	)
	guildSyncScheduler := scheduler.NewGuildSyncScheduler(
		guildUseCase,
		cfg.GuildSyncInterval,
		cfg.GuildSyncEnabled,
	)

	// Start background jobs
	characterSyncScheduler.Start()
	guildSyncScheduler.Start()
	log.Printf("Background jobs initialized")

	// Setup gRPC server
	grpcServer := setupGRPCServer(characterHandler, guildHandler)

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
	characterSyncScheduler.Stop()
	guildSyncScheduler.Stop()
	log.Println("Background jobs stopped")

	// Shutdown HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Shutdown gRPC server
	grpcServer.GracefulStop()

	log.Println("Servers stopped")
}

func setupGRPCServer(characterHandler *grpcHandler.CharacterHandler, guildHandler *grpcHandler.GuildHandler) *grpc.Server {
	// Setup interceptors
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		sharedgrpc.AuthInterceptor,
	}
	streamInterceptors := []grpc.StreamServerInterceptor{
		sharedgrpc.StreamAuthInterceptor,
	}

	// Create server options with chained interceptors
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	}

	// Create gRPC server with interceptors
	server := grpc.NewServer(opts...)

	// Register services
	pb.RegisterCharacterServiceServer(server, characterHandler)
	pb.RegisterGuildServiceServer(server, guildHandler)

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
	status := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Services:  make(map[string]string),
		Version:   Version,
	}

	// Check database
	if err := checkDatabase(); err != nil {
		status.Status = "degraded"
		status.Services["database"] = fmt.Sprintf("error: %v", err)
	} else {
		status.Services["database"] = "healthy"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	// Check if all dependencies are ready
	if err := checkDatabase(); err != nil {
		http.Error(w, fmt.Sprintf("Database not ready: %v", err), http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ready"))
}

func livenessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Alive"))
}

func checkDatabase() error {
	return database.CheckHealth(db)
}
