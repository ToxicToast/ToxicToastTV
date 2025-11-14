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
	"github.com/toxictoast/toxictoastgo/shared/kafka"
	"github.com/toxictoast/toxictoastgo/shared/logger"

	"toxictoast/services/twitchbot-service/pkg/bot"
	"toxictoast/services/twitchbot-service/pkg/config"
	"toxictoast/services/twitchbot-service/pkg/events"

	// Repository layer
	"toxictoast/services/twitchbot-service/internal/repository/entity"
	repoImpl "toxictoast/services/twitchbot-service/internal/repository/impl"

	// Use case layer
	"toxictoast/services/twitchbot-service/internal/usecase"

	// Handler layer
	grpcHandler "toxictoast/services/twitchbot-service/internal/handler/grpc"

	// Scheduler layer
	"toxictoast/services/twitchbot-service/internal/scheduler"

	// Proto definitions
	pb "toxictoast/services/twitchbot-service/api/proto"
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

	log.Printf("Starting Twitchbot Service v%s (built: %s)", Version, BuildTime)
	log.Printf("Environment: %s", cfg.Environment)

	// Connect to database with retry
	var err error
	db, err = database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Printf("Database connected successfully")

	// Run auto-migrations using entities with twitchbot_ prefix
	dbEntities := []interface{}{
		&entity.StreamEntity{},
		&entity.MessageEntity{},
		&entity.ViewerEntity{},
		&entity.ClipEntity{},
		&entity.CommandEntity{},
		&entity.ChannelViewerEntity{},
	}
	if err := database.AutoMigrate(db, dbEntities...); err != nil {
		log.Fatalf("Auto-migration failed: %v", err)
	}
	log.Printf("Database schema is up to date (using twitchbot_ table prefix)")

	// Ensure Chat-Only stream exists (for offline message logging)
	if err := repoImpl.EnsureChatOnlyStream(db); err != nil {
		log.Printf("Warning: Failed to create Chat-Only stream: %v", err)
		log.Printf("Offline messages may not be logged correctly")
	}

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
	streamRepo := repoImpl.NewStreamRepository(db)
	messageRepo := repoImpl.NewMessageRepository(db)
	viewerRepo := repoImpl.NewViewerRepository(db)
	clipRepo := repoImpl.NewClipRepository(db)
	commandRepo := repoImpl.NewCommandRepository(db)
	channelViewerRepo := repoImpl.NewChannelViewerRepository(db)
	log.Println("Repositories initialized")

	// Initialize use cases
	log.Println("Initializing use cases...")
	streamUC := usecase.NewStreamUseCase(streamRepo, messageRepo)
	messageUC := usecase.NewMessageUseCase(messageRepo, streamRepo, viewerRepo)
	viewerUC := usecase.NewViewerUseCase(viewerRepo)
	clipUC := usecase.NewClipUseCase(clipRepo, streamRepo)
	commandUC := usecase.NewCommandUseCase(commandRepo)
	channelViewerUC := usecase.NewChannelViewerUseCase(channelViewerRepo, viewerRepo)
	log.Println("Use cases initialized")

	// Initialize event publisher
	var eventPublisher *events.EventPublisher
	if kafkaProducer != nil {
		eventPublisher = events.NewEventPublisher(kafkaProducer)
		log.Println("Event publisher initialized")
	} else {
		log.Println("Event publisher disabled (Kafka not available)")
	}

	// Initialize and start Twitch bot
	botMgr := bot.NewManager(cfg, streamUC, messageUC, viewerUC, clipUC, commandUC, channelViewerUC, eventPublisher)
	botMgr.Start(context.Background())
	// Note: Bot errors are handled gracefully inside bot.Manager
	// Service continues in API-only mode if bot fails to start

	// Initialize background job schedulers
	messageCleanupScheduler := scheduler.NewMessageCleanupScheduler(
		messageRepo,
		cfg.BackgroundJobs.MessageCleanupInterval,
		cfg.BackgroundJobs.MessageCleanupRetentionDays,
		cfg.BackgroundJobs.MessageCleanupEnabled,
	)

	streamCloserScheduler := scheduler.NewStreamSessionCloserScheduler(
		streamRepo,
		cfg.BackgroundJobs.StreamCloserInterval,
		cfg.BackgroundJobs.StreamCloserInactiveTimeout,
		cfg.BackgroundJobs.StreamCloserEnabled,
	)

	// Start background jobs
	messageCleanupScheduler.Start()
	streamCloserScheduler.Start()
	log.Println("Background jobs initialized")

	// Initialize gRPC handlers
	log.Println("Initializing gRPC handlers...")
	streamHandler := grpcHandler.NewStreamHandler(streamUC)
	messageHandler := grpcHandler.NewMessageHandler(messageUC)
	viewerHandler := grpcHandler.NewViewerHandler(viewerUC)
	clipHandler := grpcHandler.NewClipHandler(clipUC)
	commandHandler := grpcHandler.NewCommandHandler(commandUC)
	botHandler := grpcHandler.NewBotHandler(botMgr)
	channelViewerHandler := grpcHandler.NewChannelViewerHandler(channelViewerUC)
	log.Println("gRPC handlers initialized")

	// Setup gRPC server
	grpcServer := setupGRPCServer(
		cfg,
		keycloakAuth,
		streamHandler,
		messageHandler,
		viewerHandler,
		clipHandler,
		commandHandler,
		botHandler,
		channelViewerHandler,
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

	// Stop Twitch bot
	if botMgr != nil {
		if err := botMgr.Stop(); err != nil {
			log.Printf("Bot manager shutdown error: %v", err)
		}
	}

	// Stop background jobs
	messageCleanupScheduler.Stop()
	streamCloserScheduler.Stop()
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
	streamHandler *grpcHandler.StreamHandler,
	messageHandler *grpcHandler.MessageHandler,
	viewerHandler *grpcHandler.ViewerHandler,
	clipHandler *grpcHandler.ClipHandler,
	commandHandler *grpcHandler.CommandHandler,
	botHandler *grpcHandler.BotHandler,
	channelViewerHandler *grpcHandler.ChannelViewerHandler,
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
	pb.RegisterStreamServiceServer(server, streamHandler)
	pb.RegisterMessageServiceServer(server, messageHandler)
	pb.RegisterViewerServiceServer(server, viewerHandler)
	pb.RegisterClipServiceServer(server, clipHandler)
	pb.RegisterCommandServiceServer(server, commandHandler)
	pb.RegisterBotServiceServer(server, botHandler)
	pb.RegisterChannelViewerServiceServer(server, channelViewerHandler)
	log.Println("All gRPC services registered (7 services)")

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
