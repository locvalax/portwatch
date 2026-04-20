package resolve

import (
	"sync"
	"time"
)

// cacheEntry holds a resolved IP and its expiry time.
type cacheEntry struct {
	addrs   []string
	expires time.Time
}

// Cache wraps a Resolver and memoises results for a configurable TTL.
type Cache struct {
	mu      sync.Mutex
	entries map[string]cacheEntry
	ttl     time.Duration
	resolver *Resolver
}

// NewCache returns a Cache that delegates to r and caches results for ttl.
func NewCache(r *Resolver, ttl time.Duration) *Cache {
	return &Cache{
		entries:  make(map[string]cacheEntry),
		ttl:      ttl,
		resolver: r,
	}
}

// Resolve returns cached addresses for host, or delegates to the underlying
// Resolver and stores the result.
func (c *Cache) Resolve(host string) ([]string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if e, ok := c.entries[host]; ok && time.Now().Before(e.expires) {
		return e.addrs, nil
	}

	addrs, err := c.resolver.Resolve(host)
	if err != nil {
		return nil, err
	}

	c.entries[host] = cacheEntry{
		addrs:   addrs,
		expires: time.Now().Add(c.ttl),
	}
	return addrs, nil
}

// Invalidate removes the cached entry for host so the next call re-resolves.
func (c *Cache) Invalidate(host string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, host)
}

// Flush clears all cached entries.
func (c *Cache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]cacheEntry)
}
