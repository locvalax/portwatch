package probe

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// DefaultOptions returns sensible defaults for the prober.
func DefaultOptions() Options {
	return Options{
		Timeout:     2 * time.Second,
		Concurrency: 10,
	}
}

// Options controls probe behaviour.
type Options struct {
	Timeout     time.Duration
	Concurrency int
}

// Result holds the outcome of probing a single host:port pair.
type Result struct {
	Host    string
	Port    int
	Open    bool
	Latency time.Duration
	Err     error
}

// Prober checks individual host:port reachability and measures latency.
type Prober struct {
	opts Options
}

// New returns a Prober with the given options.
func New(opts Options) (*Prober, error) {
	if opts.Concurrency <= 0 {
		return nil, fmt.Errorf("probe: concurrency must be > 0")
	}
	if opts.Timeout <= 0 {
		return nil, fmt.Errorf("probe: timeout must be > 0")
	}
	return &Prober{opts: opts}, nil
}

// Probe checks a single host:port and returns a Result.
func (p *Prober) Probe(ctx context.Context, host string, port int) Result {
	addr := fmt.Sprintf("%s:%d", host, port)
	start := time.Now()

	dialer := &net.Dialer{Timeout: p.opts.Timeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	latency := time.Since(start)

	if err != nil {
		return Result{Host: host, Port: port, Open: false, Latency: latency, Err: err}
	}
	_ = conn.Close()
	return Result{Host: host, Port: port, Open: true, Latency: latency}
}

// ProbeAll probes multiple host:port pairs concurrently and returns all results.
func (p *Prober) ProbeAll(ctx context.Context, targets []Target) []Result {
	if len(targets) == 0 {
		return nil
	}

	sem := make(chan struct{}, p.opts.Concurrency)
	results := make([]Result, len(targets))
	var wg sync.WaitGroup

	for i, t := range targets {
		wg.Add(1)
		go func(idx int, tgt Target) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[idx] = p.Probe(ctx, tgt.Host, tgt.Port)
		}(i, t)
	}

	wg.Wait()
	return results
}

// Target represents a host:port pair to probe.
type Target struct {
	Host string
	Port int
}
