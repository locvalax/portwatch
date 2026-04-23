package timeout

import (
	"context"
	"fmt"

	"github.com/user/portwatch/internal/scanner"
)

// TimedScanner wraps a scanner.Scan-compatible function with a per-call deadline.
type TimedScanner struct {
	guard *Guard
	next  func(ctx context.Context, host string, opts scanner.Options) ([]int, error)
}

// NewTimedScanner creates a TimedScanner that enforces the given Guard's deadline
// before delegating to next.
func NewTimedScanner(g *Guard, next func(ctx context.Context, host string, opts scanner.Options) ([]int, error)) *TimedScanner {
	return &TimedScanner{guard: g, next: next}
}

// Scan executes the underlying scan within the configured deadline.
// If the deadline is exceeded, ErrTimeout is returned.
func (t *TimedScanner) Scan(ctx context.Context, host string, opts scanner.Options) ([]int, error) {
	ctx, cancel := t.guard.Wrap(ctx)
	defer cancel()

	type result struct {
		ports []int
		err   error
	}

	ch := make(chan result, 1)
	go func() {
		ports, err := t.next(ctx, host, opts)
		ch <- result{ports, err}
	}()

	select {
	case r := <-ch:
		return r.ports, r.err
	case <-ctx.Done():
		return nil, fmt.Errorf("%w: host %s", ErrTimeout, host)
	}
}
