// Package limiter provides a concurrent scan limiter that caps the number
// of in-flight port scans across all hosts at any given time.
package limiter

import (
	"context"
	"fmt"

	"github.com/user/portwatch/internal/scanner"
)

// Options configures the concurrency limiter.
type Options struct {
	// MaxConcurrent is the maximum number of scans allowed to run simultaneously.
	MaxConcurrent int
}

// DefaultOptions returns sensible defaults for the limiter.
func DefaultOptions() Options {
	return Options{
		MaxConcurrent: 10,
	}
}

// Limiter controls how many scans may run concurrently.
type Limiter struct {
	sem chan struct{}
}

// New creates a Limiter with the given options.
func New(opts Options) (*Limiter, error) {
	if opts.MaxConcurrent <= 0 {
		return nil, fmt.Errorf("limiter: MaxConcurrent must be > 0, got %d", opts.MaxConcurrent)
	}
	return &Limiter{
		sem: make(chan struct{}, opts.MaxConcurrent),
	}, nil
}

// Acquire blocks until a slot is available or ctx is cancelled.
func (l *Limiter) Acquire(ctx context.Context) error {
	select {
	case l.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release frees a previously acquired slot.
func (l *Limiter) Release() {
	<-l.sem
}

// InFlight returns the number of scans currently running.
func (l *Limiter) InFlight() int {
	return len(l.sem)
}

// NewLimitedScanner wraps a scan function so that at most MaxConcurrent
// invocations run simultaneously.
func NewLimitedScanner(l *Limiter, next func(ctx context.Context, host string, opts scanner.Options) ([]int, error)) func(ctx context.Context, host string, opts scanner.Options) ([]int, error) {
	return func(ctx context.Context, host string, opts scanner.Options) ([]int, error) {
		if err := l.Acquire(ctx); err != nil {
			return nil, fmt.Errorf("limiter: acquire slot for %s: %w", host, err)
		}
		defer l.Release()
		return next(ctx, host, opts)
	}
}
