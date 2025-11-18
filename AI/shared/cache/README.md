# Cache

A flexible caching library with Redis and in-memory implementations.

## Features

- ✅ Redis cache support
- ✅ In-memory cache fallback
- ✅ TTL (Time-To-Live) support
- ✅ LRU eviction for memory cache
- ✅ Automatic cleanup of expired items
- ✅ Thread-safe operations
- ✅ Context support

## Usage

### In-Memory Cache

```go
import "github.com/toxictoast/toxictoastgo/shared/cache"

// Create in-memory cache with defaults
config := cache.DefaultConfig()
c, err := cache.New(config)
if err != nil {
    log.Fatal(err)
}
defer c.Close()

// Store value with 5 minute TTL
err = c.Set(context.Background(), "user:123", []byte("John Doe"), 5*time.Minute)

// Retrieve value
value, err := c.Get(context.Background(), "user:123")

// Check if exists
exists, err := c.Exists(context.Background(), "user:123")

// Delete
err = c.Delete(context.Background(), "user:123")
```

### Redis Cache

```go
// Create Redis cache
config := cache.RedisConfig("localhost:6379")
config.RedisPassword = "your-password"
config.DefaultTTL = 10 * time.Minute

c, err := cache.New(config)
if err != nil {
    log.Fatal(err)
}
defer c.Close()

// Same API as in-memory cache
err = c.Set(ctx, "key", []byte("value"), time.Hour)
```

### Custom Configuration

```go
config := &cache.Config{
    Type:       "memory",
    MaxSize:    5000,           // Max items in memory
    DefaultTTL: 15 * time.Minute,
}

// Or for Redis
config := &cache.Config{
    Type:          "redis",
    RedisAddr:     "localhost:6379",
    RedisPassword: "secret",
    RedisDB:       0,
    DefaultTTL:    10 * time.Minute,
}
```

## Error Handling

```go
value, err := c.Get(ctx, "unknown-key")
if err == cache.ErrNotFound {
    // Key doesn't exist or expired
}
```

## Features

### TTL Support
- Automatic expiration of cache entries
- Custom TTL per entry or use default
- Background cleanup of expired items

### LRU Eviction (Memory Cache)
- Automatically evicts least recently used items when max size reached
- Configurable maximum cache size

### Thread Safety
- All operations are thread-safe
- Concurrent reads and writes supported

## Performance

### Memory Cache
- O(1) Get/Set/Delete operations
- Automatic cleanup every minute
- LRU eviction when max size reached

### Redis Cache
- Network latency applies
- Supports Redis clustering (future)
- Persistent storage
