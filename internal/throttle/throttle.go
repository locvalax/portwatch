package throttle

import (
	"sync"
	"time"
)

// Throttle limits how frequently a function can be called per host
// using a token bucket approach with a fixed burst size.
type Throttle struct {
	mu       sync.Mutex
	tokens   map[string]int
	lastFill map[string]time.Time
	rate     time.Duration
	burst    int
}

// Options configures a Throttle.
type Options struct {
	// Rate is the interval at which one token is added per host.
	Rate time.Duration
	// Burst is the maximum number of tokens a host can accumulate.
	Burst int
}

// DefaultOptions returns sensible throttle defaults.
func DefaultOptions() Options {
	return Options{
		Rate:  10 * time.Second,
		Burst: 3,
	}
}

// New creates a Throttle with the given options.
func New(opts Options) *Throttle {
	if opts.Burst <= 0 {
		opts.Burst = 1
	}
	if opts.Rate <= 0 {
		opts.Rate = time.Second
	}
	return &Throttle{
		tokens:   make(map[string]int),
		lastFill: make(map[string]time.Time),
		rate:     opts.Rate,
		burst:    opts.Burst,
	}
}

// Allow returns true if the host is permitted to proceed, consuming one token.
// Tokens are replenished over time at the configured rate up to the burst limit.
func (t *Throttle) Allow(host string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	last, seen := t.lastFill[host]
	if !seen {
		t.tokens[host] = t.burst
		t.lastFill[host] = now
	} else {
		elapsed := now.Sub(last)
		added := int(elapsed / t.rate)
		if added > 0 {
			t.tokens[host] = min(t.burst, t.tokens[host]+added)
			t.lastFill[host] = last.Add(time.Duration(added) * t.rate)
		}
	}

	if t.tokens[host] <= 0 {
		return false
	}
	t.tokens[host]--
	return true
}

// Reset clears the token state for a host.
func (t *Throttle) Reset(host string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.tokens, host)
	delete(t.lastFill, host)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
