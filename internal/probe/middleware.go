package probe

import (
	"context"
	"fmt"

	"github.com/user/portwatch/internal/scanner"
)

// InstrumentedScanner wraps a scanner.Scan-compatible function and probes
// each discovered port, enriching results with latency data written to stderr.
type InstrumentedScanner struct {
	inner  func(ctx context.Context, host string, opts scanner.Options) ([]int, error)
	prober *Prober
}

// NewInstrumentedScanner returns a scanner that probes each open port for
// latency after the underlying scan completes.
func NewInstrumentedScanner(
	inner func(ctx context.Context, host string, opts scanner.Options) ([]int, error),
	prober *Prober,
) *InstrumentedScanner {
	return &InstrumentedScanner{inner: inner, prober: prober}
}

// Scan runs the underlying scanner then probes each returned port.
// It returns the same port list; probe results are available via Probe directly.
func (s *InstrumentedScanner) Scan(ctx context.Context, host string, opts scanner.Options) ([]int, error) {
	ports, err := s.inner(ctx, host, opts)
	if err != nil {
		return nil, err
	}

	targets := make([]Target, len(ports))
	for i, p := range ports {
		targets[i] = Target{Host: host, Port: p}
	}

	results := s.prober.ProbeAll(ctx, targets)
	for _, r := range results {
		if r.Err == nil {
			_ = fmt.Sprintf("probe: %s:%d latency=%s", r.Host, r.Port, r.Latency)
		}
	}

	return ports, nil
}
