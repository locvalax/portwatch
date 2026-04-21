package fingerprint_test

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/fingerprint"
	"github.com/user/portwatch/internal/store"
)

func makeEntry(host string, ports []int) store.Entry {
	return store.Entry{
		Host:      host,
		Ports:     ports,
		ScannedAt: time.Now(),
	}
}

func TestSum_StableAcrossPortOrder(t *testing.T) {
	h := fingerprint.New()
	a := makeEntry("10.0.0.1", []int{80, 443, 22})
	b := makeEntry("10.0.0.1", []int{443, 22, 80})

	if h.Sum(a) != h.Sum(b) {
		t.Error("expected same fingerprint for same ports in different order")
	}
}

func TestSum_DifferentPorts_DifferentHash(t *testing.T) {
	h := fingerprint.New()
	a := makeEntry("10.0.0.1", []int{80, 443})
	b := makeEntry("10.0.0.1", []int{80, 8080})

	if h.Sum(a) == h.Sum(b) {
		t.Error("expected different fingerprints for different port sets")
	}
}

func TestSum_DifferentHosts_DifferentHash(t *testing.T) {
	h := fingerprint.New()
	a := makeEntry("10.0.0.1", []int{80})
	b := makeEntry("10.0.0.2", []int{80})

	if h.Sum(a) == h.Sum(b) {
		t.Error("expected different fingerprints for different hosts")
	}
}

func TestEqual_IdenticalEntries(t *testing.T) {
	h := fingerprint.New()
	a := makeEntry("host-a", []int{22, 80})
	b := makeEntry("host-a", []int{80, 22})

	if !h.Equal(a, b) {
		t.Error("expected entries to be equal")
	}
}

func TestEqual_DifferentHosts(t *testing.T) {
	h := fingerprint.New()
	a := makeEntry("host-a", []int{80})
	b := makeEntry("host-b", []int{80})

	if h.Equal(a, b) {
		t.Error("expected entries to be unequal due to different hosts")
	}
}

func TestSumPorts_EmptySlice(t *testing.T) {
	h := fingerprint.New()
	s := h.SumPorts([]int{})
	if s == "" {
		t.Error("expected non-empty hash for empty port slice")
	}
}

func TestSumPorts_ConsistentWithSum(t *testing.T) {
	h := fingerprint.New()
	ports := []int{443, 80, 22}
	e := makeEntry("irrelevant", ports)

	// SumPorts must equal the port-only component used internally;
	// we verify it is stable across two calls with the same input.
	if h.SumPorts(ports) != h.SumPorts([]int{22, 80, 443}) {
		t.Error("SumPorts should be order-independent")
	}

	// Ensure Sum changes when host changes but ports stay the same.
	other := makeEntry("other-host", ports)
	if h.Sum(e) == h.Sum(other) {
		t.Error("Sum must incorporate host")
	}
}
