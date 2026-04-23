package sampler

import (
	"context"
	"fmt"

	"github.com/user/portwatch/internal/scanner"
)

// SampledScanner wraps a scanner.Scanner and skips scans that the Sampler
// does not allow, returning a sentinel error instead.
type SampledScanner struct {
	inner   *scanner.Scanner
	sampler *Sampler
}

// ErrSkipped is returned when a scan is skipped by the sampler.
var ErrSkipped = fmt.Errorf("sampler: scan skipped by rate limiter")

// NewSampledScanner wraps inner with probabilistic sampling using s.
func NewSampledScanner(inner *scanner.Scanner, s *Sampler) *SampledScanner {
	return &SampledScanner{inner: inner, sampler: s}
}

// Scan runs the underlying scan only if the sampler allows it.
// When skipped, ErrSkipped is returned and no ports are scanned.
func (ss *SampledScanner) Scan(ctx context.Context, host string) ([]int, error) {
	if !ss.sampler.Allow(host) {
		return nil, ErrSkipped
	}
	return ss.inner.Scan(ctx, host)
}
