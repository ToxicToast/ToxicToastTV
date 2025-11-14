package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/toxictoast/toxictoastgo/shared/logger"

	pb "toxictoast/services/notification-service/api/proto"
	"toxictoast/services/notification-service/internal/consumer"
	"toxictoast/services/notification-service/internal/discord"
	grpcHandler "toxictoast/services/notification-service/internal/handler/grpc"
	"toxictoast/services/notification-service/internal/repository/entity"
	"toxictoast/services/notification-service/internal/repository/impl"
	"toxictoast/services/notification-service/internal/scheduler"
	"toxictoast/services/notification-service/internal/usecase"
	"toxictoast/services/notification-service/pkg/config"
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

	// Initialize logger
	logger.Init()
	logger.Info("Starting Notification Service...")

	// Load configuration
	cfg := config.Load()
	logger.Info(fmt.Sprintf("Loaded configuration: gRPC port %s, HTTP port %s, %d Kafka topics", cfg.GRPCPort, cfg.Port, len(cfg.Kafka.Topics)))

	// Initialize database
	var err error
	db, err = database.Connect(cfg.Database)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to database: %v", err))
		os.Exit(1)
	}
	logger.Info("Connected to database successfully")

	// Auto-migrate database schemas using entities
	if err := db.AutoMigrate(
		&entity.DiscordChannelEntity{},
		&entity.NotificationEntity{},
		&entity.NotificationAttemptEntity{},
	); err != nil {
		logger.Error(fmt.Sprintf("Failed to migrate database: %v", err))
		os.Exit(1)
	}
	logger.Info("Database migration completed")

	// Initialize repositories
	channelRepo := impl.NewDiscordChannelRepository(db)
	notificationRepo := impl.NewNotificationRepository(db)
	logger.Info("Repositories initialized")

	// Initialize Discord client
	discordClient := discord.NewClient()
	logger.Info("Discord client created")

	// Initialize use cases
	channelUC := usecase.NewChannelUseCase(channelRepo)
	notificationUC := usecase.NewNotificationUseCase(notificationRepo, channelRepo, discordClient)
	logger.Info("Use cases initialized")

	// Initialize Kafka consumer
	kafkaConsumer := consumer.NewKafkaConsumer(
		consumer.Config{
			Brokers:     cfg.Kafka.Brokers,
			Topics:      cfg.Kafka.Topics,
			GroupID:     cfg.Kafka.GroupID,
			WorkerCount: 5,
		},
		notificationUC,
	)
	logger.Info("Kafka consumer created")

	// Start Kafka consumer
	if err := kafkaConsumer.Start(); err != nil {
		logger.Error(fmt.Sprintf("Failed to start Kafka consumer: %v", err))
		os.Exit(1)
	}
	logger.Info("Kafka consumer started")

	// Initialize background job schedulers
	retryScheduler := scheduler.NewNotificationRetryScheduler(
		notificationUC,
		notificationRepo,
		cfg.NotificationRetryInterval,
		cfg.NotificationRetryMaxRetries,
		cfg.NotificationRetryEnabled,
	)

	cleanupScheduler := scheduler.NewNotificationCleanupScheduler(
		notificationRepo,
		cfg.NotificationCleanupInterval,
		cfg.NotificationCleanupRetentionDays,
		cfg.NotificationCleanupEnabled,
	)

	// Start background jobs
	retryScheduler.Start()
	cleanupScheduler.Start()
	logger.Info("Background jobs initialized")

	// Initialize gRPC handlers
	channelHandler := grpcHandler.NewChannelHandler(channelUC, notificationUC)
	notificationHandler := grpcHandler.NewNotificationHandler(notificationUC)
	logger.Info("gRPC handlers initialized")

	// Setup gRPC server
	grpcServer := setupGRPCServer(channelHandler, notificationHandler)

	// Start gRPC server
	go func() {
		grpcAddr := fmt.Sprintf(":%s", cfg.GRPCPort)
		listener, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to listen on %s: %v", grpcAddr, err))
			os.Exit(1)
		}
		logger.Info(fmt.Sprintf("gRPC server listening on %s", grpcAddr))
		if err := grpcServer.Serve(listener); err != nil {
			logger.Error(fmt.Sprintf("gRPC server error: %v", err))
			os.Exit(1)
		}
	}()

	// Setup HTTP server for health checks
	httpServer := setupHTTPServer(cfg)

	// Start HTTP server
	go func() {
		logger.Info(fmt.Sprintf("HTTP server starting on port %s", cfg.Port))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(fmt.Sprintf("HTTP server failed: %v", err))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Notification Service...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error(fmt.Sprintf("HTTP server shutdown error: %v", err))
	}

	// Stop Kafka consumer
	if err := kafkaConsumer.Stop(); err != nil {
		logger.Error(fmt.Sprintf("Error stopping Kafka consumer: %v", err))
	}

	// Stop background jobs
	retryScheduler.Stop()
	cleanupScheduler.Stop()
	logger.Info("Background jobs stopped")

	// Stop gRPC server
	grpcServer.GracefulStop()

	// Close database connection
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}

	logger.Info("Notification Service stopped")
}

func setupGRPCServer(channelHandler *grpcHandler.ChannelHandler, notificationHandler *grpcHandler.NotificationHandler) *grpc.Server {
	// Create gRPC server
	server := grpc.NewServer()

	// Register services
	pb.RegisterChannelManagementServiceServer(server, channelHandler)
	pb.RegisterNotificationServiceServer(server, notificationHandler)

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
