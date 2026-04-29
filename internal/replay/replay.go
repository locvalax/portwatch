// Package replay provides utilities for replaying historical scan entries
// through the alert and notify pipeline, useful for testing configurations
// or re-processing missed events.
package replay

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/user/portwatch/internal/store"
	"github.com/user/portwatch/internal/store/diff"
)

// Options controls replay behaviour.
type Options struct {
	// Speed is a multiplier applied to the original inter-entry delay.
	// 1.0 replays at real time, 0 replays as fast as possible.
	Speed float64

	// Since discards entries older than this duration (0 = no filter).
	Since time.Duration

	// Writer receives human-readable replay progress lines.
	Writer io.Writer
}

// DefaultOptions returns sensible defaults for replay.
func DefaultOptions() Options {
	return Options{
		Speed:  0,
		Since:  0,
		Writer: os.Stdout,
	}
}

// Handler is called for every diff produced during replay.
type Handler func(ctx context.Context, d diff.Result) error

// Replayer reads stored scan history for a host and replays diffs in order.
type Replayer struct {
	st   *store.Store
	opts Options
}

// New creates a Replayer backed by st.
func New(st *store.Store, opts Options) *Replayer {
	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}
	return &Replayer{st: st, opts: opts}
}

// Run iterates the stored entries for host and calls h for each consecutive
// pair that produces a non-empty diff. It respects context cancellation.
func (r *Replayer) Run(ctx context.Context, host string, h Handler) error {
	entries, err := r.st.All(host)
	if err != nil {
		return fmt.Errorf("replay: load entries for %s: %w", host, err)
	}

	// Apply Since filter.
	if r.opts.Since > 0 {
		cutoff := time.Now().Add(-r.opts.Since)
		filtered := entries[:0]
		for _, e := range entries {
			if e.ScannedAt.After(cutoff) {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	if len(entries) < 2 {
		fmt.Fprintf(r.opts.Writer, "replay: fewer than 2 entries for %s, nothing to replay\n", host)
		return nil
	}

	fmt.Fprintf(r.opts.Writer, "replay: replaying %d entries for %s\n", len(entries), host)

	for i := 1; i < len(entries); i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		d := diff.Compare(entries[i-1], entries[i])
		if len(d.Opened) == 0 && len(d.Closed) == 0 {
			continue
		}

		// Optionally pace the replay to mimic original timing.
		if r.opts.Speed > 0 && i > 1 {
			gap := entries[i].ScannedAt.Sub(entries[i-1].ScannedAt)
			delay := time.Duration(float64(gap) / r.opts.Speed)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		fmt.Fprintf(r.opts.Writer, "replay: entry %d/%d at %s — +%d -%d\n",
			i, len(entries)-1,
			entries[i].ScannedAt.Format(time.RFC3339),
			len(d.Opened), len(d.Closed))

		if err := h(ctx, d); err != nil {
			return fmt.Errorf("replay: handler error at entry %d: %w", i, err)
		}
	}

	return nil
}
