package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	sharedConfig "github.com/toxictoast/toxictoastgo/shared/config"
)

// Config holds all configuration for the SSE service
type Config struct {
	Environment string
	Port        string // HTTP port for SSE
	GRPCPort    string
	Database    sharedConfig.DatabaseConfig
	Kafka       KafkaConfig
	Keycloak    sharedConfig.KeycloakConfig
	AuthEnabled bool
	Server      ServerConfig
	SSE         SSEConfig
	CORS        CORSConfig
	RateLimit   RateLimitConfig
}

// KafkaConfig holds Kafka consumer configuration
type KafkaConfig struct {
	Brokers         []string
	GroupID         string
	Topics          []string // Topics to subscribe to
	AutoOffsetReset string
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// SSEConfig holds SSE-specific configuration
type SSEConfig struct {
	MaxClients       int
	HeartbeatSeconds int
	EventBufferSize  int
	HistorySize      int // Number of recent events to keep for replay
	// Background Jobs
	ClientCleanupEnabled        bool
	ClientCleanupInterval       time.Duration
	ClientCleanupInactiveTimeout time.Duration
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled        bool
	RequestsPerMin int // Max requests per IP per minute
	BurstSize      int // Burst size for rate limiter
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Port:        getEnv("PORT", "8084"),
		GRPCPort:    getEnv("GRPC_PORT", "9094"),
		Database: sharedConfig.DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "sse_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Kafka: KafkaConfig{
			Brokers: getEnvAsSlice("KAFKA_BROKERS", []string{"localhost:19092"}),
			GroupID: getEnv("KAFKA_GROUP_ID", "sse-service"),
			Topics: getEnvAsSlice("KAFKA_TOPICS", []string{
				// User Service Topics (6)
				"user.created", "user.updated", "user.deleted",
				"user.activated", "user.deactivated", "user.password.changed",
				// Auth Service Topics (3)
				"auth.registered", "auth.login", "auth.token.refreshed",
				// Blog Service Topics (19)
				"blog.post.created", "blog.post.updated", "blog.post.published", "blog.post.deleted",
				"blog.category.created", "blog.category.updated", "blog.category.deleted",
				"blog.tag.created", "blog.tag.updated", "blog.tag.deleted",
				"blog.comment.created", "blog.comment.approved", "blog.comment.rejected", "blog.comment.deleted",
				"blog.media.uploaded", "blog.media.deleted", "blog.media.thumbnail.generated",
				// Twitchbot Service Topics (21)
				"twitchbot.stream.started", "twitchbot.stream.ended", "twitchbot.stream.updated",
				"twitchbot.message.received", "twitchbot.message.deleted", "twitchbot.message.timeout",
				"twitchbot.viewer.joined", "twitchbot.viewer.left", "twitchbot.viewer.banned", "twitchbot.viewer.unbanned",
				"twitchbot.viewer.mod.added", "twitchbot.viewer.mod.removed", "twitchbot.viewer.vip.added", "twitchbot.viewer.vip.removed",
				"twitchbot.clip.created", "twitchbot.clip.updated", "twitchbot.clip.deleted",
				"twitchbot.command.created", "twitchbot.command.updated", "twitchbot.command.deleted", "twitchbot.command.executed",
				// Link Service Topics (8)
				"link.created", "link.updated", "link.deleted", "link.expired",
				"link.activated", "link.deactivated", "link.clicked", "link.click.fraud.detected",
				// Foodfolio Service Topics (36)
				"foodfolio.category.created", "foodfolio.category.updated", "foodfolio.category.deleted",
				"foodfolio.company.created", "foodfolio.company.updated", "foodfolio.company.deleted",
				"foodfolio.item.created", "foodfolio.item.updated", "foodfolio.item.deleted",
				"foodfolio.variant.created", "foodfolio.variant.updated", "foodfolio.variant.deleted",
				"foodfolio.variant.stock.low", "foodfolio.variant.stock.empty",
				"foodfolio.detail.created", "foodfolio.detail.opened", "foodfolio.detail.expired",
				"foodfolio.detail.expiring.soon", "foodfolio.detail.consumed", "foodfolio.detail.moved",
				"foodfolio.detail.frozen", "foodfolio.detail.thawed",
				"foodfolio.location.created", "foodfolio.location.updated", "foodfolio.location.deleted",
				"foodfolio.warehouse.created", "foodfolio.warehouse.updated", "foodfolio.warehouse.deleted",
				"foodfolio.receipt.created", "foodfolio.receipt.scanned", "foodfolio.receipt.deleted",
				"foodfolio.shoppinglist.created", "foodfolio.shoppinglist.updated", "foodfolio.shoppinglist.deleted",
				"foodfolio.shoppinglist.item.added", "foodfolio.shoppinglist.item.removed", "foodfolio.shoppinglist.item.purchased",
			}),
			AutoOffsetReset: getEnv("KAFKA_AUTO_OFFSET_RESET", "latest"),
		},
		Keycloak: sharedConfig.KeycloakConfig{
			URL:      getEnv("KEYCLOAK_URL", "http://localhost:8080"),
			Realm:    getEnv("KEYCLOAK_REALM", "toxictoast"),
			ClientID: getEnv("KEYCLOAK_CLIENT_ID", "sse-service"),
		},
		AuthEnabled: getEnvAsBool("AUTH_ENABLED", false),
		Server: ServerConfig{
			ReadTimeout:  time.Duration(getEnvAsInt("SERVER_READ_TIMEOUT", 15)) * time.Second,
			WriteTimeout: time.Duration(getEnvAsInt("SERVER_WRITE_TIMEOUT", 15)) * time.Second,
			IdleTimeout:  time.Duration(getEnvAsInt("SERVER_IDLE_TIMEOUT", 60)) * time.Second,
		},
		SSE: SSEConfig{
			MaxClients:                   getEnvAsInt("SSE_MAX_CLIENTS", 1000),
			HeartbeatSeconds:             getEnvAsInt("SSE_HEARTBEAT_SECONDS", 30),
			EventBufferSize:              getEnvAsInt("SSE_EVENT_BUFFER_SIZE", 100),
			HistorySize:                  getEnvAsInt("SSE_HISTORY_SIZE", 100),
			// Background Jobs
			ClientCleanupEnabled:         sharedConfig.GetEnvAsBool("SSE_CLIENT_CLEANUP_ENABLED", true),
			ClientCleanupInterval:        sharedConfig.GetEnvAsDuration("SSE_CLIENT_CLEANUP_INTERVAL", "5m"),
			ClientCleanupInactiveTimeout: sharedConfig.GetEnvAsDuration("SSE_CLIENT_CLEANUP_INACTIVE_TIMEOUT", "30m"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
			AllowedMethods: getEnvAsSlice("CORS_ALLOWED_METHODS", []string{"GET", "OPTIONS"}),
			AllowedHeaders: getEnvAsSlice("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Last-Event-ID"}),
		},
		RateLimit: RateLimitConfig{
			Enabled:        getEnvAsBool("RATE_LIMIT_ENABLED", true),
			RequestsPerMin: getEnvAsInt("RATE_LIMIT_REQUESTS_PER_MIN", 60),
			BurstSize:      getEnvAsInt("RATE_LIMIT_BURST_SIZE", 10),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	return strings.Split(valueStr, ",")
}
