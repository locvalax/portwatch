package resolve

import (
	"fmt"
	"net"
	"time"
)

// Result holds DNS resolution details for a host.
type Result struct {
	Host      string
	Addresses []string
	ResolvedAt time.Time
}

// Resolver resolves hostnames to IP addresses.
type Resolver struct {
	timeout time.Duration
}

// New returns a Resolver with the given timeout.
func New(timeout time.Duration) *Resolver {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &Resolver{timeout: timeout}
}

// Resolve performs a DNS lookup for the given host.
func (r *Resolver) Resolve(host string) (*Result, error) {
	resolver := &net.Resolver{}
	addrs, err := resolver.LookupHost(netCtx(r.timeout), host)
	if err != nil {
		return nil, fmt.Errorf("resolve %q: %w", host, err)
	}
	return &Result{
		Host:       host,
		Addresses:  addrs,
		ResolvedAt: time.Now(),
	}, nil
}

// ResolveAll resolves multiple hosts, returning results and any errors keyed by host.
func (r *Resolver) ResolveAll(hosts []string) (map[string]*Result, map[string]error) {
	results := make(map[string]*Result, len(hosts))
	errors := make(map[string]error)
	for _, h := range hosts {
		res, err := r.Resolve(h)
		if err != nil {
			errors[h] = err
		} else {
			results[h] = res
		}
	}
	return results, errors
}
