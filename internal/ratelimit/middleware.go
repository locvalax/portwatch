package ratelimit

import (
	"fmt"
	"time"

	"github.com/user/portwatch/internal/scanner"
)

// GuardedScanner wraps a scan function with rate limiting per host.
type GuardedScanner struct {
	limiter *Limiter
	opts    scanner.Options
}

// NewGuardedScanner creates a GuardedScanner with the given interval and scan options.
func NewGuardedScanner(interval time.Duration, opts scanner.Options) *GuardedScanner {
	return &GuardedScanner{
		limiter: New(interval),
		opts:    opts,
	}
}

// Scan checks the rate limit for the host before scanning.
// Returns an error if the host is rate-limited.
func (g *GuardedScanner) Scan(host string) ([]int, error) {
	if !g.limiter.Allow(host) {
		return nil, fmt.Errorf("rate limit: host %q scanned too recently", host)
	}
	return scanner.Scan(host, g.opts)
}
