package decay

import (
	"context"
	"fmt"

	"github.com/user/portwatch/internal/scanner"
	"github.com/user/portwatch/internal/store"
)

// ScoredScanner wraps a scanner.Scanner and updates decay scores after each scan.
type ScoredScanner struct {
	inner  scanner.Scanner
	scorer *Scorer
	delta  float64
}

// NewScoredScanner returns a ScoredScanner that boosts each open port's score
// by delta on every successful observation.
func NewScoredScanner(inner scanner.Scanner, scorer *Scorer, delta float64) *ScoredScanner {
	if delta <= 0 {
		delta = 1.0
	}
	return &ScoredScanner{inner: inner, scorer: scorer, delta: delta}
}

// Scan delegates to the inner scanner and records decay observations for every
// open port returned in the result.
func (s *ScoredScanner) Scan(ctx context.Context, host string, opts scanner.Options) (store.Entry, error) {
	entry, err := s.inner.Scan(ctx, host, opts)
	if err != nil {
		return entry, err
	}
	for _, port := range entry.Ports {
		key := fmt.Sprintf("%s:%d", host, port)
		s.scorer.Observe(key, s.delta)
	}
	return entry, nil
}
