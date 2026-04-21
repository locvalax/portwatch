package cooldown

import (
	"context"
	"fmt"

	"github.com/user/portwatch/internal/scanner"
)

// CooledScanner wraps a scanner.Scanner and skips scans for hosts that are
// still within their cooldown window.
type CooledScanner struct {
	inner    scanner.Scanner
	cooldown *Cooldown
}

// Scanner is the interface expected from the inner scanner.
type Scanner interface {
	Scan(ctx context.Context, host string, opts scanner.Options) ([]int, error)
}

// NewCooledScanner returns a CooledScanner that gates calls to inner using cd.
func NewCooledScanner(inner Scanner, cd *Cooldown) *CooledScanner {
	return &CooledScanner{inner: inner, cooldown: cd}
}

// Scan delegates to the inner scanner only when the cooldown for host has
// elapsed. If the host is still cooling down, it returns an error describing
// the situation so callers can log or skip accordingly.
func (c *CooledScanner) Scan(ctx context.Context, host string, opts scanner.Options) ([]int, error) {
	if !c.cooldown.Allow(host) {
		return nil, fmt.Errorf("cooldown: host %q is within cooldown window, skipping scan", host)
	}
	return c.inner.Scan(ctx, host, opts)
}
