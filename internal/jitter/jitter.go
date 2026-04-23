// Package jitter adds randomised delay to scan intervals to avoid
// thundering-herd effects when many hosts are polled concurrently.
package jitter

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

// Options controls jitter behaviour.
type Options struct {
	// Factor is the maximum fraction of the base interval that may be added
	// as random delay. Must be in the range (0, 1]. Default: 0.25.
	Factor float64
	// Rand is the source of randomness. Defaults to a package-level source.
	Rand *rand.Rand
}

// DefaultOptions returns production-ready defaults.
func DefaultOptions() Options {
	return Options{Factor: 0.25}
}

// Jitter holds state for computing jittered intervals.
type Jitter struct {
	opts Options
	mu  sync.Mutex
	rng *rand.Rand
}

// New creates a Jitter with the given options.
// If opts.Factor is outside (0, 1] it is clamped to 0.25.
func New(opts Options) *Jitter {
	if opts.Factor <= 0 || opts.Factor > 1 {
		opts.Factor = 0.25
	}
	rng := opts.Rand
	if rng == nil {
		//nolint:gosec // non-cryptographic use
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	return &Jitter{opts: opts, rng: rng}
}

// Apply returns base plus a random duration in [0, base*Factor).
func (j *Jitter) Apply(base time.Duration) time.Duration {
	j.mu.Lock()
	f := j.rng.Float64()
	j.mu.Unlock()
	extra := time.Duration(float64(base) * j.opts.Factor * f)
	return base + extra
}

// Sleep blocks until base+jitter has elapsed or ctx is cancelled.
// It returns ctx.Err() if the context fires before the delay expires.
func (j *Jitter) Sleep(ctx context.Context, base time.Duration) error {
	delay := j.Apply(base)
	select {
	case <-time.After(delay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
