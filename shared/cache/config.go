package cache

import "time"

// Config holds the configuration for the cache
type Config struct {
	// Type is the cache type ("redis" or "memory")
	Type string

	// Redis configuration
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// Memory cache configuration
	MaxSize int // Maximum number of items (0 = unlimited)

	// Default TTL for cache entries (0 = no expiration)
	DefaultTTL time.Duration
}

// DefaultConfig returns a config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Type:       "memory",
		MaxSize:    1000,
		DefaultTTL: 5 * time.Minute,
	}
}

// RedisConfig returns a config for Redis cache
func RedisConfig(addr string) *Config {
	return &Config{
		Type:       "redis",
		RedisAddr:  addr,
		DefaultTTL: 5 * time.Minute,
	}
}
