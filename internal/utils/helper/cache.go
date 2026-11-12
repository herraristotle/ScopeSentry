// Package helper -----------------------------
// @file      : cache.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/1/16
// -------------------------------------------
package helper

import (
	"sync"
	"time"
)

// Cache is a generic cache structure with expiration
type Cache struct {
	data      map[string]interface{} // Cached data
	expiresAt map[string]time.Time   // Expiration times
	mu        sync.RWMutex           // Read/write lock
	ttl       time.Duration          // Default cache expiration time
}

// NewCache creates a new cache instance
// ttl: cache expiration time; if 0, the default value is used
func NewCache(defaultTTL time.Duration) *Cache {
	if defaultTTL <= 0 {
		defaultTTL = 30 * time.Second // Default 30 seconds
	}
	return &Cache{
		data:      make(map[string]interface{}),
		expiresAt: make(map[string]time.Time),
		ttl:       defaultTTL,
	}
}

// Get fetches data from the cache; returns false if expired or missing
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	value, exists := c.data[key]
	expireTime, hasExpire := c.expiresAt[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	if !hasExpire || time.Now().After(expireTime) {
		// Cache expired, delete it
		c.mu.Lock()
		// Double-check to prevent concurrent deletion
		if et, ok := c.expiresAt[key]; ok && (time.Now().After(et) || !hasExpire) {
			delete(c.data, key)
			delete(c.expiresAt, key)
		}
		c.mu.Unlock()
		return nil, false
	}

	return value, true
}

// Set stores cache data using the default TTL
func (c *Cache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.ttl)
}

// SetWithTTL stores cache data using a custom TTL
func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	if ttl <= 0 {
		ttl = c.ttl
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = value
	c.expiresAt[key] = time.Now().Add(ttl)
}

// Delete removes a cache entry
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
	delete(c.expiresAt, key)
}

// Clear empties all cache entries
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]interface{})
	c.expiresAt = make(map[string]time.Time)
}

// CleanExpired removes expired cache entries
func (c *Cache) CleanExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, expireTime := range c.expiresAt {
		if now.After(expireTime) {
			delete(c.data, key)
			delete(c.expiresAt, key)
		}
	}
}

// Size returns the current number of cache entries
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}

// GetStringSlice fetches a string-slice cache value
func (c *Cache) GetStringSlice(key string) ([]string, bool) {
	value, ok := c.Get(key)
	if !ok {
		return nil, false
	}

	if slice, ok := value.([]string); ok {
		return slice, true
	}
	return nil, false
}

// SetStringSlice stores a string-slice cache value
func (c *Cache) SetStringSlice(key string, value []string) {
	c.Set(key, value)
}

// SetStringSliceWithTTL stores a string-slice cache value using a custom TTL
func (c *Cache) SetStringSliceWithTTL(key string, value []string, ttl time.Duration) {
	c.SetWithTTL(key, value, ttl)
}
