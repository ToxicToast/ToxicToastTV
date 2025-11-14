package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	sharedConfig "github.com/toxictoast/toxictoastgo/shared/config"
)

// Config holds warcraft-service specific configuration
type Config struct {
	Port        string
	GRPCPort    string
	Environment string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Blizzard API
	BlizzardClientID     string
	BlizzardClientSecret string
	BlizzardRegion       string

	// Kafka
	KafkaBrokers []string

	// Background Jobs
	CharacterSyncEnabled  bool
	CharacterSyncInterval time.Duration
	GuildSyncEnabled      bool
	GuildSyncInterval     time.Duration

	// Embedded shared configs
	Database sharedConfig.DatabaseConfig
	Server   sharedConfig.ServerConfig
	Keycloak sharedConfig.KeycloakConfig
}

func Load() (*Config, error) {
	// Load .env file if exists
	godotenv.Load()

	// Load shared configs
	databaseCfg := sharedConfig.LoadDatabaseConfig()
	serverCfg := sharedConfig.LoadServerConfig()
	keycloakCfg := sharedConfig.LoadKeycloakConfig()

	return &Config{
		Port:        getEnv("PORT", "8080"),
		GRPCPort:    getEnv("GRPC_PORT", "9090"),
		Environment: getEnv("ENVIRONMENT", "development"),

		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "warcraft_db"),

		// Blizzard API
		BlizzardClientID:     getEnv("BLIZZARD_CLIENT_ID", ""),
		BlizzardClientSecret: getEnv("BLIZZARD_CLIENT_SECRET", ""),
		BlizzardRegion:       getEnv("BLIZZARD_REGION", "us"),

		// Kafka
		KafkaBrokers: getEnvAsSlice("KAFKA_BROKERS", []string{"localhost:19092"}),

		// Background Jobs
		CharacterSyncEnabled:  getEnvAsBool("CHARACTER_SYNC_ENABLED", true),
		CharacterSyncInterval: getEnvAsDuration("CHARACTER_SYNC_INTERVAL", 6*time.Hour),
		GuildSyncEnabled:      getEnvAsBool("GUILD_SYNC_ENABLED", true),
		GuildSyncInterval:     getEnvAsDuration("GUILD_SYNC_INTERVAL", 12*time.Hour),

		// Embedded shared configs
		Database: databaseCfg,
		Server:   serverCfg,
		Keycloak: keycloakCfg,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
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

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	// Try parsing as duration string (e.g., "6h", "30m")
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		// Try parsing as hours (e.g., "6" = 6 hours)
		hours, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return defaultValue
		}
		return time.Duration(hours * float64(time.Hour))
	}
	return value
}
