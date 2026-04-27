package backoff

import (
	"math"
	"time"
)

// Options configures exponential back-off behaviour.
type Options struct {
	// InitialInterval is the wait time after the first failure.
	InitialInterval time.Duration
	// MaxInterval caps the computed wait time.
	MaxInterval time.Duration
	// Multiplier is applied to the interval after each attempt.
	Multiplier float64
	// MaxAttempts is the maximum number of attempts (0 = unlimited).
	MaxAttempts int
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		InitialInterval: 200 * time.Millisecond,
		MaxInterval:     30 * time.Second,
		Multiplier:      2.0,
		MaxAttempts:     5,
	}
}

// Backoff computes successive wait durations for a given attempt number
// (zero-indexed). The returned duration is capped at MaxInterval.
type Backoff struct {
	opts Options
}

// New returns a Backoff using the provided options.
func New(opts Options) *Backoff {
	if opts.Multiplier <= 1 {
		opts.Multiplier = 2.0
	}
	return &Backoff{opts: opts}
}

// Duration returns the wait duration for the given attempt (0-indexed).
// If MaxAttempts is set and attempt >= MaxAttempts the second return value
// is false, indicating no further attempts should be made.
func (b *Backoff) Duration(attempt int) (time.Duration, bool) {
	if b.opts.MaxAttempts > 0 && attempt >= b.opts.MaxAttempts {
		return 0, false
	}
	scale := math.Pow(b.opts.Multiplier, float64(attempt))
	d := time.Duration(float64(b.opts.InitialInterval) * scale)
	if d > b.opts.MaxInterval {
		d = b.opts.MaxInterval
	}
	return d, true
}

// Sequence returns a slice of durations for all allowed attempts.
func (b *Backoff) Sequence() []time.Duration {
	var out []time.Duration
	for i := 0; ; i++ {
		d, ok := b.Duration(i)
		if !ok {
			break
		}
		out = append(out, d)
	}
	return out
}
