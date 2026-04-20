// Package throttle provides token-bucket rate limiting for port scans.
// middleware.go wraps the Scanner interface with per-host throttle enforcement.
package throttle

import (
	"context"
	"fmt"

	"github.com/user/portwatch/internal/scanner"
)

// Scanner is the interface satisfied by scanner.Scan and compatible wrappers.
type Scanner interface {
	Scan(ctx context.Context, host string, opts scanner.Options) ([]int, error)
}

// ScanFunc is a function adapter that implements Scanner.
type ScanFunc func(ctx context.Context, host string, opts scanner.Options) ([]int, error)

// Scan implements Scanner.
func (f ScanFunc) Scan(ctx context.Context, host string, opts scanner.Options) ([]int, error) {
	return f(ctx, host, opts)
}

// ThrottledScanner wraps a Scanner and enforces per-host token-bucket limits.
type ThrottledScanner struct {
	inner    Scanner
	throttle *Throttle
}

// NewThrottledScanner returns a ThrottledScanner that gates calls to inner
// using the provided Throttle. Calls that exceed the burst capacity are
// rejected immediately rather than queued, keeping the CLI responsive.
func NewThrottledScanner(inner Scanner, t *Throttle) *ThrottledScanner {
	return &ThrottledScanner{
		inner:    inner,
		throttle: t,
	}
}

// Scan checks whether the host is within its allowed rate before delegating
// to the underlying scanner. If the token bucket is exhausted the call
// returns an error without performing a network scan.
func (ts *ThrottledScanner) Scan(ctx context.Context, host string, opts scanner.Options) ([]int, error) {
	if !ts.throttle.Allow(host) {
		return nil, fmt.Errorf("throttle: scan of %q rejected — rate limit exceeded", host)
	}
	return ts.inner.Scan(ctx, host, opts)
}

// Reset clears the token-bucket state for host, useful in tests or after a
// deliberate operator-initiated rescan.
func (ts *ThrottledScanner) Reset(host string) {
	ts.throttle.Reset(host)
}
