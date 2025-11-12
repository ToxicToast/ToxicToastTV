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

	pb "toxictoast/services/notification-service/api/proto"
	"toxictoast/services/notification-service/internal/consumer"
	"toxictoast/services/notification-service/internal/discord"
	grpcHandler "toxictoast/services/notification-service/internal/handler/grpc"
	"toxictoast/services/notification-service/internal/repository/entity"
	"toxictoast/services/notification-service/internal/repository/impl"
	"toxictoast/services/notification-service/internal/usecase"
	"toxictoast/services/notification-service/pkg/config"
)

func main() {
	// Load .env file (ignore error in production where env vars are set directly)
	_ = godotenv.Load()

	// Initialize logger
	logger.Init()
	logger.Info("Starting Notification Service...")

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

	// Initialize gRPC handlers
	channelHandler := grpcHandler.NewChannelHandler(channelUC, notificationUC)
	notificationHandler := grpcHandler.NewNotificationHandler(notificationUC)
	logger.Info("gRPC handlers initialized")

	// Create gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterChannelManagementServiceServer(grpcServer, channelHandler)
	pb.RegisterNotificationServiceServer(grpcServer, notificationHandler)
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

	logger.Info("Shutting down Notification Service...")

	// Graceful shutdown with timeout
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop Kafka consumer
	if err := kafkaConsumer.Stop(); err != nil {
		logger.Error(fmt.Sprintf("Error stopping Kafka consumer: %v", err))
	}

	// Stop gRPC server
	grpcServer.GracefulStop()

	// Close database connection
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}

	logger.Info("Notification Service stopped")
}
