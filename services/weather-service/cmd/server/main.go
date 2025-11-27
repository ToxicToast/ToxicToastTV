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

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	pb "toxictoast/services/weather-service/api/proto"
	grpcHandler "toxictoast/services/weather-service/internal/handler/grpc"
	"toxictoast/services/weather-service/internal/query"
	"toxictoast/services/weather-service/pkg/config"
	"toxictoast/services/weather-service/pkg/openmeteo"
)

const version = "dev"

func main() {
	log.Printf("Starting Weather Service v%s (built: unknown)", version)

	// Load configuration
	cfg := config.Load()
	log.Printf("Environment: %s", cfg.Environment)

	// Initialize OpenMeteo client
	openMeteoClient := openmeteo.New()

	// Initialize CQRS buses
	log.Println("Initializing CQRS buses...")
	queryBus := cqrs.NewQueryBus()
	log.Println("CQRS buses initialized")

	// Register query handlers
	log.Println("Registering query handlers...")
	queryBus.RegisterHandler("get_current_weather", query.NewGetCurrentWeatherHandler(openMeteoClient))
	queryBus.RegisterHandler("get_forecast", query.NewGetForecastHandler(openMeteoClient))
	log.Println("Query handlers registered")

	// Initialize gRPC handler
	log.Println("Initializing gRPC handlers...")
	weatherHandler := grpcHandler.NewWeatherHandler(queryBus, version)
	log.Println("gRPC handlers initialized")

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register services
	log.Println("Registering gRPC services...")
	pb.RegisterWeatherServiceServer(grpcServer, weatherHandler)

	// Register health check
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	log.Println("All gRPC services registered")

	// Start HTTP server for health checks
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: httpMux,
	}

	// Start HTTP server in goroutine
	go func() {
		log.Printf("HTTP server starting on port %s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Start gRPC server in goroutine
	go func() {
		lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}

		log.Printf("gRPC server starting on port %s", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
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

	// Stop gRPC server
	grpcServer.GracefulStop()

	// Stop HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	log.Println("Servers stopped")
}
