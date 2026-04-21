package watch

import (
	"context"
	"fmt"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/scanner"
	"github.com/user/portwatch/internal/store"
)

// Options configures the watch loop.
type Options struct {
	Hosts    []string
	Interval time.Duration
	Once     bool
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		Interval: 60 * time.Second,
		Once:     false,
	}
}

// Watcher continuously scans hosts and alerts on port changes.
type Watcher struct {
	opts    Options
	scanner *scanner.Scanner
	store   *store.Store
	alerter *alert.Alerter
}

// New creates a Watcher with the given dependencies.
func New(opts Options, sc *scanner.Scanner, st *store.Store, al *alert.Alerter) *Watcher {
	return &Watcher{
		opts:    opts,
		scanner: sc,
		store:   st,
		alerter: al,
	}
}

// Run starts the watch loop. It blocks until ctx is cancelled or Once is set.
func (w *Watcher) Run(ctx context.Context) error {
	if err := w.tick(ctx); err != nil {
		return err
	}
	if w.opts.Once {
		return nil
	}
	ticker := time.NewTicker(w.opts.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := w.tick(ctx); err != nil {
				return err
			}
		}
	}
}

func (w *Watcher) tick(ctx context.Context) error {
	for _, host := range w.opts.Hosts {
		ports, err := w.scanner.Scan(ctx, host)
		if err != nil {
			return fmt.Errorf("scan %s: %w", host, err)
		}
		if err := w.store.Append(host, ports); err != nil {
			return fmt.Errorf("store %s: %w", host, err)
		}
		if err := w.alerter.Notify(ctx, host); err != nil {
			return fmt.Errorf("alert %s: %w", host, err)
		}
	}
	return nil
}
