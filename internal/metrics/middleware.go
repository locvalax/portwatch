package metrics

import (
	"context"
	"fmt"

	"github.com/user/portwatch/internal/scanner"
)

// InstrumentedScanner wraps a scanner.Scan-compatible function and records
// per-host metrics for every invocation.
type InstrumentedScanner struct {
	counter *Counter
	next    func(ctx context.Context, host string, opts scanner.Options) ([]int, error)
}

// NewInstrumentedScanner returns an InstrumentedScanner that delegates to next
// and updates counter on each call.
func NewInstrumentedScanner(
	counter *Counter,
	next func(ctx context.Context, host string, opts scanner.Options) ([]int, error),
) *InstrumentedScanner {
	return &InstrumentedScanner{counter: counter, next: next}
}

// Scan records a scan attempt, delegates to the wrapped scanner, and records
// alerts or errors as appropriate.
func (s *InstrumentedScanner) Scan(
	ctx context.Context,
	host string,
	opts scanner.Options,
) ([]int, error) {
	s.counter.RecordScan(host)

	ports, err := s.next(ctx, host, opts)
	if err != nil {
		s.counter.RecordError(host)
		return nil, fmt.Errorf("instrumented scan %s: %w", host, err)
	}

	if len(ports) > 0 {
		s.counter.RecordAlert(host)
	}

	return ports, nil
}
