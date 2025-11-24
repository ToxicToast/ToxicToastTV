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
	"github.com/toxictoast/toxictoastgo/shared/jwt"
	"github.com/toxictoast/toxictoastgo/shared/kafka"
	sharedlogger "github.com/toxictoast/toxictoastgo/shared/logger"

	authpb "toxictoast/services/auth-service/api/proto"
	grpchandler "toxictoast/services/auth-service/internal/handler/grpc"
	"toxictoast/services/auth-service/internal/repository/entity"
	"toxictoast/services/auth-service/internal/repository/impl"
	"toxictoast/services/auth-service/internal/usecase"
	"toxictoast/services/auth-service/pkg/config"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger := sharedlogger.NewLogger(cfg.ServiceName)
	logger.Info("Starting Auth Service")

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
		&entity.RoleEntity{},
		&entity.PermissionEntity{},
		&entity.UserRoleEntity{},
		&entity.RolePermissionEntity{},
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
	roleRepo := impl.NewRoleRepository(db)
	permissionRepo := impl.NewPermissionRepository(db)
	userRoleRepo := impl.NewUserRoleRepository(db)
	rolePermissionRepo := impl.NewRolePermissionRepository(db)

	// Initialize JWT helper
	jwtHelper := jwt.NewJWTHelper(
		cfg.JWT.SecretKey,
		cfg.JWT.AccessTokenDuration,
		cfg.JWT.RefreshTokenDuration,
	)

	// Initialize use cases
	authUseCase := usecase.NewAuthUseCase(
		userRoleRepo,
		rolePermissionRepo,
		jwtHelper,
		cfg.UserServiceAddr,
		kafkaProducer,
	)
	roleUseCase := usecase.NewRoleUseCase(roleRepo)
	permissionUseCase := usecase.NewPermissionUseCase(permissionRepo)
	rbacUseCase := usecase.NewRBACUseCase(
		userRoleRepo,
		rolePermissionRepo,
		roleRepo,
		permissionRepo,
	)

	// Initialize gRPC handler
	authHandler := grpchandler.NewAuthHandler(
		authUseCase,
		roleUseCase,
		permissionUseCase,
		rbacUseCase,
	)

	// Create gRPC server
	server := grpc.NewServer()
	authpb.RegisterAuthServiceServer(server, authHandler)

	// Register reflection service (for grpcurl)
	reflection.Register(server)

	// Start gRPC server
	grpcAddr := fmt.Sprintf(":%d", cfg.GRPCPort)
	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to listen on %s: %v", grpcAddr, err))
	}

	logger.Info(fmt.Sprintf("Starting gRPC server on port %d", cfg.GRPCPort))

	// Handle graceful shutdown
	go func() {
		if err := server.Serve(listener); err != nil {
			logger.Fatal(fmt.Sprintf("Failed to serve gRPC: %v", err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Auth Service...")
	server.GracefulStop()
	logger.Info("Auth Service stopped")
}
