// Package batch provides concurrent scanning of multiple hosts with
// bounded parallelism and per-host result collection.
package batch

import (
	"context"
	"sync"

	"github.com/user/portwatch/internal/scanner"
	"github.com/user/portwatch/internal/store"
)

// Result holds the outcome of scanning a single host.
type Result struct {
	Host  string
	Entry store.Entry
	Err   error
}

// Options controls batch scan behaviour.
type Options struct {
	// Workers is the maximum number of concurrent scans. Defaults to 5.
	Workers int
	// ScanOptions are forwarded to scanner.Scan.
	ScanOptions scanner.Options
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		Workers:     5,
		ScanOptions: scanner.DefaultOptions(),
	}
}

// Scanner is the interface satisfied by scanner.Scan and middleware wrappers.
type Scanner interface {
	Scan(ctx context.Context, host string, opts scanner.Options) (store.Entry, error)
}

// ScanFunc is an adapter so plain functions satisfy Scanner.
type ScanFunc func(ctx context.Context, host string, opts scanner.Options) (store.Entry, error)

func (f ScanFunc) Scan(ctx context.Context, host string, opts scanner.Options) (store.Entry, error) {
	return f(ctx, host, opts)
}

// Run scans all hosts concurrently, respecting the worker limit, and returns
// one Result per host in the same order as the input slice.
func Run(ctx context.Context, hosts []string, sc Scanner, opts Options) []Result {
	if opts.Workers <= 0 {
		opts.Workers = DefaultOptions().Workers
	}

	results := make([]Result, len(hosts))
	sem := make(chan struct{}, opts.Workers)
	var wg sync.WaitGroup

	for i, host := range hosts {
		wg.Add(1)
		go func(idx int, h string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			entry, err := sc.Scan(ctx, h, opts.ScanOptions)
			results[idx] = Result{Host: h, Entry: entry, Err: err}
		}(i, host)
	}

	wg.Wait()
	return results
}
