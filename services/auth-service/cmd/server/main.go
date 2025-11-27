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

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"github.com/toxictoast/toxictoastgo/shared/database"
	"github.com/toxictoast/toxictoastgo/shared/jwt"
	"github.com/toxictoast/toxictoastgo/shared/kafka"
	sharedlogger "github.com/toxictoast/toxictoastgo/shared/logger"

	authpb "toxictoast/services/auth-service/api/proto"
	"toxictoast/services/auth-service/internal/command"
	grpchandler "toxictoast/services/auth-service/internal/handler/grpc"
	"toxictoast/services/auth-service/internal/query"
	"toxictoast/services/auth-service/internal/repository/entity"
	"toxictoast/services/auth-service/internal/repository/impl"
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

	// ============================================
	// CQRS SETUP
	// ============================================

	// Initialize Command Bus
	commandBus := cqrs.NewCommandBus()

	// Register Command Handlers - Role
	commandBus.RegisterHandler("create_role", command.NewCreateRoleHandler(roleRepo))
	commandBus.RegisterHandler("update_role", command.NewUpdateRoleHandler(roleRepo))
	commandBus.RegisterHandler("delete_role", command.NewDeleteRoleHandler(roleRepo))

	// Register Command Handlers - Permission
	commandBus.RegisterHandler("create_permission", command.NewCreatePermissionHandler(permissionRepo))
	commandBus.RegisterHandler("update_permission", command.NewUpdatePermissionHandler(permissionRepo))
	commandBus.RegisterHandler("delete_permission", command.NewDeletePermissionHandler(permissionRepo))

	// Register Command Handlers - RBAC
	commandBus.RegisterHandler("assign_role", command.NewAssignRoleHandler(userRoleRepo, roleRepo))
	commandBus.RegisterHandler("revoke_role", command.NewRevokeRoleHandler(userRoleRepo))
	commandBus.RegisterHandler("assign_permission", command.NewAssignPermissionHandler(rolePermissionRepo, roleRepo, permissionRepo))
	commandBus.RegisterHandler("revoke_permission", command.NewRevokePermissionHandler(rolePermissionRepo))

	// Register Command Handlers - Auth
	commandBus.RegisterHandler("register", command.NewRegisterHandler(userRoleRepo, rolePermissionRepo, jwtHelper, cfg.UserServiceAddr, kafkaProducer))
	commandBus.RegisterHandler("login", command.NewLoginHandler(cfg.UserServiceAddr))
	commandBus.RegisterHandler("refresh_token", command.NewRefreshTokenHandler(jwtHelper, cfg.UserServiceAddr))

	logger.Info("Command Bus initialized with 13 command handlers")

	// Initialize Query Bus
	queryBus := cqrs.NewQueryBus()

	// Register Query Handlers - Role
	queryBus.RegisterHandler("get_role", query.NewGetRoleHandler(roleRepo))
	queryBus.RegisterHandler("list_roles", query.NewListRolesHandler(roleRepo))

	// Register Query Handlers - Permission
	queryBus.RegisterHandler("get_permission", query.NewGetPermissionHandler(permissionRepo))
	queryBus.RegisterHandler("list_permissions", query.NewListPermissionsHandler(permissionRepo))

	// Register Query Handlers - RBAC
	queryBus.RegisterHandler("get_user_roles", query.NewGetUserRolesHandler(userRoleRepo))
	queryBus.RegisterHandler("get_user_permissions", query.NewGetUserPermissionsHandler(rolePermissionRepo))
	queryBus.RegisterHandler("get_role_permissions", query.NewGetRolePermissionsHandler(rolePermissionRepo))
	queryBus.RegisterHandler("check_permission", query.NewCheckPermissionHandler(rolePermissionRepo))

	// Register Query Handlers - Auth
	queryBus.RegisterHandler("validate_token", query.NewValidateTokenHandler(jwtHelper))

	logger.Info("Query Bus initialized with 9 query handlers")

	// Initialize gRPC handler with CQRS components
	authHandler := grpchandler.NewAuthHandler(
		commandBus,
		queryBus,
		jwtHelper,
		userRoleRepo,
		rolePermissionRepo,
		kafkaProducer,
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
