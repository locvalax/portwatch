package ratelimit

import (
	"sync"
	"time"
)

// Limiter enforces a minimum interval between scans per host.
type Limiter struct {
	mu       sync.Mutex
	last     map[string]time.Time
	interval time.Duration
}

// New creates a Limiter with the given minimum interval between events.
func New(interval time.Duration) *Limiter {
	return &Limiter{
		last:     make(map[string]time.Time),
		interval: interval,
	}
}

// Allow returns true if the host is allowed to proceed based on the rate limit.
func (l *Limiter) Allow(host string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	if t, ok := l.last[host]; ok {
		if now.Sub(t) < l.interval {
			return false
		}
	}
	l.last[host] = now
	return true
}

// Reset clears the rate limit state for a specific host.
func (l *Limiter) Reset(host string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.last, host)
}

// ResetAll clears all rate limit state.
func (l *Limiter) ResetAll() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.last = make(map[string]time.Time)
}
