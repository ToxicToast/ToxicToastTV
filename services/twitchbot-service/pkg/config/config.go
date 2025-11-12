package config

import (
	sharedConfig "github.com/toxictoast/toxictoastgo/shared/config"
)

// Config holds twitchbot-service specific configuration
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
	Twitch TwitchConfig
}

// KafkaConfig extends shared Kafka config
type KafkaConfig struct {
	sharedConfig.KafkaConfig
}

// TwitchConfig holds Twitch API and IRC configuration
type TwitchConfig struct {
	Channel      string
	ClientID     string
	ClientSecret string
	AccessToken  string
	BotUsername  string
	IRCServer    string
	IRCPort      string
	IRCDebug     bool
}

// Load loads twitchbot-service configuration
func Load() *Config {
	// Load .env file first
	sharedConfig.LoadEnvFile()

	return &Config{
		Port:        sharedConfig.GetEnv("PORT", "8083"),
		GRPCPort:    sharedConfig.GetEnv("GRPC_PORT", "9093"),
		Environment: sharedConfig.GetEnv("ENVIRONMENT", "development"),
		LogLevel:    sharedConfig.GetEnv("LOG_LEVEL", "info"),
		AuthEnabled: sharedConfig.GetEnvAsBool("AUTH_ENABLED", false),
		Database:    sharedConfig.LoadDatabaseConfig(),
		Server:      sharedConfig.LoadServerConfig(),
		Keycloak:    sharedConfig.LoadKeycloakConfig(),
		Kafka: KafkaConfig{
			KafkaConfig: sharedConfig.LoadKafkaConfig(),
		},
		Twitch: TwitchConfig{
			Channel:      sharedConfig.GetEnv("TWITCH_CHANNEL", ""),
			ClientID:     sharedConfig.GetEnv("TWITCH_CLIENT_ID", ""),
			ClientSecret: sharedConfig.GetEnv("TWITCH_CLIENT_SECRET", ""),
			AccessToken:  sharedConfig.GetEnv("TWITCH_ACCESS_TOKEN", ""),
			BotUsername:  sharedConfig.GetEnv("TWITCH_BOT_USERNAME", ""),
			IRCServer:    sharedConfig.GetEnv("TWITCH_IRC_SERVER", "irc.chat.twitch.tv"),
			IRCPort:      sharedConfig.GetEnv("TWITCH_IRC_PORT", "6667"),
			IRCDebug:     sharedConfig.GetEnvAsBool("TWITCH_IRC_DEBUG", false),
		},
	}
}
