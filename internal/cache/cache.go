package cache

import (
	"sync"
	"time"
)

// Options controls cache behaviour.
type Options struct {
	TTL      time.Duration
	MaxSize  int
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		TTL:     5 * time.Minute,
		MaxSize: 256,
	}
}

type entry struct {
	ports  []uint16
	storeAt time.Time
}

// Cache holds the most-recent scan result for each host so that
// downstream consumers can skip redundant work when the port set
// has not changed between two consecutive scans.
type Cache struct {
	mu      sync.RWMutex
	items   map[string]entry
	opts    Options
	nowFunc func() time.Time
}

// New creates a Cache with the supplied options.
func New(opts Options) *Cache {
	if opts.TTL <= 0 {
		opts.TTL = DefaultOptions().TTL
	}
	if opts.MaxSize <= 0 {
		opts.MaxSize = DefaultOptions().MaxSize
	}
	return &Cache{
		items:   make(map[string]entry, opts.MaxSize),
		opts:    opts,
		nowFunc: time.Now,
	}
}

// Set stores ports for the given host, evicting the oldest entry
// when the cache is full.
func (c *Cache) Set(host string, ports []uint16) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.items) >= c.opts.MaxSize {
		c.evictOldestLocked()
	}
	c.items[host] = entry{ports: ports, storeAt: c.nowFunc()}
}

// Get returns the cached ports for host and whether the entry is
// still valid (present and not expired).
func (c *Cache) Get(host string) ([]uint16, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.items[host]
	if !ok {
		return nil, false
	}
	if c.nowFunc().Sub(e.storeAt) > c.opts.TTL {
		return nil, false
	}
	return e.ports, true
}

// Invalidate removes the entry for host.
func (c *Cache) Invalidate(host string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, host)
}

// Flush removes all cached entries.
func (c *Cache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]entry, c.opts.MaxSize)
}

func (c *Cache) evictOldestLocked() {
	var oldest string
	var oldestTime time.Time
	for host, e := range c.items {
		if oldest == "" || e.storeAt.Before(oldestTime) {
			oldest = host
			oldestTime = e.storeAt
		}
	}
	if oldest != "" {
		delete(c.items, oldest)
	}
}
