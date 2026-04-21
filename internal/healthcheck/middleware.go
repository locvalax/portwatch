package healthcheck

import (
	"context"
	"fmt"

	"github.com/user/portwatch/internal/scanner"
)

// GuardedScanner wraps a scanner.Scan function and skips hosts that
// fail the health probe, returning an error for unreachable hosts.
type GuardedScanner struct {
	checker *Checker
	scan    func(ctx context.Context, host string, opts scanner.Options) ([]int, error)
}

// NewGuardedScanner creates a GuardedScanner using the provided checker
// and underlying scan function.
func NewGuardedScanner(c *Checker, scan func(ctx context.Context, host string, opts scanner.Options) ([]int, error)) *GuardedScanner {
	return &GuardedScanner{checker: c, scan: scan}
}

// Scan probes the host first; if unreachable it returns an error without
// invoking the underlying scanner.
func (g *GuardedScanner) Scan(ctx context.Context, host string, opts scanner.Options) ([]int, error) {
	status := g.checker.Probe(ctx, host)
	if !status.Reachable {
		return nil, fmt.Errorf("healthcheck: host %q unreachable: %w", host, status.Err)
	}
	return g.scan(ctx, host, opts)
}
