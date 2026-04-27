package shadow

import (
	"context"
	"io"
	"os"
)

// ShadowedScanner is a convenience constructor that wraps primary with a
// shadow scanner, logging divergences to w (or os.Stderr if nil).
func NewShadowedScanner(primary, secondary Scanner, w io.Writer) *Runner {
	if w == nil {
		w = os.Stderr
	}
	return New(primary, secondary, Options{Log: w})
}

// noopScanner is a shadow that always returns empty results. Useful when
// shadow mode is configured but no secondary target is specified.
type noopScanner struct{}

func (n *noopScanner) Scan(_ context.Context, _ string) ([]int, error) {
	return nil, nil
}

// NewPassthrough wraps primary with a no-op shadow, effectively disabling
// divergence tracking while keeping the Runner interface.
func NewPassthrough(primary Scanner, w io.Writer) *Runner {
	return NewShadowedScanner(primary, &noopScanner{}, w)
}
