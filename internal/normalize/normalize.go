// Package normalize standardises host strings before they are used as
// store or cache keys, ensuring that equivalent addresses (e.g. an IPv6
// loopback written with or without brackets) are treated identically.
package normalize

import (
	"fmt"
	"net"
	"strings"
)

// Options controls normalisation behaviour.
type Options struct {
	// Lowercase converts the host to lower-case before returning it.
	Lowercase bool
	// StripPort removes a trailing ":port" suffix when present.
	StripPort bool
}

// DefaultOptions returns a sensible default configuration.
func DefaultOptions() Options {
	return Options{
		Lowercase: true,
		StripPort: false,
	}
}

// Normalizer transforms host strings according to the configured options.
type Normalizer struct {
	opts Options
}

// New creates a Normalizer with the supplied options.
func New(opts Options) *Normalizer {
	return &Normalizer{opts: opts}
}

// Host normalises a single host string and returns the result.
// It strips surrounding brackets from IPv6 literals and optionally
// removes a port suffix and converts the result to lower-case.
func (n *Normalizer) Host(raw string) (string, error) {
	if raw == "" {
		return "", fmt.Errorf("normalize: empty host")
	}

	host := raw

	// If the string contains a port, split it off.
	if n.opts.StripPort {
		h, _, err := net.SplitHostPort(raw)
		if err == nil {
			host = h
		}
		// If SplitHostPort fails the string has no port — use it as-is.
	}

	// Remove brackets from bare IPv6 literals such as "[::1]".
	host = strings.TrimPrefix(host, "[")
	host = strings.TrimSuffix(host, "]")

	if n.opts.Lowercase {
		host = strings.ToLower(host)
	}

	if host == "" {
		return "", fmt.Errorf("normalize: host reduced to empty string from %q", raw)
	}

	return host, nil
}

// Hosts normalises a slice of host strings, returning the first error
// encountered or a slice of normalised values in the same order.
func (n *Normalizer) Hosts(raws []string) ([]string, error) {
	out := make([]string, 0, len(raws))
	for _, r := range raws {
		h, err := n.Host(r)
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, nil
}
