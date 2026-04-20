// Package suppress provides a suppression list that prevents alerts
// from firing repeatedly for the same port change within a quiet window.
package suppress

import (
	"fmt"
	"sync"
	"time"
)

// key uniquely identifies a host+port+direction combination.
type key struct {
	host      string
	port      int
	opened    bool
}

// Suppressor tracks recently alerted port changes and suppresses
// duplicate notifications within a configurable window.
type Suppressor struct {
	mu      sync.Mutex
	entries map[key]time.Time
	window  time.Duration
	now     func() time.Time
}

// New creates a Suppressor with the given quiet window duration.
// Alerts for the same host/port/direction are suppressed until the
// window has elapsed since the first alert.
func New(window time.Duration) *Suppressor {
	return &Suppressor{
		entries: make(map[key]time.Time),
		window:  window,
		now:     time.Now,
	}
}

// IsSuppressed reports whether an alert for the given host, port, and
// direction (opened=true / opened=false for closed) should be suppressed.
// If not suppressed, the entry is recorded so future calls within the
// window will be suppressed.
func (s *Suppressor) IsSuppressed(host string, port int, opened bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := key{host: host, port: port, opened: opened}
	now := s.now()

	if t, ok := s.entries[k]; ok && now.Before(t.Add(s.window)) {
		return true
	}

	s.entries[k] = now
	return false
}

// Reset clears the suppression record for a specific host.
func (s *Suppressor) Reset(host string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for k := range s.entries {
		if k.host == host {
			delete(s.entries, k)
		}
	}
}

// Flush removes all expired entries, freeing memory.
func (s *Suppressor) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	for k, t := range s.entries {
		if now.After(t.Add(s.window)) {
			delete(s.entries, k)
		}
	}
}

// String returns a human-readable summary of active suppressions.
func (s *Suppressor) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return fmt.Sprintf("suppress: %d active entries (window=%s)", len(s.entries), s.window)
}
