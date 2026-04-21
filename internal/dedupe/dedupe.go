// Package dedupe provides a deduplication layer that suppresses repeated
// scan results when the open port set has not changed since the last scan.
package dedupe

import (
	"sync"

	"github.com/user/portwatch/internal/store"
)

// Fingerprint is a compact representation of a port set.
type Fingerprint = string

// Cache tracks the last-seen port fingerprint per host.
type Cache struct {
	mu      sync.Mutex
	records map[string]Fingerprint
}

// New returns an initialised Cache.
func New() *Cache {
	return &Cache{records: make(map[string]Fingerprint)}
}

// IsDuplicate reports whether the entry carries the same open-port set as the
// previous entry recorded for the same host. If it is not a duplicate the
// cache is updated so the current entry becomes the new baseline.
func (c *Cache) IsDuplicate(e store.Entry) bool {
	fp := fingerprint(e.Ports)

	c.mu.Lock()
	defer c.mu.Unlock()

	prev, ok := c.records[e.Host]
	if ok && prev == fp {
		return true
	}
	c.records[e.Host] = fp
	return false
}

// Reset removes the cached fingerprint for host, forcing the next entry to be
// treated as non-duplicate regardless of its port set.
func (c *Cache) Reset(host string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.records, host)
}

// Flush clears all cached fingerprints.
func (c *Cache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.records = make(map[string]Fingerprint)
}

// fingerprint derives a deterministic string key from a sorted port slice.
// The slice is expected to already be in ascending order (as produced by the
// scanner), so we just concatenate the values.
func fingerprint(ports []int) Fingerprint {
	if len(ports) == 0 {
		return ""
	}
	buf := make([]byte, 0, len(ports)*6)
	for i, p := range ports {
		if i > 0 {
			buf = append(buf, ',')
		}
		buf = appendInt(buf, p)
	}
	return string(buf)
}

func appendInt(b []byte, n int) []byte {
	if n == 0 {
		return append(b, '0')
	}
	var tmp [10]byte
	pos := len(tmp)
	for n > 0 {
		pos--
		tmp[pos] = byte('0' + n%10)
		n /= 10
	}
	return append(b, tmp[pos:]...)
}
