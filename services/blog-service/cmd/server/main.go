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

	"github.com/toxictoast/toxictoastgo/shared/auth"
	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"github.com/toxictoast/toxictoastgo/shared/database"
	sharedgrpc "github.com/toxictoast/toxictoastgo/shared/grpc"
	"github.com/toxictoast/toxictoastgo/shared/kafka"
	"github.com/toxictoast/toxictoastgo/shared/logger"

	pb "toxictoast/services/blog-service/api/proto"
	"toxictoast/services/blog-service/internal/command"
	grpcHandler "toxictoast/services/blog-service/internal/handler/grpc"
	"toxictoast/services/blog-service/internal/query"
	"toxictoast/services/blog-service/internal/repository"
	"toxictoast/services/blog-service/internal/repository/entity"
	"toxictoast/services/blog-service/internal/scheduler"
	"toxictoast/services/blog-service/pkg/config"
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
	cfg := config.Load()

	// Initialize logger
	logger.Init()

	log.Printf("Starting Blog Service v%s (built: %s)", Version, BuildTime)
	log.Printf("Environment: %s", cfg.Environment)

	// Connect to database with retry
	var err error
	db, err = database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Printf("Database connected successfully")

	// Run auto-migrations using entities with blog_ prefix
	dbEntities := []interface{}{
		&entity.PostEntity{},
		&entity.CategoryEntity{},
		&entity.TagEntity{},
		&entity.MediaEntity{},
		&entity.CommentEntity{},
	}
	if err := database.AutoMigrate(db, dbEntities...); err != nil {
		log.Fatalf("Auto-migration failed: %v", err)
	}
	log.Printf("Database schema is up to date (using blog_ table prefix)")

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
	postRepo := repository.NewPostRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	tagRepo := repository.NewTagRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	mediaRepo := repository.NewMediaRepository(db)

	// Initialize Command Bus
	commandBus := cqrs.NewCommandBus()

	// Register Command Handlers - Post (6 commands)
	commandBus.RegisterHandler("create_post", command.NewCreatePostHandler(postRepo, categoryRepo, tagRepo, kafkaProducer))
	commandBus.RegisterHandler("update_post", command.NewUpdatePostHandler(postRepo, categoryRepo, tagRepo, kafkaProducer))
	commandBus.RegisterHandler("delete_post", command.NewDeletePostHandler(postRepo, kafkaProducer))
	commandBus.RegisterHandler("publish_post", command.NewPublishPostHandler(postRepo, kafkaProducer))
	commandBus.RegisterHandler("increment_post_view_count", command.NewIncrementPostViewCountHandler(postRepo))
	commandBus.RegisterHandler("publish_scheduled_post", command.NewPublishScheduledPostHandler(postRepo, kafkaProducer))

	// Register Command Handlers - Category (3 commands)
	commandBus.RegisterHandler("create_category", command.NewCreateCategoryHandler(categoryRepo, kafkaProducer))
	commandBus.RegisterHandler("update_category", command.NewUpdateCategoryHandler(categoryRepo, kafkaProducer))
	commandBus.RegisterHandler("delete_category", command.NewDeleteCategoryHandler(categoryRepo, kafkaProducer))

	// Register Command Handlers - Tag (3 commands)
	commandBus.RegisterHandler("create_tag", command.NewCreateTagHandler(tagRepo, kafkaProducer))
	commandBus.RegisterHandler("update_tag", command.NewUpdateTagHandler(tagRepo, kafkaProducer))
	commandBus.RegisterHandler("delete_tag", command.NewDeleteTagHandler(tagRepo, kafkaProducer))

	// Register Command Handlers - Comment (4 commands)
	commandBus.RegisterHandler("create_comment", command.NewCreateCommentHandler(commentRepo, postRepo, kafkaProducer))
	commandBus.RegisterHandler("update_comment", command.NewUpdateCommentHandler(commentRepo))
	commandBus.RegisterHandler("delete_comment", command.NewDeleteCommentHandler(commentRepo, kafkaProducer))
	commandBus.RegisterHandler("moderate_comment", command.NewModerateCommentHandler(commentRepo, kafkaProducer))

	// Register Command Handlers - Media (2 commands)
	uploadMediaHandler, err := command.NewUploadMediaHandler(mediaRepo, kafkaProducer, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize upload media handler: %v", err)
	}
	deleteMediaHandler, err := command.NewDeleteMediaHandler(mediaRepo, kafkaProducer, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize delete media handler: %v", err)
	}
	commandBus.RegisterHandler("upload_media", uploadMediaHandler)
	commandBus.RegisterHandler("delete_media", deleteMediaHandler)

	logger.Info("Command Bus initialized with 18 command handlers")

	// Initialize Query Bus
	queryBus := cqrs.NewQueryBus()

	// Register Query Handlers - Post (3 queries)
	queryBus.RegisterHandler("get_post_by_id", query.NewGetPostByIDHandler(postRepo))
	queryBus.RegisterHandler("get_post_by_slug", query.NewGetPostBySlugHandler(postRepo))
	queryBus.RegisterHandler("list_posts", query.NewListPostsHandler(postRepo))

	// Register Query Handlers - Category (4 queries)
	queryBus.RegisterHandler("get_category_by_id", query.NewGetCategoryByIDHandler(categoryRepo))
	queryBus.RegisterHandler("get_category_by_slug", query.NewGetCategoryBySlugHandler(categoryRepo))
	queryBus.RegisterHandler("list_categories", query.NewListCategoriesHandler(categoryRepo))
	queryBus.RegisterHandler("get_category_children", query.NewGetCategoryChildrenHandler(categoryRepo))

	// Register Query Handlers - Tag (3 queries)
	queryBus.RegisterHandler("get_tag_by_id", query.NewGetTagByIDHandler(tagRepo))
	queryBus.RegisterHandler("get_tag_by_slug", query.NewGetTagBySlugHandler(tagRepo))
	queryBus.RegisterHandler("list_tags", query.NewListTagsHandler(tagRepo))

	// Register Query Handlers - Comment (3 queries)
	queryBus.RegisterHandler("get_comment_by_id", query.NewGetCommentByIDHandler(commentRepo))
	queryBus.RegisterHandler("list_comments", query.NewListCommentsHandler(commentRepo))
	queryBus.RegisterHandler("get_comment_replies", query.NewGetCommentRepliesHandler(commentRepo))

	// Register Query Handlers - Media (2 queries)
	queryBus.RegisterHandler("get_media_by_id", query.NewGetMediaByIDHandler(mediaRepo))
	queryBus.RegisterHandler("list_media", query.NewListMediaHandler(mediaRepo))

	logger.Info("Query Bus initialized with 15 query handlers")

	// Initialize background job schedulers
	postPublisherScheduler := scheduler.NewPostPublisherScheduler(
		commandBus,
		postRepo,
		cfg.PostPublisherInterval,
		cfg.PostPublisherEnabled,
	)

	// Start background jobs
	postPublisherScheduler.Start()
	log.Printf("Background jobs initialized")

	// Initialize gRPC handlers with CQRS components
	postHandler := grpcHandler.NewPostHandler(commandBus, queryBus, cfg.AuthEnabled)
	categoryHandler := grpcHandler.NewCategoryHandler(commandBus, queryBus, cfg.AuthEnabled)
	tagHandler := grpcHandler.NewTagHandler(commandBus, queryBus, cfg.AuthEnabled)
	commentHandler := grpcHandler.NewCommentHandler(commandBus, queryBus, cfg.AuthEnabled)
	mediaHandler := grpcHandler.NewMediaHandler(commandBus, queryBus, cfg.AuthEnabled)

	// Create composed blog handler
	blogHandler := grpcHandler.NewBlogHandler(postHandler, categoryHandler, tagHandler, commentHandler, mediaHandler)

	// Setup gRPC server
	grpcServer := setupGRPCServer(cfg, keycloakAuth, blogHandler)

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
	postPublisherScheduler.Stop()
	log.Println("Background jobs stopped")

	// Shutdown HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Shutdown gRPC server
	grpcServer.GracefulStop()

	log.Println("Servers stopped")
}

func setupGRPCServer(cfg *config.Config, keycloakAuth *auth.KeycloakAuth, blogHandler *grpcHandler.BlogHandler) *grpc.Server {
	// Setup interceptors
	var opts []grpc.ServerOption

	// Always add auth interceptor (extracts user from metadata)
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		sharedgrpc.AuthInterceptor,
	}
	streamInterceptors := []grpc.StreamServerInterceptor{
		sharedgrpc.StreamAuthInterceptor,
	}

	// Add Keycloak interceptor if enabled
	if keycloakAuth != nil {
		unaryInterceptors = append(unaryInterceptors, keycloakAuth.UnaryInterceptor())
		streamInterceptors = append(streamInterceptors, keycloakAuth.StreamInterceptor())
	}

	// Chain interceptors
	opts = append(opts,
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	)

	// Create gRPC server
	server := grpc.NewServer(opts...)

	// Register services
	pb.RegisterBlogServiceServer(server, blogHandler)

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
