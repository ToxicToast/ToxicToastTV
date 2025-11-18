package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache is a Redis cache implementation
type RedisCache struct {
	client *redis.Client
	config *Config
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(config *Config) (*RedisCache, error) {
	if config == nil {
		config = RedisConfig("localhost:6379")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisCache{
		client: client,
		config: config,
	}, nil
}

// Get retrieves a value from the cache
func (rc *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := rc.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, ErrNotFound
	}
	return val, err
}

// Set stores a value in the cache
func (rc *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	// Use default TTL if not specified
	if ttl == 0 {
		ttl = rc.config.DefaultTTL
	}

	return rc.client.Set(ctx, key, value, ttl).Err()
}

// Delete removes a value from the cache
func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	return rc.client.Del(ctx, key).Err()
}

// Exists checks if a key exists in the cache
func (rc *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := rc.client.Exists(ctx, key).Result()
	return count > 0, err
}

// Clear removes all values from the cache
func (rc *RedisCache) Clear(ctx context.Context) error {
	return rc.client.FlushDB(ctx).Err()
}

// Close closes the Redis connection
func (rc *RedisCache) Close() error {
	return rc.client.Close()
}
