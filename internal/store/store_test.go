package store

import (
	"os"
	"testing"
	"time"
)

func tempStore(t *testing.T) *Store {
	t.Helper()
	f, err := os.CreateTemp("", "portwatch-*.json")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	os.Remove(f.Name()) // start fresh (no file)
	t.Cleanup(func() { os.Remove(f.Name()) })
	return New(f.Name())
}

func TestStore_AppendAndLatest(t *testing.T) {
	s := tempStore(t)

	snap := Snapshot{Host: "localhost", Ports: []int{80, 443}, ScannedAt: time.Now()}
	if err := s.Append(snap); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	latest, err := s.Latest("localhost")
	if err != nil {
		t.Fatalf("Latest failed: %v", err)
	}
	if latest == nil {
		t.Fatal("expected a snapshot, got nil")
	}
	if len(latest.Ports) != 2 {
		t.Errorf("expected 2 ports, got %d", len(latest.Ports))
	}
}

func TestStore_Latest_UnknownHost(t *testing.T) {
	s := tempStore(t)
	latest, err := s.Latest("unknown")
	if err != nil {
		t.Fatal(err)
	}
	if latest != nil {
		t.Error("expected nil for unknown host")
	}
}

func TestCompare_OpensAndCloses(t *testing.T) {
	prev := &Snapshot{Host: "h", Ports: []int{22, 80}}
	diff := Compare("h", prev, []int{80, 443})

	if len(diff.Opened) != 1 || diff.Opened[0] != 443 {
		t.Errorf("expected opened [443], got %v", diff.Opened)
	}
	if len(diff.Closed) != 1 || diff.Closed[0] != 22 {
		t.Errorf("expected closed [22], got %v", diff.Closed)
	}
	if !diff.HasChanges() {
		t.Error("expected HasChanges to be true")
	}
}

func TestCompare_NoPrev(t *testing.T) {
	diff := Compare("h", nil, []int{8080})
	if len(diff.Opened) != 1 || diff.Opened[0] != 8080 {
		t.Errorf("expected opened [8080], got %v", diff.Opened)
	}
	if len(diff.Closed) != 0 {
		t.Errorf("expected no closed ports, got %v", diff.Closed)
	}
}
