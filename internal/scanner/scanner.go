package scanner

import (
	"fmt"
	"net"
	"sort"
	"time"
)

// Result holds the scan result for a single host.
type Result struct {
	Host      string
	OpenPorts []int
	ScannedAt time.Time
}

// Options configures a port scan.
type Options struct {
	Timeout    time.Duration
	Concurrency int
}

// DefaultOptions returns sensible scan defaults.
func DefaultOptions() Options {
	return Options{
		Timeout:     500 * time.Millisecond,
		Concurrency: 100,
	}
}

// Scan checks the given ports on host and returns open ones.
func Scan(host string, ports []int, opts Options) (Result, error) {
	if len(ports) == 0 {
		return Result{}, fmt.Errorf("no ports specified")
	}

	type work struct{ port int }
	type hit struct{ port int }

	jobs := make(chan work, len(ports))
	results := make(chan hit, len(ports))

	worker := func() {
		for w := range jobs {
			addr := fmt.Sprintf("%s:%d", host, w.port)
			conn, err := net.DialTimeout("tcp", addr, opts.Timeout)
			if err == nil {
				conn.Close()
				results <- hit{port: w.port}
			}
		}
	}

	conc := opts.Concurrency
	if conc <= 0 {
		conc = 1
	}
	for i := 0; i < conc; i++ {
		go worker()
	}

	for _, p := range ports {
		jobs <- work{port: p}
	}
	close(jobs)

	// collect — we need a done signal; use a simple counter via channel drain
	open := make([]int, 0)
	for i := 0; i < len(ports); i++ {
		select {
		case h := <-results:
			open = append(open, h.port)
		default:
		}
	}

	// drain remaining
	close(results)
	for h := range results {
		open = append(open, h.port)
	}

	sort.Ints(open)
	return Result{
		Host:      host,
		OpenPorts: open,
		ScannedAt: time.Now().UTC(),
	}, nil
}
