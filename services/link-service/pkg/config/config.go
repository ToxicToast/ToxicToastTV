package config

import (
	"time"

	sharedConfig "github.com/toxictoast/toxictoastgo/shared/config"
)

// Config holds link-service specific configuration
type Config struct {
	Port        string
	GRPCPort    string
	Environment string
	LogLevel    string
	AuthEnabled bool
	BaseURL     string // Base URL for generating short URLs (e.g., https://short.link)

	// Embedded shared configs
	Database sharedConfig.DatabaseConfig
	Server   sharedConfig.ServerConfig
	Keycloak sharedConfig.KeycloakConfig
	Kafka    KafkaConfig

	// Service-specific config
	Redis RedisConfig

	// Background Jobs
	LinkExpirationEnabled  bool
	LinkExpirationInterval time.Duration
}

// KafkaConfig extends shared Kafka config with service-specific topics
type KafkaConfig struct {
	sharedConfig.KafkaConfig
	TopicPrefix      string
	TopicLinkEvents  string
	TopicClickEvents string
}

// RedisConfig holds Redis cache configuration
type RedisConfig struct {
	Enabled  bool
	Host     string
	Port     string
	Password string
	DB       int
	TTL      int // Cache TTL in seconds
}

// Load loads link-service configuration
func Load() *Config {
	// Load .env file first
	sharedConfig.LoadEnvFile()

	return &Config{
		Port:        sharedConfig.GetEnv("PORT", "8080"),
		GRPCPort:    sharedConfig.GetEnv("GRPC_PORT", "9090"),
		Environment: sharedConfig.GetEnv("ENVIRONMENT", "development"),
		LogLevel:    sharedConfig.GetEnv("LOG_LEVEL", "info"),
		AuthEnabled: sharedConfig.GetEnvAsBool("AUTH_ENABLED", true),
		BaseURL:     sharedConfig.GetEnv("BASE_URL", "http://localhost:8080"),
		Database:    sharedConfig.LoadDatabaseConfig(),
		Server:      sharedConfig.LoadServerConfig(),
		Keycloak:    sharedConfig.LoadKeycloakConfig(),
		Kafka: KafkaConfig{
			KafkaConfig:      sharedConfig.LoadKafkaConfig(),
			TopicPrefix:      sharedConfig.GetEnv("KAFKA_TOPIC_PREFIX", "link"),
			TopicLinkEvents:  sharedConfig.GetEnv("KAFKA_TOPIC_LINK_EVENTS", "link.events.link"),
			TopicClickEvents: sharedConfig.GetEnv("KAFKA_TOPIC_CLICK_EVENTS", "link.events.click"),
		},
		Redis: RedisConfig{
			Enabled:  sharedConfig.GetEnvAsBool("REDIS_ENABLED", false),
			Host:     sharedConfig.GetEnv("REDIS_HOST", "localhost"),
			Port:     sharedConfig.GetEnv("REDIS_PORT", "6379"),
			Password: sharedConfig.GetEnv("REDIS_PASSWORD", ""),
			DB:       sharedConfig.GetEnvAsInt("REDIS_DB", 0),
			TTL:      sharedConfig.GetEnvAsInt("REDIS_TTL", 3600),
		},

		// Background Jobs
		LinkExpirationEnabled:  sharedConfig.GetEnvAsBool("LINK_EXPIRATION_ENABLED", true),
		LinkExpirationInterval: sharedConfig.GetEnvAsDuration("LINK_EXPIRATION_INTERVAL", "1h"),
	}
}
