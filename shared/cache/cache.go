package cache

import (
	"context"
	"time"
)

// Cache is the interface for cache operations
type Cache interface {
	// Get retrieves a value from the cache
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value in the cache with optional TTL
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete removes a value from the cache
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in the cache
	Exists(ctx context.Context, key string) (bool, error)

	// Clear removes all values from the cache
	Clear(ctx context.Context) error

	// Close closes the cache connection
	Close() error
}

// New creates a new cache instance based on the config
func New(config *Config) (Cache, error) {
	if config.Type == "redis" {
		return NewRedisCache(config)
	}
	return NewMemoryCache(config), nil
}
