package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// KeycloakConfig holds Keycloak authentication configuration
type KeycloakConfig struct {
	URL          string
	Realm        string
	ClientID     string
	ClientSecret string
	PublicKey    string
}

// KafkaConfig holds Kafka/Redpanda configuration
type KafkaConfig struct {
	Brokers []string
	GroupID string
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	Name         string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
}

// ServerConfig holds HTTP/gRPC server configuration
type ServerConfig struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// GetDatabaseURL returns PostgreSQL connection string
func (c *DatabaseConfig) GetDatabaseURL() string {
	return "host=" + c.Host + " port=" + c.Port + " user=" + c.User +
		" password=" + c.Password + " dbname=" + c.Name + " sslmode=" + c.SSLMode
}

// LoadEnvFile loads .env file if it exists
func LoadEnvFile() {
	err := godotenv.Load()
	if err != nil {
		// .env file is optional, don't fail if it doesn't exist
		log.Printf("Info: No .env file found (this is optional)")
	} else {
		log.Printf("Loaded .env file")
	}
}

// Helper functions for environment variable loading

func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Warning: Invalid integer value for %s: %s, using default: %d", key, value, defaultValue)
	}
	return defaultValue
}

func GetEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
		log.Printf("Warning: Invalid int64 value for %s: %s, using default: %d", key, value, defaultValue)
	}
	return defaultValue
}

func GetEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
		log.Printf("Warning: Invalid bool value for %s: %s, using default: %t", key, value, defaultValue)
	}
	return defaultValue
}

func GetEnvAsDuration(key string, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		log.Printf("Warning: Invalid duration value for %s: %s, using default: %s", key, value, defaultValue)
	}
	duration, _ := time.ParseDuration(defaultValue)
	return duration
}

func GetEnvAsSlice(key string, defaultValue string) []string {
	value := GetEnv(key, defaultValue)
	if value == "" {
		return []string{}
	}

	parts := []string{}
	current := ""
	for i := 0; i < len(value); i++ {
		if value[i] == ',' {
			trimmed := trimSpace(current)
			if trimmed != "" {
				parts = append(parts, trimmed)
			}
			current = ""
		} else {
			current += string(value[i])
		}
	}
	trimmed := trimSpace(current)
	if trimmed != "" {
		parts = append(parts, trimmed)
	}

	return parts
}

func trimSpace(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}

// LoadKeycloakConfig loads Keycloak configuration from environment
func LoadKeycloakConfig() KeycloakConfig {
	return KeycloakConfig{
		URL:          GetEnv("KEYCLOAK_URL", "http://localhost:8080"),
		Realm:        GetEnv("KEYCLOAK_REALM", ""),
		ClientID:     GetEnv("KEYCLOAK_CLIENT_ID", ""),
		ClientSecret: GetEnv("KEYCLOAK_CLIENT_SECRET", ""),
		PublicKey:    GetEnv("KEYCLOAK_PUBLIC_KEY", ""),
	}
}

// LoadKafkaConfig loads Kafka configuration from environment
func LoadKafkaConfig() KafkaConfig {
	return KafkaConfig{
		Brokers: GetEnvAsSlice("KAFKA_BROKERS", "localhost:9092"),
		GroupID: GetEnv("KAFKA_GROUP_ID", "default-group"),
	}
}

// LoadDatabaseConfig loads database configuration from environment
func LoadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:         GetEnv("DB_HOST", "localhost"),
		Port:         GetEnv("DB_PORT", "5432"),
		User:         GetEnv("DB_USER", "postgres"),
		Password:     GetEnv("DB_PASSWORD", ""),
		Name:         GetEnv("DB_NAME", ""),
		SSLMode:      GetEnv("DB_SSL_MODE", "disable"),
		MaxOpenConns: GetEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns: GetEnvAsInt("DB_MAX_IDLE_CONNS", 25),
		MaxLifetime:  GetEnvAsDuration("DB_MAX_LIFETIME", "5m"),
	}
}

// LoadServerConfig loads server configuration from environment
func LoadServerConfig() ServerConfig {
	return ServerConfig{
		ReadTimeout:  GetEnvAsDuration("SERVER_READ_TIMEOUT", "10s"),
		WriteTimeout: GetEnvAsDuration("SERVER_WRITE_TIMEOUT", "10s"),
		IdleTimeout:  GetEnvAsDuration("SERVER_IDLE_TIMEOUT", "60s"),
	}
}
