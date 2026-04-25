package trend

import (
	"context"
	"time"

	"github.com/user/portwatch/internal/store"
)

// ScanFunc matches the signature used by scanner.Scan.
type ScanFunc func(ctx context.Context, host string) (store.Entry, error)

// TrackingScanner wraps a ScanFunc and feeds results into an Analyzer.
type TrackingScanner struct {
	inner    ScanFunc
	analyzer *Analyzer
	clock    func() time.Time
}

// NewTrackingScanner returns a TrackingScanner that records port counts after
// each successful scan.
func NewTrackingScanner(inner ScanFunc, a *Analyzer) *TrackingScanner {
	return &TrackingScanner{
		inner:    inner,
		analyzer: a,
		clock:    time.Now,
	}
}

// Scan delegates to the inner scanner and records the resulting port count.
func (t *TrackingScanner) Scan(ctx context.Context, host string) (store.Entry, error) {
	entry, err := t.inner(ctx, host)
	if err != nil {
		return entry, err
	}
	t.analyzer.Record(host, len(entry.Ports), t.clock())
	return entry, nil
}
