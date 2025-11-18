package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServiceName string
	Environment string
	Port        string
	GRPCPort    string
}

func Load() *Config {
	// Try to load .env file (optional)
	if err := godotenv.Load(); err != nil {
		log.Println("Info: No .env file found (this is optional)")
	}

	cfg := &Config{
		ServiceName: getEnv("SERVICE_NAME", "weather-service"),
		Environment: getEnv("ENVIRONMENT", "development"),
		Port:        getEnv("PORT", "8080"),
		GRPCPort:    getEnv("GRPC_PORT", "9090"),
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
