package correlation

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/store"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if opts.Window != 2*time.Minute {
		t.Fatalf("expected 2m window, got %v", opts.Window)
	}
	if opts.MinHosts != 2 {
		t.Fatalf("expected MinHosts=2, got %d", opts.MinHosts)
	}
}

func TestObserve_SingleHost_NoEvent(t *testing.T) {
	c := New(DefaultOptions())
	events := c.Observe("host-a", store.Diff{Opened: []int{80}})
	if len(events) != 0 {
		t.Fatalf("expected no events, got %d", len(events))
	}
}

func TestObserve_TwoHosts_SamePort_EmitsEvent(t *testing.T) {
	c := New(DefaultOptions())
	c.Observe("host-a", store.Diff{Opened: []int{443}})
	events := c.Observe("host-b", store.Diff{Opened: []int{443}})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].PortCount != 443 {
		t.Errorf("expected port 443, got %d", events[0].PortCount)
	}
	if len(events[0].Hosts) != 2 {
		t.Errorf("expected 2 hosts, got %d", len(events[0].Hosts))
	}
}

func TestObserve_StaleEntries_NotCorrelated(t *testing.T) {
	now := time.Now()
	c := New(Options{Window: 1 * time.Minute, MinHosts: 2})
	// inject a fixed clock so we can advance time
	c.clock = func() time.Time { return now }
	c.Observe("host-a", store.Diff{Opened: []int{22}})

	// advance past the window
	c.clock = func() time.Time { return now.Add(2 * time.Minute) }
	events := c.Observe("host-b", store.Diff{Opened: []int{22}})
	if len(events) != 0 {
		t.Fatalf("expected no events after window expiry, got %d", len(events))
	}
}

func TestObserve_DifferentPorts_NoEvent(t *testing.T) {
	c := New(DefaultOptions())
	c.Observe("host-a", store.Diff{Opened: []int{80}})
	events := c.Observe("host-b", store.Diff{Opened: []int{443}})
	if len(events) != 0 {
		t.Fatalf("expected no events for different ports, got %d", len(events))
	}
}

func TestObserve_BucketClearedAfterEvent(t *testing.T) {
	c := New(Options{Window: 5 * time.Minute, MinHosts: 2})
	c.Observe("host-a", store.Diff{Opened: []int{8080}})
	c.Observe("host-b", store.Diff{Opened: []int{8080}}) // fires event, clears bucket
	events := c.Observe("host-c", store.Diff{Opened: []int{8080}})
	// bucket was cleared, so host-c alone should not trigger a new event
	if len(events) != 0 {
		t.Fatalf("expected no event after bucket cleared, got %d", len(events))
	}
}
