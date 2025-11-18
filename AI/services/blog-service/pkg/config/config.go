package config

import (
	"time"

	sharedConfig "github.com/toxictoast/toxictoastgo/shared/config"
)

// Config holds blog-service specific configuration
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
	Media MediaConfig

	// Background Jobs
	PostPublisherEnabled  bool
	PostPublisherInterval time.Duration
}

// KafkaConfig extends shared Kafka config with service-specific topics
type KafkaConfig struct {
	sharedConfig.KafkaConfig
	TopicPrefix         string
	TopicPostEvents     string
	TopicCommentEvents  string
	TopicMediaEvents    string
}

// MediaConfig holds media storage configuration
type MediaConfig struct {
	StoragePath          string
	MaxSize              int64
	AllowedTypes         []string
	GenerateThumbnails   bool
	ThumbnailSizes       []string // e.g., "small,medium,large"
	AutoResizeLargeImage bool
	MaxImageWidth        int
	MaxImageHeight       int
}

// Load loads blog-service configuration
func Load() *Config {
	// Load .env file first
	sharedConfig.LoadEnvFile()

	return &Config{
		Port:        sharedConfig.GetEnv("PORT", "8080"),
		GRPCPort:    sharedConfig.GetEnv("GRPC_PORT", "9090"),
		Environment: sharedConfig.GetEnv("ENVIRONMENT", "development"),
		LogLevel:    sharedConfig.GetEnv("LOG_LEVEL", "info"),
		AuthEnabled: sharedConfig.GetEnvAsBool("AUTH_ENABLED", true),
		Database:    sharedConfig.LoadDatabaseConfig(),
		Server:      sharedConfig.LoadServerConfig(),
		Keycloak:    sharedConfig.LoadKeycloakConfig(),
		Kafka: KafkaConfig{
			KafkaConfig:         sharedConfig.LoadKafkaConfig(),
			TopicPrefix:         sharedConfig.GetEnv("KAFKA_TOPIC_PREFIX", "blog"),
			TopicPostEvents:     sharedConfig.GetEnv("KAFKA_TOPIC_POST_EVENTS", "blog.events.post"),
			TopicCommentEvents:  sharedConfig.GetEnv("KAFKA_TOPIC_COMMENT_EVENTS", "blog.events.comment"),
			TopicMediaEvents:    sharedConfig.GetEnv("KAFKA_TOPIC_MEDIA_EVENTS", "blog.events.media"),
		},
		Media: MediaConfig{
			StoragePath:          sharedConfig.GetEnv("MEDIA_STORAGE_PATH", "./uploads"),
			MaxSize:              sharedConfig.GetEnvAsInt64("MEDIA_MAX_SIZE", 10485760),
			AllowedTypes:         sharedConfig.GetEnvAsSlice("MEDIA_ALLOWED_TYPES", "image/jpeg,image/png,image/gif,image/webp"),
			GenerateThumbnails:   sharedConfig.GetEnvAsBool("MEDIA_GENERATE_THUMBNAILS", true),
			ThumbnailSizes:       sharedConfig.GetEnvAsSlice("MEDIA_THUMBNAIL_SIZES", "small,medium,large"),
			AutoResizeLargeImage: sharedConfig.GetEnvAsBool("MEDIA_AUTO_RESIZE_LARGE", false),
			MaxImageWidth:        sharedConfig.GetEnvAsInt("MEDIA_MAX_IMAGE_WIDTH", 3840),
			MaxImageHeight:       sharedConfig.GetEnvAsInt("MEDIA_MAX_IMAGE_HEIGHT", 2160),
		},

		// Background Jobs
		PostPublisherEnabled:  sharedConfig.GetEnvAsBool("POST_PUBLISHER_ENABLED", true),
		PostPublisherInterval: sharedConfig.GetEnvAsDuration("POST_PUBLISHER_INTERVAL", "5m"),
	}
}
