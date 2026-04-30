// Package redact provides hostname and IP masking for scan results
// before they are written to logs, alerts, or exports.
package redact

import (
	"crypto/sha256"
	"fmt"
	"net"
	"strings"
	"sync"
)

// Mode controls how hosts are redacted.
type Mode int

const (
	// ModeHash replaces the host with a stable SHA-256 prefix.
	ModeHash Mode = iota
	// ModeMask replaces the host with a fixed placeholder.
	ModeMask
)

// Options configures the Redactor.
type Options struct {
	Mode   Mode
	Salt   string
	Length int // prefix length for ModeHash (default 8)
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		Mode:   ModeHash,
		Length: 8,
	}
}

// Redactor masks host identifiers in scan output.
type Redactor struct {
	opts  Options
	mu    sync.Mutex
	cache map[string]string
}

// New creates a Redactor with the given options.
func New(opts Options) *Redactor {
	if opts.Length <= 0 {
		opts.Length = 8
	}
	return &Redactor{
		opts:  opts,
		cache: make(map[string]string),
	}
}

// Host returns the redacted representation of host.
func (r *Redactor) Host(host string) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if v, ok := r.cache[host]; ok {
		return v
	}

	var out string
	switch r.opts.Mode {
	case ModeMask:
		out = "<redacted>"
		if ip := net.ParseIP(host); ip != nil {
			out = "<ip-redacted>"
		}
	default: // ModeHash
		input := r.opts.Salt + host
		sum := sha256.Sum256([]byte(input))
		out = fmt.Sprintf("host-%s", strings.ToLower(
			fmt.Sprintf("%x", sum[:])[:r.opts.Length]))
	}

	r.cache[host] = out
	return out
}

// Flush clears the internal cache.
func (r *Redactor) Flush() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache = make(map[string]string)
}
