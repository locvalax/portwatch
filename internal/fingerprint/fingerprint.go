// Package fingerprint generates stable hashes for scan results,
// enabling fast equality checks without full port-list comparison.
package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/user/portwatch/internal/store"
)

// Hasher computes deterministic fingerprints for scan entries.
type Hasher struct{}

// New returns a new Hasher.
func New() *Hasher {
	return &Hasher{}
}

// Sum returns a hex-encoded SHA-256 fingerprint for the given entry.
// The hash is computed over the host and a sorted, deduplicated port list
// so that two entries with identical open ports always produce the same hash.
func (h *Hasher) Sum(e store.Entry) string {
	ports := make([]int, len(e.Ports))
	copy(ports, e.Ports)
	sort.Ints(ports)

	parts := make([]string, len(ports))
	for i, p := range ports {
		parts[i] = fmt.Sprintf("%d", p)
	}

	raw := fmt.Sprintf("%s|%s", e.Host, strings.Join(parts, ","))
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// Equal reports whether two entries have the same host and open-port set.
func (h *Hasher) Equal(a, b store.Entry) bool {
	if a.Host != b.Host {
		return false
	}
	return h.Sum(a) == h.Sum(b)
}

// SumPorts returns a fingerprint for a raw port slice, independent of a host.
// Useful when comparing partial scan results.
func (h *Hasher) SumPorts(ports []int) string {
	sorted := make([]int, len(ports))
	copy(sorted, ports)
	sort.Ints(sorted)

	parts := make([]string, len(sorted))
	for i, p := range sorted {
		parts[i] = fmt.Sprintf("%d", p)
	}

	sum := sha256.Sum256([]byte(strings.Join(parts, ",")))
	return hex.EncodeToString(sum[:])
}
