package cache

import (
	"context"

	"github.com/user/portwatch/internal/scanner"
)

// CachedScanner wraps a scanner.Scanner and skips the underlying
// scan when a fresh result is already held in the cache.
type CachedScanner struct {
	inner scanner.Scanner
	cache *Cache
}

// NewCachedScanner returns a CachedScanner backed by the given cache.
func NewCachedScanner(inner scanner.Scanner, c *Cache) *CachedScanner {
	return &CachedScanner{inner: inner, cache: c}
}

// Scan returns cached ports when available; otherwise it delegates to
// the underlying scanner and stores the result before returning.
func (cs *CachedScanner) Scan(ctx context.Context, host string, opts scanner.Options) ([]uint16, error) {
	if ports, ok := cs.cache.Get(host); ok {
		return ports, nil
	}
	ports, err := cs.inner.Scan(ctx, host, opts)
	if err != nil {
		return nil, err
	}
	cs.cache.Set(host, ports)
	return ports, nil
}
