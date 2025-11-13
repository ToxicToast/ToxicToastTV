package proxy

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"toxictoast/services/gateway-service/internal/handler"
)

// Router handles HTTP routing to gRPC backends
type Router struct {
	clients *ServiceClients
	router  *mux.Router
	devMode bool
}

// NewRouter creates a new HTTP to gRPC router
func NewRouter(clients *ServiceClients, devMode bool) *Router {
	r := &Router{
		clients: clients,
		router:  mux.NewRouter(),
		devMode: devMode,
	}

	r.setupRoutes()
	return r
}

// setupRoutes configures path-based routing
func (r *Router) setupRoutes() {
	// Health check
	r.router.HandleFunc("/health", r.healthCheck).Methods("GET")
	r.router.HandleFunc("/ready", r.readinessCheck).Methods("GET")

	// Prometheus metrics endpoint
	r.router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// Swagger UI (only in DEV mode)
	if r.devMode {
		r.router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
			httpSwagger.URL("/swagger/doc.yaml"),
		))
		// Serve the swagger.yaml file
		r.router.HandleFunc("/swagger/doc.yaml", r.serveSwaggerSpec).Methods("GET")
	}

	// Blog service routes - /api/blog/*
	if r.clients.BlogConn != nil {
		blogHandler := handler.NewBlogHandler(r.clients.BlogConn)
		blogRouter := r.router.PathPrefix("/api/blog").Subrouter()
		blogHandler.RegisterRoutes(blogRouter)
	}

	// Link service routes - /api/links/*
	if r.clients.LinkConn != nil {
		linkHandler := handler.NewLinkHandler(r.clients.LinkConn)
		linkRouter := r.router.PathPrefix("/api/links").Subrouter()
		linkHandler.RegisterRoutes(linkRouter)
	}

	// Foodfolio service routes - /api/foodfolio/*
	if r.clients.FoodfolioConn != nil {
		foodfolioHandler := handler.NewFoodFolioHandler(r.clients.FoodfolioConn)
		foodfolioRouter := r.router.PathPrefix("/api/foodfolio").Subrouter()
		foodfolioHandler.RegisterRoutes(foodfolioRouter)
	}

	// Notification service routes - /api/notifications/*
	if r.clients.NotificationConn != nil {
		notificationHandler := handler.NewNotificationHandler(r.clients.NotificationConn)
		notificationRouter := r.router.PathPrefix("/api/notifications").Subrouter()
		notificationHandler.RegisterRoutes(notificationRouter)
	}

	// SSE service routes - /api/events/*
	if r.clients.SSEConn != nil {
		sseHandler := handler.NewSSEHandler(r.clients.SSEConn)
		sseRouter := r.router.PathPrefix("/api/events").Subrouter()
		sseHandler.RegisterRoutes(sseRouter)
	}

	// TwitchBot service routes - /api/twitch/*
	if r.clients.TwitchBotConn != nil {
		twitchbotHandler := handler.NewTwitchBotHandler(r.clients.TwitchBotConn)
		twitchRouter := r.router.PathPrefix("/api/twitch").Subrouter()
		twitchbotHandler.RegisterRoutes(twitchRouter)
	}

	// Webhook service routes - /api/webhooks/*
	if r.clients.WebhookConn != nil {
		webhookHandler := handler.NewWebhookHandler(r.clients.WebhookConn)
		webhookRouter := r.router.PathPrefix("/api/webhooks").Subrouter()
		webhookHandler.RegisterRoutes(webhookRouter)
	}

	// Warcraft service routes - /api/warcraft/*
	if r.clients.WarcraftConn != nil {
		warcraftHandler := handler.NewWarcraftHandler(r.clients.WarcraftConn)
		warcraftRouter := r.router.PathPrefix("/api/warcraft").Subrouter()
		warcraftHandler.RegisterRoutes(warcraftRouter)
	}
}

// GetRouter returns the mux router
func (r *Router) GetRouter() *mux.Router {
	return r.router
}

// healthCheck endpoint
func (r *Router) healthCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}

// readinessCheck endpoint
func (r *Router) readinessCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := map[string]interface{}{
		"ready": true,
		"services": map[string]bool{
			"blog":         r.clients.BlogConn != nil,
			"link":         r.clients.LinkConn != nil,
			"foodfolio":    r.clients.FoodfolioConn != nil,
			"notification": r.clients.NotificationConn != nil,
			"sse":          r.clients.SSEConn != nil,
			"twitchbot":    r.clients.TwitchBotConn != nil,
			"webhook":      r.clients.WebhookConn != nil,
			"warcraft":     r.clients.WarcraftConn != nil,
		},
	}

	json.NewEncoder(w).Encode(status)
}

// Proxy handlers - These will forward requests to gRPC services
// For now, they return a placeholder response

func (r *Router) proxyToBlog(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/api/blog")
	r.handleProxy(w, req, "blog", path)
}

func (r *Router) proxyToLink(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/api/links")
	r.handleProxy(w, req, "link", path)
}

func (r *Router) proxyToFoodfolio(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/api/foodfolio")
	r.handleProxy(w, req, "foodfolio", path)
}

func (r *Router) proxyToNotification(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/api/notifications")
	r.handleProxy(w, req, "notification", path)
}

func (r *Router) proxyToSSE(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/api/events")
	r.handleProxy(w, req, "sse", path)
}

func (r *Router) proxyToTwitchBot(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/api/twitch")
	r.handleProxy(w, req, "twitchbot", path)
}

func (r *Router) proxyToWebhook(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/api/webhooks")
	r.handleProxy(w, req, "webhook", path)
}

// handleProxy is a generic proxy handler
// This is a placeholder - full implementation would translate HTTP to gRPC
func (r *Router) handleProxy(w http.ResponseWriter, req *http.Request, service, path string) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"message": "Gateway proxy",
		"service": service,
		"path":    path,
		"method":  req.Method,
		"note":    "Full HTTP-to-gRPC translation to be implemented per endpoint",
	}

	json.NewEncoder(w).Encode(response)
}

// serveSwaggerSpec serves the OpenAPI specification file
func (r *Router) serveSwaggerSpec(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/x-yaml")
	http.ServeFile(w, req, "docs/swagger.yaml")
}
