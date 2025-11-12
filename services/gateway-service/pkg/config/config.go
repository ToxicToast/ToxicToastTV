package config

import (
	"github.com/toxictoast/toxictoastgo/shared/config"
)

type Config struct {
	// Keycloak configuration
	KeycloakURL      string
	KeycloakRealm    string
	KeycloakClientID string

	// Gateway specific config
	HTTPPort       string
	GRPCPort       string
	EnableCORS     bool
	RateLimitRPS   int
	RateLimitBurst int

	// Service endpoints
	BlogServiceURL         string
	LinkServiceURL         string
	FoodfolioServiceURL    string
	NotificationServiceURL string
	SSEServiceURL          string
	TwitchBotServiceURL    string
	WebhookServiceURL      string
}

func Load() (*Config, error) {
	// Load environment file
	config.LoadEnvFile()

	// Load Keycloak config
	keycloakCfg := config.LoadKeycloakConfig()

	cfg := &Config{
		// Keycloak
		KeycloakURL:      keycloakCfg.URL,
		KeycloakRealm:    keycloakCfg.Realm,
		KeycloakClientID: keycloakCfg.ClientID,

		// Gateway config
		HTTPPort:       config.GetEnv("HTTP_PORT", "8080"),
		GRPCPort:       config.GetEnv("GRPC_PORT", "9090"),
		EnableCORS:     config.GetEnvAsBool("ENABLE_CORS", true),
		RateLimitRPS:   config.GetEnvAsInt("RATE_LIMIT_RPS", 100),
		RateLimitBurst: config.GetEnvAsInt("RATE_LIMIT_BURST", 200),

		// Service URLs
		BlogServiceURL:         config.GetEnv("BLOG_SERVICE_URL", "localhost:9090"),
		LinkServiceURL:         config.GetEnv("LINK_SERVICE_URL", "localhost:9090"),
		FoodfolioServiceURL:    config.GetEnv("FOODFOLIO_SERVICE_URL", "localhost:9091"),
		NotificationServiceURL: config.GetEnv("NOTIFICATION_SERVICE_URL", "localhost:9090"),
		SSEServiceURL:          config.GetEnv("SSE_SERVICE_URL", "localhost:9090"),
		TwitchBotServiceURL:    config.GetEnv("TWITCHBOT_SERVICE_URL", "localhost:9093"),
		WebhookServiceURL:      config.GetEnv("WEBHOOK_SERVICE_URL", "localhost:9090"),
	}

	return cfg, nil
}
