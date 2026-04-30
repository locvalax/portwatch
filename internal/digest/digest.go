// Package digest computes and tracks periodic scan digests,
// summarising port-state across hosts over a configurable window.
package digest

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/user/portwatch/internal/store"
)

// Summary is a rolled-up view of a single host for one digest window.
type Summary struct {
	Host      string
	Ports     []int
	Hash      string
	CreatedAt time.Time
}

// Options controls digest behaviour.
type Options struct {
	// Window is how far back in time entries are considered.
	Window time.Duration
	// Clock is used for time comparisons; defaults to time.Now.
	Clock func() time.Time
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		Window: 24 * time.Hour,
		Clock:  time.Now,
	}
}

// Digester builds summaries from store entries.
type Digester struct {
	mu   sync.Mutex
	opts Options
}

// New creates a Digester with the given options.
func New(opts Options) *Digester {
	if opts.Clock == nil {
		opts.Clock = time.Now
	}
	if opts.Window <= 0 {
		opts.Window = DefaultOptions().Window
	}
	return &Digester{opts: opts}
}

// Summarise returns a Summary for the given host using entries that fall
// within the configured window.  It returns the most-recent port snapshot
// found in that window.
func (d *Digester) Summarise(host string, entries []store.Entry) (Summary, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	cutoff := d.opts.Clock().Add(-d.opts.Window)
	var latest store.Entry
	found := false
	for _, e := range entries {
		if e.Host != host {
			continue
		}
		if e.ScannedAt.Before(cutoff) {
			continue
		}
		if !found || e.ScannedAt.After(latest.ScannedAt) {
			latest = e
			found = true
		}
	}
	if !found {
		return Summary{}, fmt.Errorf("digest: no entries for host %q within window", host)
	}

	ports := make([]int, len(latest.Ports))
	copy(ports, latest.Ports)
	sort.Ints(ports)

	return Summary{
		Host:      host,
		Ports:     ports,
		Hash:      hashPorts(host, ports),
		CreatedAt: d.opts.Clock(),
	}, nil
}

func hashPorts(host string, ports []int) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s", host)
	for _, p := range ports {
		fmt.Fprintf(h, ":%d", p)
	}
	return fmt.Sprintf("%x", h.Sum(nil))[:16]
}
