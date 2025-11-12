package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/toxictoast/toxictoastgo/shared/database"
	"github.com/toxictoast/toxictoastgo/shared/logger"

	pb "toxictoast/services/webhook-service/api/proto"
	"toxictoast/services/webhook-service/internal/consumer"
	"toxictoast/services/webhook-service/internal/delivery"
	grpcHandler "toxictoast/services/webhook-service/internal/handler/grpc"
	"toxictoast/services/webhook-service/internal/repository/entity"
	"toxictoast/services/webhook-service/internal/repository/impl"
	"toxictoast/services/webhook-service/internal/usecase"
	"toxictoast/services/webhook-service/pkg/config"
)

func main() {
	// Load .env file (ignore error in production where env vars are set directly)
	_ = godotenv.Load()

	// Initialize logger
	logger.Init()
	logger.Info("Starting Webhook Service...")

	// Load configuration
	cfg := config.Load()
	logger.Info(fmt.Sprintf("Loaded configuration: gRPC port %s, %d Kafka topics", cfg.GRPCPort, len(cfg.Kafka.Topics)))

	// Initialize database
	db, err := database.Connect(cfg.Database)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to database: %v", err))
		os.Exit(1)
	}
	logger.Info("Connected to database successfully")

	// Auto-migrate database schemas
	if err := db.AutoMigrate(
		&entity.WebhookEntity{},
		&entity.DeliveryEntity{},
		&entity.DeliveryAttemptEntity{},
	); err != nil {
		logger.Error(fmt.Sprintf("Failed to migrate database: %v", err))
		os.Exit(1)
	}
	logger.Info("Database migration completed (using webhook_ table prefix)")

	// Initialize repositories
	webhookRepo := impl.NewWebhookRepository(db)
	deliveryRepo := impl.NewDeliveryRepository(db)
	logger.Info("Repositories initialized")

	// Initialize delivery worker
	deliveryWorker := delivery.NewWorker(
		deliveryRepo,
		delivery.WorkerConfig{
			MaxRetries:        cfg.Webhook.MaxRetries,
			InitialRetryDelay: cfg.Webhook.InitialRetryDelay,
			MaxRetryDelay:     cfg.Webhook.MaxRetryDelay,
			Timeout:           cfg.Webhook.DeliveryTimeout,
		},
	)
	logger.Info("Delivery worker created")

	// Initialize delivery pool
	deliveryPool := delivery.NewPool(
		deliveryWorker,
		deliveryRepo,
		webhookRepo,
		delivery.PoolConfig{
			WorkerCount:        cfg.Webhook.WorkerCount,
			QueueSize:          cfg.Webhook.QueueSize,
			RetryCheckInterval: 1 * time.Minute, // Check for retries every minute
		},
	)
	logger.Info("Delivery pool created")

	// Start delivery pool
	deliveryPool.Start()
	logger.Info("Delivery pool started")

	// Initialize use cases
	webhookUC := usecase.NewWebhookUseCase(webhookRepo)
	deliveryUC := usecase.NewDeliveryUseCase(deliveryRepo, webhookRepo, deliveryPool)
	logger.Info("Use cases initialized")

	// Initialize Kafka consumer
	kafkaConsumer := consumer.NewKafkaConsumer(
		consumer.Config{
			Brokers:     cfg.Kafka.Brokers,
			Topics:      cfg.Kafka.Topics,
			GroupID:     cfg.Kafka.GroupID,
			WorkerCount: cfg.Webhook.WorkerCount,
		},
		deliveryUC,
	)
	logger.Info("Kafka consumer created")

	// Start Kafka consumer
	if err := kafkaConsumer.Start(); err != nil {
		logger.Error(fmt.Sprintf("Failed to start Kafka consumer: %v", err))
		os.Exit(1)
	}
	logger.Info("Kafka consumer started")

	// Initialize gRPC handlers
	webhookHandler := grpcHandler.NewWebhookHandler(webhookUC, deliveryUC)
	deliveryHandler := grpcHandler.NewDeliveryHandler(deliveryUC)
	logger.Info("gRPC handlers initialized")

	// Create gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterWebhookManagementServiceServer(grpcServer, webhookHandler)
	pb.RegisterDeliveryServiceServer(grpcServer, deliveryHandler)
	reflection.Register(grpcServer)
	logger.Info("gRPC server created and services registered")

	// Start gRPC server
	grpcAddr := fmt.Sprintf(":%s", cfg.GRPCPort)
	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to listen on %s: %v", grpcAddr, err))
		os.Exit(1)
	}

	go func() {
		logger.Info(fmt.Sprintf("gRPC server listening on %s", grpcAddr))
		if err := grpcServer.Serve(listener); err != nil {
			logger.Error(fmt.Sprintf("gRPC server error: %v", err))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Webhook Service...")

	// Graceful shutdown with timeout
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop Kafka consumer
	if err := kafkaConsumer.Stop(); err != nil {
		logger.Error(fmt.Sprintf("Error stopping Kafka consumer: %v", err))
	}

	// Stop delivery pool
	deliveryPool.Stop()

	// Stop gRPC server
	grpcServer.GracefulStop()

	// Close database connection
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}

	logger.Info("Webhook Service stopped")
}
