package quota

import (
	"context"
	"fmt"

	"github.com/user/portwatch/internal/scanner"
)

// guardedScanner wraps a scanner.Scanner and enforces per-host quotas.
type guardedScanner struct {
	inner   scanner.Scanner
	limiter *Limiter
}

// NewGuardedScanner returns a Scanner that rejects scans once the host
// quota is exhausted within the configured window.
func NewGuardedScanner(inner scanner.Scanner, limiter *Limiter) scanner.Scanner {
	return &guardedScanner{inner: inner, limiter: limiter}
}

func (g *guardedScanner) Scan(ctx context.Context, host string, ports []int) ([]int, error) {
	if err := g.limiter.Allow(host); err != nil {
		return nil, fmt.Errorf("%w: host=%s remaining=%d", err, host, g.limiter.Remaining(host))
	}
	return g.inner.Scan(ctx, host, ports)
}
