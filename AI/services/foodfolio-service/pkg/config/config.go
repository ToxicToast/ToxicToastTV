package config

import (
	"time"

	sharedConfig "github.com/toxictoast/toxictoastgo/shared/config"
)

// Config holds foodfolio-service specific configuration
type Config struct {
	Port        string
	GRPCPort    string
	Environment string
	LogLevel    string
	AuthEnabled bool

	// Embedded shared configs
	Database sharedConfig.DatabaseConfig
	Server   sharedConfig.ServerConfig
	Keycloak sharedConfig.KeycloakConfig
	Kafka    KafkaConfig

	// Service-specific config
	OCR OCRConfig

	// Background Jobs
	ItemExpirationEnabled   bool
	ItemExpirationInterval  time.Duration
	StockLevelEnabled       bool
	StockLevelInterval      time.Duration
}

// KafkaConfig extends shared Kafka config with service-specific topics
type KafkaConfig struct {
	sharedConfig.KafkaConfig
	TopicPrefix          string
	TopicItemEvents      string
	TopicInventoryEvents string
	TopicReceiptEvents   string
	TopicAlertEvents     string
}

// OCRConfig holds OCR/Receipt scanning configuration
type OCRConfig struct {
	StoragePath  string
	MaxSize      int64
	AllowedTypes []string
	OCRProvider  string // "tesseract", "google-vision", etc.
}

// Load loads foodfolio-service configuration
func Load() *Config {
	// Load .env file first
	sharedConfig.LoadEnvFile()

	return &Config{
		Port:        sharedConfig.GetEnv("PORT", "8081"),
		GRPCPort:    sharedConfig.GetEnv("GRPC_PORT", "9091"),
		Environment: sharedConfig.GetEnv("ENVIRONMENT", "development"),
		LogLevel:    sharedConfig.GetEnv("LOG_LEVEL", "info"),
		AuthEnabled: sharedConfig.GetEnvAsBool("AUTH_ENABLED", true),
		Database:    sharedConfig.LoadDatabaseConfig(),
		Server:      sharedConfig.LoadServerConfig(),
		Keycloak:    sharedConfig.LoadKeycloakConfig(),
		Kafka: KafkaConfig{
			KafkaConfig:          sharedConfig.LoadKafkaConfig(),
			TopicPrefix:          sharedConfig.GetEnv("KAFKA_TOPIC_PREFIX", "foodfolio"),
			TopicItemEvents:      sharedConfig.GetEnv("KAFKA_TOPIC_ITEM_EVENTS", "foodfolio.events.item"),
			TopicInventoryEvents: sharedConfig.GetEnv("KAFKA_TOPIC_INVENTORY_EVENTS", "foodfolio.events.inventory"),
			TopicReceiptEvents:   sharedConfig.GetEnv("KAFKA_TOPIC_RECEIPT_EVENTS", "foodfolio.events.receipt"),
			TopicAlertEvents:     sharedConfig.GetEnv("KAFKA_TOPIC_ALERT_EVENTS", "foodfolio.events.alert"),
		},
		OCR: OCRConfig{
			StoragePath:  sharedConfig.GetEnv("OCR_STORAGE_PATH", "./receipts"),
			MaxSize:      sharedConfig.GetEnvAsInt64("OCR_MAX_SIZE", 10485760), // 10MB
			AllowedTypes: sharedConfig.GetEnvAsSlice("OCR_ALLOWED_TYPES", "image/jpeg,image/png,application/pdf"),
			OCRProvider:  sharedConfig.GetEnv("OCR_PROVIDER", "tesseract"),
		},

		// Background Jobs
		ItemExpirationEnabled:  sharedConfig.GetEnvAsBool("ITEM_EXPIRATION_ENABLED", true),
		ItemExpirationInterval: sharedConfig.GetEnvAsDuration("ITEM_EXPIRATION_INTERVAL", "24h"),
		StockLevelEnabled:      sharedConfig.GetEnvAsBool("STOCK_LEVEL_ENABLED", true),
		StockLevelInterval:     sharedConfig.GetEnvAsDuration("STOCK_LEVEL_INTERVAL", "6h"),
	}
}
