package config

import (
	"os"

	"github.com/joho/godotenv"
	sharedConfig "github.com/toxictoast/toxictoastgo/shared/config"
)

// Config holds warcraft-service specific configuration
type Config struct {
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

	// Embedded shared configs
	Database sharedConfig.DatabaseConfig
	Server   sharedConfig.ServerConfig
	Keycloak sharedConfig.KeycloakConfig
}

func Load() (*Config, error) {
	// Load .env file if exists
	godotenv.Load()

	return &Config{
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
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
