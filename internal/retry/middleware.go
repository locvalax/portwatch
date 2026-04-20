package retry

import (
	"context"
	"fmt"

	"github.com/user/portwatch/internal/scanner"
)

// RetryScanner wraps a scanner.Scan call with retry logic.
type RetryScanner struct {
	policy Policy
}

// NewRetryScanner returns a RetryScanner using the supplied policy.
func NewRetryScanner(p Policy) *RetryScanner {
	return &RetryScanner{policy: p}
}

// ScanResult holds the ports returned by a successful scan.
type ScanResult struct {
	Host  string
	Ports []int
}

// Scan attempts to scan host, retrying according to the policy.
func (r *RetryScanner) Scan(ctx context.Context, host string, opts scanner.Options) (ScanResult, error) {
	var result ScanResult

	err := r.policy.Do(ctx, func() error {
		ports, err := scanner.Scan(host, opts)
		if err != nil {
			return fmt.Errorf("scan %s: %w", host, err)
		}
		result = ScanResult{Host: host, Ports: ports}
		return nil
	})

	return result, err
}
