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

	"toxictoast/services/sse-service/internal/broker"
	"toxictoast/services/sse-service/internal/consumer"
	grpcHandler "toxictoast/services/sse-service/internal/handler/grpc"
	httpHandler "toxictoast/services/sse-service/internal/handler/http"
	"toxictoast/services/sse-service/internal/scheduler"
	"toxictoast/services/sse-service/pkg/config"
	pb "toxictoast/services/sse-service/api/proto"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	// Load .env file (ignore error in production where env vars are set directly)
	_ = godotenv.Load()

	// Load configuration
	cfg := config.Load()

	log.Printf("Starting SSE Service v%s (built: %s)", Version, BuildTime)
	log.Printf("Environment: %s", cfg.Environment)

	// Create SSE broker
	sseBroker := broker.NewBroker(cfg.SSE.MaxClients, cfg.SSE.HeartbeatSeconds, cfg.SSE.HistorySize)
	sseBroker.Start()

	// Create Kafka consumer
	ctx, cancel := context.WithCancel(context.Background())

	kafkaConsumer, err := consumer.NewKafkaConsumer(&cfg.Kafka, sseBroker)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	kafkaConsumer.Start(ctx)

	// Initialize background job schedulers
	clientCleanupScheduler := scheduler.NewClientCleanupScheduler(
		sseBroker,
		cfg.SSE.ClientCleanupInterval,
		cfg.SSE.ClientCleanupInactiveTimeout,
		cfg.SSE.ClientCleanupEnabled,
	)

	// Start background jobs
	clientCleanupScheduler.Start()
	log.Println("Background jobs initialized")

	// Create HTTP handlers
	sseHandler := httpHandler.NewSSEHandler(
		sseBroker,
		cfg.SSE.EventBufferSize,
		cfg.CORS.AllowedOrigins,
		cfg.CORS.AllowedHeaders,
	)

	// Create rate limiter
	rateLimiter := httpHandler.NewRateLimiter(
		cfg.RateLimit.RequestsPerMin,
		cfg.RateLimit.BurstSize,
		cfg.RateLimit.Enabled,
	)

	// Setup HTTP server
	httpServer := setupHTTPServer(cfg, sseHandler, rateLimiter)

	// Start HTTP server
	go func() {
		log.Printf("🌐 HTTP server starting on port %s", cfg.Port)
		log.Printf("   SSE Endpoint: http://localhost:%s/events", cfg.Port)
		log.Printf("   Health: http://localhost:%s/health", cfg.Port)
		log.Printf("   Stats: http://localhost:%s/stats", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Create gRPC handlers
	managementHandler := grpcHandler.NewManagementHandler(sseBroker)

	// Setup gRPC server
	grpcServer := setupGRPCServer(cfg, managementHandler)

	// Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
		if err != nil {
			log.Fatalf("Failed to listen on gRPC port: %v", err)
		}
		log.Printf("🔧 gRPC server starting on port %s", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down servers...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// 1. Cancel main context to stop Kafka consumer goroutine
	cancel()
	log.Println("   Context cancelled")

	// 2. Stop Kafka consumer (this closes the consumer group)
	if err := kafkaConsumer.Stop(); err != nil {
		log.Printf("   Kafka consumer shutdown error: %v", err)
	}

	// 3. Shutdown HTTP server (this stops new connections)
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("   HTTP server shutdown error: %v", err)
	}

	// 4. Shutdown gRPC server
	grpcServer.GracefulStop()
	log.Println("   gRPC server stopped")

	// 5. Stop background jobs
	clientCleanupScheduler.Stop()
	log.Println("   Background jobs stopped")

	// 6. Stop SSE broker last (this closes all client connections)
	sseBroker.Stop()

	log.Println("✅ All servers stopped gracefully")
}

func setupHTTPServer(cfg *config.Config, sseHandler *httpHandler.SSEHandler, rateLimiter *httpHandler.RateLimiter) *http.Server {
	router := mux.NewRouter()

	// SSE endpoint with rate limiting
	router.HandleFunc("/events", rateLimiter.Middleware(sseHandler.HandleSSE)).Methods("GET", "OPTIONS")

	// Health check endpoints
	router.HandleFunc("/health", sseHandler.HandleHealth).Methods("GET")
	router.HandleFunc("/stats", sseHandler.HandleStats).Methods("GET")

	return &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: 0, // Disable WriteTimeout for SSE (persistent connections)
		IdleTimeout:  cfg.Server.IdleTimeout,
	}
}

func setupGRPCServer(cfg *config.Config, managementHandler *grpcHandler.ManagementHandler) *grpc.Server {
	var opts []grpc.ServerOption

	// Create gRPC server
	server := grpc.NewServer(opts...)

	// Register services
	log.Println("Registering gRPC services...")
	pb.RegisterSSEManagementServiceServer(server, managementHandler)
	log.Println("gRPC services registered")

	// Enable reflection for tools like grpcurl
	reflection.Register(server)

	return server
}
