package timeout

import (
	"context"
	"errors"
	"time"
)

// ErrTimeout is returned when a scan exceeds its deadline.
var ErrTimeout = errors.New("timeout: scan deadline exceeded")

// Options configures the per-host scan timeout.
type Options struct {
	// Deadline is the maximum duration allowed for a single host scan.
	Deadline time.Duration
}

// DefaultOptions returns sensible timeout defaults.
func DefaultOptions() Options {
	return Options{
		Deadline: 10 * time.Second,
	}
}

// Guard enforces a per-scan deadline.
type Guard struct {
	opts Options
}

// New creates a Guard with the provided options.
func New(opts Options) (*Guard, error) {
	if opts.Deadline <= 0 {
		return nil, errors.New("timeout: deadline must be positive")
	}
	return &Guard{opts: opts}, nil
}

// Wrap returns a context that is cancelled after the configured deadline.
// The caller is responsible for calling the returned cancel function.
func (g *Guard) Wrap(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, g.opts.Deadline)
}

// IsTimeout reports whether err represents a timeout condition.
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrTimeout) {
		return true
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	return false
}
