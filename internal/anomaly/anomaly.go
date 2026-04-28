// Package anomaly detects statistically unusual port scan results by comparing
// observed open-port counts against a rolling baseline mean and standard deviation.
package anomaly

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// DefaultOptions returns a sensible Detector configuration.
func DefaultOptions() Options {
	return Options{
		WindowSize:  20,
		ZThreshold:  2.5,
		MinSamples:  5,
		MaxAge:      24 * time.Hour,
	}
}

// Options controls how the Detector identifies anomalies.
type Options struct {
	// WindowSize is the maximum number of historical samples retained per host.
	WindowSize int
	// ZThreshold is the minimum z-score magnitude considered anomalous.
	ZThreshold float64
	// MinSamples is the minimum number of observations required before scoring.
	MinSamples int
	// MaxAge discards samples older than this duration (0 = no expiry).
	MaxAge time.Duration
}

// Event describes a detected anomaly for a single host.
type Event struct {
	Host      string
	Observed  int
	Mean      float64
	StdDev    float64
	ZScore    float64
	Timestamp time.Time
}

func (e Event) String() string {
	return fmt.Sprintf("anomaly host=%s observed=%d mean=%.2f stddev=%.2f z=%.2f",
		e.Host, e.Observed, e.Mean, e.StdDev, e.ZScore)
}

type sample struct {
	count int
	at    time.Time
}

// Detector tracks per-host port-count history and flags anomalous observations.
type Detector struct {
	mu      sync.Mutex
	opts    Options
	samples map[string][]sample
	now     func() time.Time
}

// New creates a Detector with the supplied options.
func New(opts Options) (*Detector, error) {
	if opts.WindowSize < 1 {
		return nil, fmt.Errorf("anomaly: WindowSize must be >= 1")
	}
	if opts.ZThreshold <= 0 {
		return nil, fmt.Errorf("anomaly: ZThreshold must be > 0")
	}
	if opts.MinSamples < 2 {
		return nil, fmt.Errorf("anomaly: MinSamples must be >= 2")
	}
	return &Detector{
		opts:    opts,
		samples: make(map[string][]sample),
		now:     time.Now,
	}, nil
}

// Observe records portCount for host and returns an Event if the observation is
// anomalous, or nil when there is insufficient history or the value is normal.
func (d *Detector) Observe(host string, portCount int) *Event {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.now()
	d.prune(host, now)

	history := d.samples[host]
	mean, stddev := stats(history)

	// Record the new sample after computing stats so it does not bias its own score.
	d.samples[host] = append(history, sample{count: portCount, at: now})
	if len(d.samples[host]) > d.opts.WindowSize {
		d.samples[host] = d.samples[host][1:]
	}

	if len(history) < d.opts.MinSamples {
		return nil
	}
	if stddev == 0 {
		return nil
	}

	z := (float64(portCount) - mean) / stddev
	if math.Abs(z) < d.opts.ZThreshold {
		return nil
	}

	return &Event{
		Host:      host,
		Observed:  portCount,
		Mean:      mean,
		StdDev:    stddev,
		ZScore:    z,
		Timestamp: now,
	}
}

// Reset removes all stored samples for host.
func (d *Detector) Reset(host string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.samples, host)
}

// prune removes samples older than MaxAge (caller must hold d.mu).
func (d *Detector) prune(host string, now time.Time) {
	if d.opts.MaxAge == 0 {
		return
	}
	cutoff := now.Add(-d.opts.MaxAge)
	ss := d.samples[host]
	i := 0
	for i < len(ss) && ss[i].at.Before(cutoff) {
		i++
	}
	d.samples[host] = ss[i:]
}

// stats returns mean and population standard deviation of the sample slice.
func stats(ss []sample) (mean, stddev float64) {
	if len(ss) == 0 {
		return 0, 0
	}
	sum := 0.0
	for _, s := range ss {
		sum += float64(s.count)
	}
	mean = sum / float64(len(ss))
	variance := 0.0
	for _, s := range ss {
		d := float64(s.count) - mean
		variance += d * d
	}
	variance /= float64(len(ss))
	return mean, math.Sqrt(variance)
}
