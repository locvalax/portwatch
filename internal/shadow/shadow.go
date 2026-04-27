// Package shadow implements shadow-mode scanning: runs a secondary scanner
// alongside the primary and logs divergences without affecting results.
package shadow

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"sync"

	"github.com/user/portwatch/internal/store"
)

// Scanner is the interface expected by the shadow runner.
type Scanner interface {
	Scan(ctx context.Context, host string) ([]int, error)
}

// Options configures shadow-mode behaviour.
type Options struct {
	// Log is the writer used for divergence reports. Defaults to os.Stderr.
	Log io.Writer
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{Log: os.Stderr}
}

// Runner wraps a primary and shadow Scanner. It always returns the primary
// result; divergences are written to Options.Log.
type Runner struct {
	primary Scanner
	shadow  Scanner
	log     io.Writer
	mu      sync.Mutex
	divs    []Divergence
}

// Divergence records a single mismatch between primary and shadow scans.
type Divergence struct {
	Host        string
	PrimaryOnly []int
	ShadowOnly  []int
}

// New creates a Runner.
func New(primary, shadow Scanner, opts Options) *Runner {
	if opts.Log == nil {
		opts.Log = os.Stderr
	}
	return &Runner{primary: primary, shadow: shadow, log: opts.Log}
}

// Scan runs both scanners concurrently and returns the primary result.
func (r *Runner) Scan(ctx context.Context, host string) ([]int, error) {
	type result struct {
		ports []int
		err   error
	}

	pCh := make(chan result, 1)
	sCh := make(chan result, 1)

	go func() { p, e := r.primary.Scan(ctx, host); pCh <- result{p, e} }()
	go func() { s, e := r.shadow.Scan(ctx, host); sCh <- result{s, e} }()

	pr := <-pCh
	sr := <-sCh

	if pr.err == nil && sr.err == nil {
		if d, ok := compare(host, pr.ports, sr.ports); ok {
			r.mu.Lock()
			r.divs = append(r.divs, d)
			r.mu.Unlock()
			fmt.Fprintf(r.log, "[shadow] divergence on %s: primary_only=%v shadow_only=%v\n",
				host, d.PrimaryOnly, d.ShadowOnly)
		}
	} else if sr.err != nil {
		fmt.Fprintf(r.log, "[shadow] error scanning %s: %v\n", host, sr.err)
	}

	return pr.ports, pr.err
}

// Divergences returns a snapshot of all recorded divergences.
func (r *Runner) Divergences() []Divergence {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]Divergence, len(r.divs))
	copy(out, r.divs)
	return out
}

func compare(host string, a, b []int) (Divergence, bool) {
	setA := toSet(a)
	setB := toSet(b)
	var onlyA, onlyB []int
	for p := range setA {
		if !setB[p] {
			onlyA = append(onlyA, p)
		}
	}
	for p := range setB {
		if !setA[p] {
			onlyB = append(onlyB, p)
		}
	}
	if len(onlyA) == 0 && len(onlyB) == 0 {
		return Divergence{}, false
	}
	sort.Ints(onlyA)
	sort.Ints(onlyB)
	return Divergence{Host: host, PrimaryOnly: onlyA, ShadowOnly: onlyB}, true
}

func toSet(ports []int) map[int]bool {
	s := make(map[int]bool, len(ports))
	for _, p := range ports {
		s[p] = true
	}
	return s
}

// Ensure Runner satisfies Scanner.
var _ Scanner = (*Runner)(nil)

// Ensure store.Entry is importable without direct use warnings.
var _ = store.Entry{}
