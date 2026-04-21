// Package rollup aggregates multiple port-change events within a time
// window into a single batched notification, reducing alert noise.
package rollup

import (
	"sync"
	"time"

	"github.com/user/portwatch/internal/store"
)

// Options configures the rollup window behaviour.
type Options struct {
	// Window is how long to accumulate events before flushing.
	Window time.Duration
	// MaxBatch is the maximum number of diffs held before an early flush.
	MaxBatch int
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		Window:   30 * time.Second,
		MaxBatch: 50,
	}
}

// Diff mirrors store.Diff so callers don't need an extra import.
type Diff = store.Diff

// FlushFunc is called with the accumulated batch when the window expires or
// the batch limit is reached.
type FlushFunc func(batch []Diff)

// Aggregator collects diffs and flushes them in batches.
type Aggregator struct {
	opts  Options
	flush FlushFunc

	mu      sync.Mutex
	batch   []Diff
	timer   *time.Timer
	closed  bool
}

// New creates an Aggregator and starts its internal timer.
func New(opts Options, fn FlushFunc) *Aggregator {
	a := &Aggregator{
		opts:  opts,
		flush: fn,
	}
	a.resetTimer()
	return a
}

// Add enqueues a diff. If the batch limit is reached the batch is flushed
// immediately.
func (a *Aggregator) Add(d Diff) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.closed {
		return
	}

	a.batch = append(a.batch, d)
	if len(a.batch) >= a.opts.MaxBatch {
		a.flushLocked()
		a.resetTimer()
	}
}

// Close flushes any remaining events and stops the timer.
func (a *Aggregator) Close() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.closed {
		return
	}
	a.closed = true
	if a.timer != nil {
		a.timer.Stop()
	}
	a.flushLocked()
}

// flushLocked sends the current batch to the FlushFunc and resets the slice.
// Caller must hold a.mu.
func (a *Aggregator) flushLocked() {
	if len(a.batch) == 0 {
		return
	}
	copy := make([]Diff, len(a.batch))
	copy(copy, a.batch)
	a.batch = a.batch[:0]
	go a.flush(copy)
}

func (a *Aggregator) resetTimer() {
	if a.timer != nil {
		a.timer.Stop()
	}
	a.timer = time.AfterFunc(a.opts.Window, func() {
		a.mu.Lock()
		defer a.mu.Unlock()
		if !a.closed {
			a.flushLocked()
			a.resetTimer()
		}
	})
}
