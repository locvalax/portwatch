// Package window provides a sliding-window counter for tracking scan events
// over a rolling time period per host.
package window

import (
	"sync"
	"time"
)

// DefaultOptions returns a Window with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Size:     5 * time.Minute,
		MaxCount: 100,
	}
}

// Options configures the sliding window.
type Options struct {
	Size     time.Duration
	MaxCount int
	Clock    func() time.Time // injectable for testing
}

// Window tracks timestamped events per host within a rolling time window.
type Window struct {
	mu     sync.Mutex
	opts   Options
	events map[string][]time.Time
}

// New creates a new Window with the given options.
func New(opts Options) *Window {
	if opts.Clock == nil {
		opts.Clock = time.Now
	}
	if opts.Size <= 0 {
		opts.Size = DefaultOptions().Size
	}
	if opts.MaxCount <= 0 {
		opts.MaxCount = DefaultOptions().MaxCount
	}
	return &Window{
		opts:   opts,
		events: make(map[string][]time.Time),
	}
}

// Record adds an event for the given host and returns the current count
// within the window. Returns true if the event was accepted (count <= max),
// false if the host has exceeded the limit.
func (w *Window) Record(host string) (int, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := w.opts.Clock()
	cutoff := now.Add(-w.opts.Size)

	// evict old events
	prev := w.events[host]
	filtered := prev[:0]
	for _, t := range prev {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}

	filtered = append(filtered, now)
	w.events[host] = filtered

	count := len(filtered)
	return count, count <= w.opts.MaxCount
}

// Count returns the number of events recorded for host within the current window.
func (w *Window) Count(host string) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := w.opts.Clock()
	cutoff := now.Add(-w.opts.Size)
	count := 0
	for _, t := range w.events[host] {
		if t.After(cutoff) {
			count++
		}
	}
	return count
}

// Reset clears all recorded events for the given host.
func (w *Window) Reset(host string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.events, host)
}
