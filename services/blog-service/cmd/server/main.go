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
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/gorm"

	"github.com/toxictoast/toxictoastgo/shared/auth"
	"github.com/toxictoast/toxictoastgo/shared/database"
	"github.com/toxictoast/toxictoastgo/shared/kafka"
	"github.com/toxictoast/toxictoastgo/shared/logger"

	pb "toxictoast/services/blog-service/api/proto"
	"toxictoast/services/blog-service/internal/domain"
	grpcHandler "toxictoast/services/blog-service/internal/handler/grpc"
	"toxictoast/services/blog-service/internal/repository"
	"toxictoast/services/blog-service/internal/usecase"
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

	// Run auto-migrations
	entities := []interface{}{
		&domain.Post{},
		&domain.Category{},
		&domain.Tag{},
		&domain.Media{},
		&domain.Comment{},
	}
	if err := database.AutoMigrate(db, entities...); err != nil {
		log.Fatalf("Auto-migration failed: %v", err)
	}
	log.Printf("Database schema is up to date")

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

	// Initialize use cases
	postUseCase := usecase.NewPostUseCase(postRepo, categoryRepo, tagRepo, kafkaProducer, cfg)
	categoryUseCase := usecase.NewCategoryUseCase(categoryRepo, kafkaProducer, cfg)
	tagUseCase := usecase.NewTagUseCase(tagRepo, kafkaProducer, cfg)
	commentUseCase := usecase.NewCommentUseCase(commentRepo, postRepo, kafkaProducer, cfg)
	mediaUseCase, err := usecase.NewMediaUseCase(mediaRepo, kafkaProducer, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize media use case: %v", err)
	}

	// Initialize gRPC handlers
	postHandler := grpcHandler.NewPostHandler(postUseCase, cfg.AuthEnabled)
	categoryHandler := grpcHandler.NewCategoryHandler(categoryUseCase, cfg.AuthEnabled)
	tagHandler := grpcHandler.NewTagHandler(tagUseCase, cfg.AuthEnabled)
	commentHandler := grpcHandler.NewCommentHandler(commentUseCase, cfg.AuthEnabled)
	mediaHandler := grpcHandler.NewMediaHandler(mediaUseCase, cfg.AuthEnabled)

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

	if keycloakAuth != nil {
		opts = append(opts,
			grpc.UnaryInterceptor(keycloakAuth.UnaryInterceptor()),
			grpc.StreamInterceptor(keycloakAuth.StreamInterceptor()),
		)
	}

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
