package healthcheck

import (
	"context"
	"fmt"
	"net"
	"time"
)

// Status represents the reachability state of a host.
type Status struct {
	Host      string
	Reachable bool
	Latency   time.Duration
	Err       error
}

// Checker probes hosts for basic TCP reachability before scanning.
type Checker struct {
	timeout    time.Duration
	probePort  int
}

// Options configures a Checker.
type Options struct {
	Timeout   time.Duration
	ProbePort int
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		Timeout:   3 * time.Second,
		ProbePort: 80,
	}
}

// New creates a Checker with the given options.
func New(opts Options) *Checker {
	if opts.Timeout <= 0 {
		opts.Timeout = DefaultOptions().Timeout
	}
	if opts.ProbePort <= 0 {
		opts.ProbePort = DefaultOptions().ProbePort
	}
	return &Checker{
		timeout:   opts.Timeout,
		probePort: opts.ProbePort,
	}
}

// Probe checks whether a single host is reachable.
func (c *Checker) Probe(ctx context.Context, host string) Status {
	addr := fmt.Sprintf("%s:%d", host, c.probePort)
	start := time.Now()

	conn, err := (&net.Dialer{Timeout: c.timeout}).DialContext(ctx, "tcp", addr)
	latency := time.Since(start)

	if err != nil {
		return Status{Host: host, Reachable: false, Latency: latency, Err: err}
	}
	_ = conn.Close()
	return Status{Host: host, Reachable: true, Latency: latency}
}

// ProbeAll checks all hosts concurrently and returns their statuses.
func (c *Checker) ProbeAll(ctx context.Context, hosts []string) []Status {
	results := make([]Status, len(hosts))
	type indexed struct {
		i int
		s Status
	}
	ch := make(chan indexed, len(hosts))

	for i, h := range hosts {
		go func(idx int, host string) {
			ch <- indexed{i: idx, s: c.Probe(ctx, host)}
		}(i, h)
	}

	for range hosts {
		r := <-ch
		results[r.i] = r.s
	}
	return results
}
