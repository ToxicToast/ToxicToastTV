package config

import (
	sharedconfig "github.com/toxictoast/toxictoastgo/shared/config"
)

// Config holds user-service configuration
type Config struct {
	ServiceName string
	Port        int
	GRPCPort    int
	Database    sharedconfig.DatabaseConfig
	Kafka       sharedconfig.KafkaConfig
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	sharedconfig.LoadEnvFile()

	cfg := &Config{
		ServiceName: sharedconfig.GetEnv("SERVICE_NAME", "user-service"),
		Port:        sharedconfig.GetEnvAsInt("PORT", 8080),
		GRPCPort:    sharedconfig.GetEnvAsInt("GRPC_PORT", 9090),
		Database:    sharedconfig.LoadDatabaseConfig(),
		Kafka:       sharedconfig.LoadKafkaConfig(),
	}

	return cfg, nil
}
