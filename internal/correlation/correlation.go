package correlation

import (
	"fmt"
	"sync"
	"time"

	"github.com/user/portwatch/internal/store"
)

// Event represents a correlated port-change event across hosts.
type Event struct {
	ID        string
	Hosts     []string
	OpenedAt  time.Time
	PortCount int
}

// Options configures the Correlator.
type Options struct {
	// Window is the time range within which changes on different hosts
	// are considered part of the same event.
	Window time.Duration
	// MinHosts is the minimum number of hosts that must show the same
	// port change before an event is emitted.
	MinHosts int
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		Window:   2 * time.Minute,
		MinHosts: 2,
	}
}

// Correlator groups port-change diffs from multiple hosts into Events.
type Correlator struct {
	opts    Options
	mu      sync.Mutex
	buckets map[int][]bucket // keyed by port number
	clock   func() time.Time
}

type bucket struct {
	host      string
	recordedAt time.Time
}

// New creates a Correlator with the given options.
func New(opts Options) *Correlator {
	return &Correlator{
		opts:    opts,
		buckets: make(map[int][]bucket),
		clock:   time.Now,
	}
}

// Observe records opened ports for a host and returns any correlated Events.
func (c *Correlator) Observe(host string, diff store.Diff) []Event {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.clock()
	var events []Event

	for _, port := range diff.Opened {
		c.prune(port, now)
		c.buckets[port] = append(c.buckets[port], bucket{host: host, recordedAt: now})
		if len(c.buckets[port]) >= c.opts.MinHosts {
			hosts := make([]string, len(c.buckets[port]))
			for i, b := range c.buckets[port] {
				hosts[i] = b.host
			}
			events = append(events, Event{
				ID:        fmt.Sprintf("port-%d-%d", port, now.UnixNano()),
				Hosts:     hosts,
				OpenedAt:  now,
				PortCount: port,
			})
			delete(c.buckets, port)
		}
	}

	return events
}

// prune removes stale bucket entries outside the correlation window.
func (c *Correlator) prune(port int, now time.Time) {
	valid := c.buckets[port][:0]
	for _, b := range c.buckets[port] {
		if now.Sub(b.recordedAt) <= c.opts.Window {
			valid = append(valid, b)
		}
	}
	c.buckets[port] = valid
}
