// Package cooldown enforces a minimum quiet period between repeated alerts
// for the same host, preventing notification storms after a flap.
package cooldown

import (
	"sync"
	"time"
)

// Options configures the cooldown behaviour.
type Options struct {
	// Period is the minimum duration that must elapse before the same host
	// can trigger another alert.
	Period time.Duration
	// Now is an injectable clock; defaults to time.Now.
	Now func() time.Time
}

// DefaultOptions returns sensible production defaults.
func DefaultOptions() Options {
	return Options{
		Period: 5 * time.Minute,
		Now:    time.Now,
	}
}

// Cooldown tracks per-host alert timestamps and gates repeated firings.
type Cooldown struct {
	mu   sync.Mutex
	opts Options
	last map[string]time.Time
}

// New creates a Cooldown with the given options.
func New(opts Options) *Cooldown {
	if opts.Now == nil {
		opts.Now = time.Now
	}
	if opts.Period <= 0 {
		opts.Period = DefaultOptions().Period
	}
	return &Cooldown{
		opts: opts,
		last: make(map[string]time.Time),
	}
}

// Allow returns true if enough time has passed since the last alert for host.
// Calling Allow also records the current time as the latest alert for that host
// when it returns true.
func (c *Cooldown) Allow(host string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.opts.Now()
	if t, ok := c.last[host]; ok && now.Sub(t) < c.opts.Period {
		return false
	}
	c.last[host] = now
	return true
}

// Reset clears the recorded timestamp for host, allowing the next call to
// Allow to pass immediately.
func (c *Cooldown) Reset(host string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.last, host)
}

// Flush clears all recorded timestamps.
func (c *Cooldown) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.last = make(map[string]time.Time)
}
