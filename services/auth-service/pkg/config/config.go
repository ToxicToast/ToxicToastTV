package config

import (
	"time"

	sharedconfig "github.com/toxictoast/toxictoastgo/shared/config"
)

// Config holds auth-service configuration
type Config struct {
	ServiceName     string
	Port            int
	GRPCPort        int
	Database        sharedconfig.DatabaseConfig
	JWT             JWTConfig
	Kafka           sharedconfig.KafkaConfig
	UserServiceAddr string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey            string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	sharedconfig.LoadEnvFile()

	cfg := &Config{
		ServiceName: sharedconfig.GetEnv("SERVICE_NAME", "auth-service"),
		Port:        sharedconfig.GetEnvAsInt("PORT", 8080),
		GRPCPort:    sharedconfig.GetEnvAsInt("GRPC_PORT", 9090),
		Database:    sharedconfig.LoadDatabaseConfig(),
		JWT: JWTConfig{
			SecretKey:            sharedconfig.GetEnv("JWT_SECRET", "your-secret-key-change-me"),
			AccessTokenDuration:  sharedconfig.GetEnvAsDuration("JWT_ACCESS_DURATION", "15m"),
			RefreshTokenDuration: sharedconfig.GetEnvAsDuration("JWT_REFRESH_DURATION", "168h"),
		},
		Kafka:           sharedconfig.LoadKafkaConfig(),
		UserServiceAddr: sharedconfig.GetEnv("USER_SERVICE_ADDR", "user-service:9090"),
	}

	return cfg, nil
}
