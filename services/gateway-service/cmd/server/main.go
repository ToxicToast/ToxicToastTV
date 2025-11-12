package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/toxictoast/toxictoastgo/shared/auth"
	"github.com/toxictoast/toxictoastgo/shared/config"
	"github.com/toxictoast/toxictoastgo/shared/logger"

	"toxictoast/services/gateway-service/internal/metrics"
	"toxictoast/services/gateway-service/internal/middleware"
	"toxictoast/services/gateway-service/internal/proxy"
	gwconfig "toxictoast/services/gateway-service/pkg/config"
)

func main() {
	// Load configuration (includes loading .env file)
	cfg, err := gwconfig.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Initialize logger
	logger.Init()
	logger.Info("Starting Gateway Service")

	// Initialize metrics
	m := metrics.NewMetrics()
	logger.Info("Prometheus metrics initialized")

	// Initialize Keycloak auth (optional - only if configured)
	if cfg.KeycloakURL != "" && cfg.KeycloakRealm != "" {
		keycloakCfg := &config.KeycloakConfig{
			URL:      cfg.KeycloakURL,
			Realm:    cfg.KeycloakRealm,
			ClientID: cfg.KeycloakClientID,
		}
		_, err = auth.NewKeycloakAuth(keycloakCfg)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to initialize Keycloak auth: %v", err))
		} else {
			logger.Info(fmt.Sprintf("Keycloak authentication enabled - URL: %s, Realm: %s", cfg.KeycloakURL, cfg.KeycloakRealm))
		}
	} else {
		logger.Info("Keycloak authentication disabled")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to backend services
	serviceURLs := proxy.ServiceURLs{
		BlogURL:         cfg.BlogServiceURL,
		LinkURL:         cfg.LinkServiceURL,
		FoodfolioURL:    cfg.FoodfolioServiceURL,
		NotificationURL: cfg.NotificationServiceURL,
		SSEURL:          cfg.SSEServiceURL,
		TwitchBotURL:    cfg.TwitchBotServiceURL,
		WebhookURL:      cfg.WebhookServiceURL,
	}

	clients, err := proxy.NewServiceClients(ctx, serviceURLs)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to backend services: %v", err))
		panic(err)
	}
	defer clients.Close()

	logger.Info("Connected to backend services")

	// Update backend health metrics
	m.SetBackendHealthStatus("blog", clients.BlogConn != nil)
	m.SetBackendHealthStatus("link", clients.LinkConn != nil)
	m.SetBackendHealthStatus("foodfolio", clients.FoodfolioConn != nil)
	m.SetBackendHealthStatus("notification", clients.NotificationConn != nil)
	m.SetBackendHealthStatus("sse", clients.SSEConn != nil)
	m.SetBackendHealthStatus("twitchbot", clients.TwitchBotConn != nil)
	m.SetBackendHealthStatus("webhook", clients.WebhookConn != nil)

	// Create router
	router := proxy.NewRouter(clients, cfg.DevMode)
	handler := router.GetRouter()

	if cfg.DevMode {
		logger.Info("DEV mode enabled - Swagger UI available at /swagger")
	}

	// Apply middleware in order (innermost to outermost)
	var finalHandler http.Handler = handler

	// Metrics middleware (should be innermost to capture all metrics)
	finalHandler = middleware.Metrics(m)(finalHandler)
	logger.Info("Metrics middleware enabled")

	if cfg.EnableCORS {
		finalHandler = middleware.CORS(finalHandler)
		logger.Info("CORS middleware enabled")
	}

	// Rate limiting (with metrics)
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst, m)
	finalHandler = rateLimiter.Middleware(finalHandler)
	logger.Info(fmt.Sprintf("Rate limiting enabled - RPS: %d, Burst: %d", cfg.RateLimitRPS, cfg.RateLimitBurst))

	// Logging middleware (should be outermost)
	finalHandler = middleware.Logging(finalHandler)

	// Optional: Add authentication middleware for protected routes
	// This would require more sophisticated routing setup

	// HTTP Server
	httpServer := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      finalHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server
	go func() {
		logger.Info(fmt.Sprintf("Starting HTTP server on port %s", cfg.HTTPPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(fmt.Sprintf("HTTP server error: %v", err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down gateway service...")

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error(fmt.Sprintf("HTTP server shutdown error: %v", err))
	}

	logger.Info("Gateway service stopped")
}
