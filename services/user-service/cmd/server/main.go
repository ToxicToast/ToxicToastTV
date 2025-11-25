package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"github.com/toxictoast/toxictoastgo/shared/database"
	"github.com/toxictoast/toxictoastgo/shared/eventstore"
	"github.com/toxictoast/toxictoastgo/shared/kafka"
	sharedlogger "github.com/toxictoast/toxictoastgo/shared/logger"

	userpb "toxictoast/services/user-service/api/proto"
	"toxictoast/services/user-service/internal/command"
	grpchandler "toxictoast/services/user-service/internal/handler/grpc"
	"toxictoast/services/user-service/internal/projection"
	"toxictoast/services/user-service/internal/query"
	"toxictoast/services/user-service/internal/repository/entity"
	"toxictoast/services/user-service/internal/repository/impl"
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

	// Get raw SQL database connection for Event Store
	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to get SQL DB: %v", err))
	}

	// Initialize Event Store (automatically creates tables)
	eventStore, err := eventstore.NewPostgresEventStore(sqlDB)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to initialize event store: %v", err))
	}
	logger.Info("Event Store initialized")

	// Initialize Aggregate Repository
	aggRepo := eventstore.NewAggregateRepository(eventStore)

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

	// Initialize legacy repository (for command handlers validation)
	userRepo := impl.NewUserRepository(db)

	// ============================================
	// CQRS SETUP
	// ============================================

	// Initialize Command Bus
	commandBus := cqrs.NewCommandBus()

	// Register Command Handlers
	commandBus.RegisterHandler("create_user", command.NewCreateUserHandler(aggRepo, userRepo))
	commandBus.RegisterHandler("change_email", command.NewChangeEmailHandler(aggRepo, userRepo))
	commandBus.RegisterHandler("change_password", command.NewChangePasswordHandler(aggRepo))
	commandBus.RegisterHandler("update_password_hash", command.NewUpdatePasswordHashHandler(aggRepo))
	commandBus.RegisterHandler("update_profile", command.NewUpdateProfileHandler(aggRepo))
	commandBus.RegisterHandler("activate_user", command.NewActivateUserHandler(aggRepo))
	commandBus.RegisterHandler("deactivate_user", command.NewDeactivateUserHandler(aggRepo))
	commandBus.RegisterHandler("delete_user", command.NewDeleteUserHandler(aggRepo))
	logger.Info("Command Bus initialized with 8 command handlers")

	// Initialize Read Model Repository
	readModelRepo := projection.NewUserReadModelRepository(sqlDB)

	// Create Read Model table
	if err := readModelRepo.CreateTable(); err != nil {
		logger.Fatal(fmt.Sprintf("Failed to create read model table: %v", err))
	}
	logger.Info("Read Model tables created")

	// Initialize Query Bus
	queryBus := cqrs.NewQueryBus()

	// Register Query Handlers
	queryBus.RegisterHandler("get_user_by_id", query.NewGetUserByIDHandler(readModelRepo))
	queryBus.RegisterHandler("get_user_by_email", query.NewGetUserByEmailHandler(readModelRepo))
	queryBus.RegisterHandler("get_user_by_username", query.NewGetUserByUsernameHandler(readModelRepo))
	queryBus.RegisterHandler("list_users", query.NewListUsersHandler(readModelRepo))
	queryBus.RegisterHandler("get_user_password_hash", query.NewGetUserPasswordHashHandler(eventStore))
	logger.Info("Query Bus initialized with 5 query handlers")

	// Initialize Projector
	userProjector := projection.NewUserProjector(readModelRepo)

	// Initialize Projector Manager
	projectorManager := cqrs.NewProjectorManager(eventStore)
	projectorManager.RegisterProjector(userProjector)
	logger.Info("Projector Manager initialized")

	// Optional: Rebuild projections from history
	ctx := context.Background()
	if err := projectorManager.RebuildProjections(ctx, eventstore.AggregateTypeUser); err != nil {
		logger.Info(fmt.Sprintf("Warning: Failed to rebuild projections: %v", err))
	} else {
		logger.Info("Projections rebuilt successfully")
	}

	// Initialize gRPC handler with CQRS components
	userHandler := grpchandler.NewUserHandler(commandBus, queryBus, readModelRepo)

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
