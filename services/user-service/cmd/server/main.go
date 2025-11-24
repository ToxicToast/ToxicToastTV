package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/toxictoast/toxictoastgo/shared/database"
	"github.com/toxictoast/toxictoastgo/shared/kafka"
	sharedlogger "github.com/toxictoast/toxictoastgo/shared/logger"

	userpb "toxictoast/services/user-service/api/proto"
	grpchandler "toxictoast/services/user-service/internal/handler/grpc"
	"toxictoast/services/user-service/internal/repository/entity"
	"toxictoast/services/user-service/internal/repository/impl"
	"toxictoast/services/user-service/internal/usecase"
	"toxictoast/services/user-service/pkg/config"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger := sharedlogger.NewLogger(cfg.ServiceName)
	logger.Info("Starting User Service")

	// Connect to database
	db, err := database.NewPostgresConnection(
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
	)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to connect to database: %v", err))
	}
	logger.Info("Connected to database")

	// Auto-migrate database tables
	if err := db.AutoMigrate(
		&entity.UserEntity{},
	); err != nil {
		logger.Fatal(fmt.Sprintf("Failed to auto-migrate: %v", err))
	}
	logger.Info("Database migration completed")

	// Initialize Kafka producer
	kafkaProducer, err := kafka.NewProducer(cfg.Kafka.Brokers)
	if err != nil {
		logger.Info(fmt.Sprintf("Warning: Failed to initialize Kafka producer: %v", err))
		logger.Info("Service will continue without event publishing")
		kafkaProducer = nil
	} else {
		logger.Info("Kafka producer connected successfully")
		defer kafkaProducer.Close()
	}

	// Initialize repositories
	userRepo := impl.NewUserRepository(db)

	// Initialize use cases
	userUseCase := usecase.NewUserUseCase(userRepo, kafkaProducer)

	// Initialize gRPC handler
	userHandler := grpchandler.NewUserHandler(userUseCase)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	userpb.RegisterUserServiceServer(grpcServer, userHandler)

	// Register reflection service (for grpcurl)
	reflection.Register(grpcServer)

	// Start gRPC server
	grpcAddr := fmt.Sprintf(":%d", cfg.GRPCPort)
	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to listen on %s: %v", grpcAddr, err))
	}

	logger.Info(fmt.Sprintf("Starting gRPC server on port %d", cfg.GRPCPort))

	// Handle graceful shutdown
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			logger.Fatal(fmt.Sprintf("Failed to serve gRPC: %v", err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down User Service...")
	grpcServer.GracefulStop()
	logger.Info("User Service stopped")
}
