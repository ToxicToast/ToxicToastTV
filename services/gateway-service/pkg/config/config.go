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
	DevMode        bool

	// Service endpoints
	BlogServiceURL         string
	LinkServiceURL         string
	FoodfolioServiceURL    string
	NotificationServiceURL string
	SSEServiceURL          string
	TwitchBotServiceURL    string
	WebhookServiceURL      string
	WarcraftServiceURL     string
	WeatherServiceURL      string
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
		HTTPPort:       config.GetEnv("HTTP_PORT", "8081"),
		GRPCPort:       config.GetEnv("GRPC_PORT", "9090"),
		EnableCORS:     config.GetEnvAsBool("ENABLE_CORS", true),
		RateLimitRPS:   config.GetEnvAsInt("RATE_LIMIT_RPS", 100),
		RateLimitBurst: config.GetEnvAsInt("RATE_LIMIT_BURST", 200),
		DevMode:        config.GetEnvAsBool("DEV_MODE", false),

		// Service URLs
		BlogServiceURL:         config.GetEnv("BLOG_SERVICE_URL", "localhost:9091"),
		LinkServiceURL:         config.GetEnv("LINK_SERVICE_URL", "localhost:9092"),
		FoodfolioServiceURL:    config.GetEnv("FOODFOLIO_SERVICE_URL", "localhost:9093"),
		NotificationServiceURL: config.GetEnv("NOTIFICATION_SERVICE_URL", "localhost:9094"),
		SSEServiceURL:          config.GetEnv("SSE_SERVICE_URL", "localhost:9095"),
		TwitchBotServiceURL:    config.GetEnv("TWITCHBOT_SERVICE_URL", "localhost:9096"),
		WebhookServiceURL:      config.GetEnv("WEBHOOK_SERVICE_URL", "localhost:9097"),
		WarcraftServiceURL:     config.GetEnv("WARCRAFT_SERVICE_URL", "localhost:9098"),
		WeatherServiceURL:      config.GetEnv("WEATHER_SERVICE_URL", "localhost:9099"),
	}

	return cfg, nil
}
