package metrics

import (
	"sync"
	"time"
)

// Counter tracks scan and alert counts per host.
type Counter struct {
	mu      sync.Mutex
	scans   map[string]int
	alerts  map[string]int
	errors  map[string]int
	lastScan map[string]time.Time
}

// New returns an initialised Counter.
func New() *Counter {
	return &Counter{
		scans:    make(map[string]int),
		alerts:   make(map[string]int),
		errors:   make(map[string]int),
		lastScan: make(map[string]time.Time),
	}
}

// RecordScan increments the scan count for host and records the timestamp.
func (c *Counter) RecordScan(host string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.scans[host]++
	c.lastScan[host] = time.Now()
}

// RecordAlert increments the alert count for host.
func (c *Counter) RecordAlert(host string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.alerts[host]++
}

// RecordError increments the error count for host.
func (c *Counter) RecordError(host string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.errors[host]++
}

// Snapshot returns a point-in-time copy of all metrics.
func (c *Counter) Snapshot() map[string]HostMetrics {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make(map[string]HostMetrics, len(c.scans))
	for host, s := range c.scans {
		out[host] = HostMetrics{
			Host:     host,
			Scans:    s,
			Alerts:   c.alerts[host],
			Errors:   c.errors[host],
			LastScan: c.lastScan[host],
		}
	}
	return out
}

// Reset clears all counters for host.
func (c *Counter) Reset(host string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.scans, host)
	delete(c.alerts, host)
	delete(c.errors, host)
	delete(c.lastScan, host)
}

// HostMetrics holds aggregated metrics for a single host.
type HostMetrics struct {
	Host     string
	Scans    int
	Alerts   int
	Errors   int
	LastScan time.Time
}
