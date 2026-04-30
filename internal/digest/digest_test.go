package digest

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/store"
)

var fixedNow = time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

func fixedClock() time.Time { return fixedNow }

func makeEntries(host string, ports []int, age time.Duration) store.Entry {
	return store.Entry{
		Host:      host,
		Ports:     ports,
		ScannedAt: fixedNow.Add(-age),
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if opts.Window != 24*time.Hour {
		t.Fatalf("expected 24h window, got %v", opts.Window)
	}
	if opts.Clock == nil {
		t.Fatal("expected non-nil clock")
	}
}

func TestSummarise_ReturnsLatestInWindow(t *testing.T) {
	d := New(Options{Window: 1 * time.Hour, Clock: fixedClock})
	entries := []store.Entry{
		makeEntries("host-a", []int{80, 443}, 30*time.Minute),
		makeEntries("host-a", []int{80, 443, 8080}, 10*time.Minute),
	}
	s, err := d.Summarise("host-a", entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Ports) != 3 {
		t.Fatalf("expected 3 ports, got %d", len(s.Ports))
	}
	if s.Host != "host-a" {
		t.Fatalf("expected host-a, got %s", s.Host)
	}
}

func TestSummarise_IgnoresOutOfWindowEntries(t *testing.T) {
	d := New(Options{Window: 1 * time.Hour, Clock: fixedClock})
	entries := []store.Entry{
		makeEntries("host-b", []int{22}, 2*time.Hour),
	}
	_, err := d.Summarise("host-b", entries)
	if err == nil {
		t.Fatal("expected error for out-of-window entries")
	}
}

func TestSummarise_UnknownHost(t *testing.T) {
	d := New(Options{Window: 1 * time.Hour, Clock: fixedClock})
	_, err := d.Summarise("ghost", nil)
	if err == nil {
		t.Fatal("expected error for unknown host")
	}
}

func TestSummarise_HashIsStable(t *testing.T) {
	d := New(Options{Window: 1 * time.Hour, Clock: fixedClock})
	entry := makeEntries("host-c", []int{443, 80}, 5*time.Minute)
	s1, _ := d.Summarise("host-c", []store.Entry{entry})
	s2, _ := d.Summarise("host-c", []store.Entry{entry})
	if s1.Hash != s2.Hash {
		t.Fatalf("hash not stable: %s != %s", s1.Hash, s2.Hash)
	}
}

func TestSummarise_DifferentPortsProduceDifferentHashes(t *testing.T) {
	d := New(Options{Window: 1 * time.Hour, Clock: fixedClock})
	e1 := makeEntries("host-d", []int{80}, 5*time.Minute)
	e2 := makeEntries("host-d", []int{443}, 5*time.Minute)
	s1, _ := d.Summarise("host-d", []store.Entry{e1})
	s2, _ := d.Summarise("host-d", []store.Entry{e2})
	if s1.Hash == s2.Hash {
		t.Fatal("expected different hashes for different ports")
	}
}
