package quota

import (
	"errors"
	"sync"
	"time"
)

// ErrQuotaExceeded is returned when a host has exhausted its scan quota.
var ErrQuotaExceeded = errors.New("quota: scan quota exceeded for host")

// Options configures the quota limiter.
type Options struct {
	// MaxScans is the maximum number of scans allowed per window.
	MaxScans int
	// Window is the duration of each quota window.
	Window time.Duration
}

// DefaultOptions returns sensible quota defaults.
func DefaultOptions() Options {
	return Options{
		MaxScans: 10,
		Window:   time.Hour,
	}
}

type bucket struct {
	count     int
	resetsAt  time.Time
}

// Limiter tracks per-host scan quotas over a rolling window.
type Limiter struct {
	mu      sync.Mutex
	opts    Options
	buckets map[string]*bucket
	now     func() time.Time
}

// New creates a Limiter with the given options.
func New(opts Options) *Limiter {
	return &Limiter{
		opts:    opts,
		buckets: make(map[string]*bucket),
		now:     time.Now,
	}
}

// Allow reports whether the host is within its quota.
// It increments the counter and returns ErrQuotaExceeded if the limit is hit.
func (l *Limiter) Allow(host string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	b, ok := l.buckets[host]
	if !ok || now.After(b.resetsAt) {
		l.buckets[host] = &bucket{count: 1, resetsAt: now.Add(l.opts.Window)}
		return nil
	}
	if b.count >= l.opts.MaxScans {
		return ErrQuotaExceeded
	}
	b.count++
	return nil
}

// Remaining returns how many scans are left in the current window for host.
func (l *Limiter) Remaining(host string) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	b, ok := l.buckets[host]
	if !ok || now.After(b.resetsAt) {
		return l.opts.MaxScans
	}
	return max(0, l.opts.MaxScans-b.count)
}

// Reset clears the quota bucket for host.
func (l *Limiter) Reset(host string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.buckets, host)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
