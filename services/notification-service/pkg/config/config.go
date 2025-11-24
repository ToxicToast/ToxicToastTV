package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	sharedConfig "github.com/toxictoast/toxictoastgo/shared/config"
)

// Config holds all configuration for the notification service
type Config struct {
	Environment string
	Port        string
	GRPCPort    string
	Database    sharedConfig.DatabaseConfig
	Server      sharedConfig.ServerConfig
	Kafka       KafkaConfig
	// Background Jobs
	NotificationRetryEnabled          bool
	NotificationRetryInterval         time.Duration
	NotificationRetryMaxRetries       int
	NotificationCleanupEnabled        bool
	NotificationCleanupInterval       time.Duration
	NotificationCleanupRetentionDays  int
}

// KafkaConfig holds Kafka consumer configuration
type KafkaConfig struct {
	Brokers []string
	GroupID string
	Topics  []string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Port:        getEnv("PORT", "8080"),
		GRPCPort:    getEnv("GRPC_PORT", "9096"),
		Database: sharedConfig.DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "azkaban"),
			Password: getEnv("DB_PASSWORD", "Lizzar90"),
			Name:     getEnv("DB_NAME", "azkaban"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Server: sharedConfig.LoadServerConfig(),
		Kafka: KafkaConfig{
			Brokers: getEnvAsSlice("KAFKA_BROKERS", []string{"localhost:19092"}),
			GroupID: getEnv("KAFKA_GROUP_ID", "notification-service"),
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
				"blog.author.created", "blog.author.updated",
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
		},
		// Background Jobs
		NotificationRetryEnabled:          sharedConfig.GetEnvAsBool("NOTIFICATION_RETRY_ENABLED", true),
		NotificationRetryInterval:         sharedConfig.GetEnvAsDuration("NOTIFICATION_RETRY_INTERVAL", "5m"),
		NotificationRetryMaxRetries:       getEnvAsInt("NOTIFICATION_RETRY_MAX_RETRIES", 3),
		NotificationCleanupEnabled:        sharedConfig.GetEnvAsBool("NOTIFICATION_CLEANUP_ENABLED", true),
		NotificationCleanupInterval:       sharedConfig.GetEnvAsDuration("NOTIFICATION_CLEANUP_INTERVAL", "24h"),
		NotificationCleanupRetentionDays:  getEnvAsInt("NOTIFICATION_CLEANUP_RETENTION_DAYS", 30),
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

func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	return strings.Split(valueStr, ",")
}
