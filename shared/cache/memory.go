package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrNotFound = errors.New("key not found")
	ErrMaxSize  = errors.New("cache is full")
)

type cacheItem struct {
	value      []byte
	expiresAt  time.Time
	hasExpiry  bool
	accessedAt time.Time
}

// MemoryCache is an in-memory cache implementation
type MemoryCache struct {
	items   map[string]*cacheItem
	mu      sync.RWMutex
	config  *Config
	cleanup *time.Ticker
	done    chan bool
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(config *Config) *MemoryCache {
	if config == nil {
		config = DefaultConfig()
	}

	mc := &MemoryCache{
		items:  make(map[string]*cacheItem),
		config: config,
		done:   make(chan bool),
	}

	// Start cleanup goroutine
	mc.cleanup = time.NewTicker(1 * time.Minute)
	go mc.cleanupExpired()

	return mc
}

// Get retrieves a value from the cache
func (mc *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	item, exists := mc.items[key]
	if !exists {
		return nil, ErrNotFound
	}

	// Check expiry
	if item.hasExpiry && time.Now().After(item.expiresAt) {
		return nil, ErrNotFound
	}

	// Update access time
	item.accessedAt = time.Now()

	return item.value, nil
}

// Set stores a value in the cache
func (mc *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Check max size
	if mc.config.MaxSize > 0 && len(mc.items) >= mc.config.MaxSize {
		// Evict least recently used
		mc.evictLRU()
	}

	// Use default TTL if not specified
	if ttl == 0 {
		ttl = mc.config.DefaultTTL
	}

	item := &cacheItem{
		value:      value,
		accessedAt: time.Now(),
	}

	if ttl > 0 {
		item.expiresAt = time.Now().Add(ttl)
		item.hasExpiry = true
	}

	mc.items[key] = item
	return nil
}

// Delete removes a value from the cache
func (mc *MemoryCache) Delete(ctx context.Context, key string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	delete(mc.items, key)
	return nil
}

// Exists checks if a key exists in the cache
func (mc *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	item, exists := mc.items[key]
	if !exists {
		return false, nil
	}

	// Check expiry
	if item.hasExpiry && time.Now().After(item.expiresAt) {
		return false, nil
	}

	return true, nil
}

// Clear removes all values from the cache
func (mc *MemoryCache) Clear(ctx context.Context) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.items = make(map[string]*cacheItem)
	return nil
}

// Close closes the cache
func (mc *MemoryCache) Close() error {
	mc.cleanup.Stop()
	mc.done <- true
	return nil
}

// cleanupExpired removes expired items from the cache
func (mc *MemoryCache) cleanupExpired() {
	for {
		select {
		case <-mc.cleanup.C:
			mc.mu.Lock()
			now := time.Now()
			for key, item := range mc.items {
				if item.hasExpiry && now.After(item.expiresAt) {
					delete(mc.items, key)
				}
			}
			mc.mu.Unlock()
		case <-mc.done:
			return
		}
	}
}

// evictLRU removes the least recently used item
func (mc *MemoryCache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range mc.items {
		if oldestKey == "" || item.accessedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.accessedAt
		}
	}

	if oldestKey != "" {
		delete(mc.items, oldestKey)
	}
}
