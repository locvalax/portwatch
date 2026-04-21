// Package debounce delays and coalesces rapid port-change events so that
// short-lived fluctuations do not trigger spurious alerts.
package debounce

import (
	"sync"
	"time"

	"github.com/user/portwatch/internal/store"
)

// Options configures the Debouncer.
type Options struct {
	// Wait is how long to wait after the last event before forwarding it.
	Wait time.Duration
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		Wait: 10 * time.Second,
	}
}

type pending struct {
	timer *time.Timer
	diff  store.Diff
}

// Debouncer holds back diff events until they have been stable for Wait.
type Debouncer struct {
	opts    Options
	mu      sync.Mutex
	pending map[string]*pending
	out     chan store.Diff
}

// New creates a Debouncer. Read coalesced diffs from C().
func New(opts Options) *Debouncer {
	return &Debouncer{
		opts:    opts,
		pending: make(map[string]*pending),
		out:     make(chan store.Diff, 64),
	}
}

// C returns the channel on which stable diffs are delivered.
func (d *Debouncer) C() <-chan store.Diff { return d.out }

// Add schedules diff for delivery after the debounce window.
// If a pending event already exists for the same host it is replaced,
// resetting the timer.
func (d *Debouncer) Add(diff store.Diff) {
	d.mu.Lock()
	defer d.mu.Unlock()

	host := diff.Host
	if p, ok := d.pending[host]; ok {
		p.timer.Stop()
	}

	t := time.AfterFunc(d.opts.Wait, func() {
		d.mu.Lock()
		p, ok := d.pending[host]
		if ok {
			delete(d.pending, host)
		}
		d.mu.Unlock()
		if ok {
			select {
			case d.out <- p.diff:
			default:
			}
		}
	})

	d.pending[host] = &pending{timer: t, diff: diff}
}

// Flush immediately delivers all pending diffs, cancelling their timers.
func (d *Debouncer) Flush() {
	d.mu.Lock()
	snap := make(map[string]*pending, len(d.pending))
	for k, v := range d.pending {
		snap[k] = v
	}
	d.pending = make(map[string]*pending)
	d.mu.Unlock()

	for _, p := range snap {
		p.timer.Stop()
		select {
		case d.out <- p.diff:
		default:
		}
	}
}
