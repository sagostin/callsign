package xmlcache

import (
	"strings"
	"sync"
	"time"
)

// CacheItem represents a cached XML document with expiration
type CacheItem struct {
	Value     string
	ExpiresAt time.Time
}

// IsExpired checks if the cache item has expired
func (c *CacheItem) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// XMLCache provides an in-memory cache for FreeSWITCH XML configurations
type XMLCache struct {
	items map[string]*CacheItem
	mu    sync.RWMutex
}

// New creates a new XMLCache instance
func New() *XMLCache {
	cache := &XMLCache{
		items: make(map[string]*CacheItem),
	}

	// Start background cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves an item from the cache
func (c *XMLCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return "", false
	}

	if item.IsExpired() {
		return "", false
	}

	return item.Value, true
}

// Set stores an item in the cache with a TTL
func (c *XMLCache) Set(key string, value string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &CacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Delete removes a specific item from the cache
func (c *XMLCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// DeleteByPrefix removes all items matching a key prefix
func (c *XMLCache) DeleteByPrefix(prefix string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := 0
	for key := range c.items {
		if strings.HasPrefix(key, prefix) {
			delete(c.items, key)
			count++
		}
	}
	return count
}

// DeleteByPattern removes all items matching a pattern (simple wildcard support)
func (c *XMLCache) DeleteByPattern(pattern string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := 0
	for key := range c.items {
		if matchPattern(pattern, key) {
			delete(c.items, key)
			count++
		}
	}
	return count
}

// Flush clears all items from the cache
func (c *XMLCache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*CacheItem)
}

// Stats returns cache statistics
func (c *XMLCache) Stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := len(c.items)
	expired := 0
	now := time.Now()

	for _, item := range c.items {
		if now.After(item.ExpiresAt) {
			expired++
		}
	}

	return map[string]interface{}{
		"total_items":   total,
		"expired_items": expired,
		"active_items":  total - expired,
	}
}

// cleanup periodically removes expired items
func (c *XMLCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.After(item.ExpiresAt) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

// matchPattern provides simple wildcard matching (* matches any substring)
func matchPattern(pattern, str string) bool {
	if pattern == "*" {
		return true
	}

	parts := strings.Split(pattern, "*")
	if len(parts) == 1 {
		return pattern == str
	}

	// Check prefix
	if parts[0] != "" && !strings.HasPrefix(str, parts[0]) {
		return false
	}

	// Check suffix
	if parts[len(parts)-1] != "" && !strings.HasSuffix(str, parts[len(parts)-1]) {
		return false
	}

	// Check middle parts
	pos := len(parts[0])
	for i := 1; i < len(parts)-1; i++ {
		idx := strings.Index(str[pos:], parts[i])
		if idx < 0 {
			return false
		}
		pos += idx + len(parts[i])
	}

	return true
}

// Common cache key builders
const (
	PrefixConfiguration = "config:"
	PrefixDirectory     = "directory:"
	PrefixDialplan      = "dialplan:"
)

// ConfigKey generates a cache key for configuration items
func ConfigKey(hostname, configName string) string {
	return PrefixConfiguration + hostname + ":" + configName
}

// DirectoryKey generates a cache key for directory items
func DirectoryKey(domain, user string) string {
	return PrefixDirectory + domain + ":" + user
}

// DialplanKey generates a cache key for dialplan items
func DialplanKey(context string) string {
	return PrefixDialplan + context
}

// DialplanSingleKey generates a cache key for single-mode dialplan lookup
func DialplanSingleKey(context, destination string) string {
	return PrefixDialplan + context + ":" + destination
}
