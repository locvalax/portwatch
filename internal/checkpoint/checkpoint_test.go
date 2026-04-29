package checkpoint_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/user/portwatch/internal/checkpoint"
)

func tempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "checkpoint-*")
	if err != nil {
		t.Fatalf("tempDir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func TestSaveAndLoad(t *testing.T) {
	s, err := checkpoint.New(tempDir(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	e := checkpoint.Entry{Host: "192.168.1.1", Ports: []int{22, 80, 443}}
	if err := s.Save(e); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := s.Load("192.168.1.1")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Host != e.Host {
		t.Errorf("host: got %q, want %q", got.Host, e.Host)
	}
	if len(got.Ports) != len(e.Ports) {
		t.Errorf("ports len: got %d, want %d", len(got.Ports), len(e.Ports))
	}
	if got.RecordedAt.IsZero() {
		t.Error("RecordedAt should not be zero")
	}
}

func TestLoad_MissingHost_ReturnsErrNotFound(t *testing.T) {
	s, _ := checkpoint.New(tempDir(t))
	_, err := s.Load("unknown-host")
	if !errors.Is(err, checkpoint.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDelete_RemovesEntry(t *testing.T) {
	s, _ := checkpoint.New(tempDir(t))
	e := checkpoint.Entry{Host: "10.0.0.1", Ports: []int{8080}}
	s.Save(e)

	if err := s.Delete("10.0.0.1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := s.Load("10.0.0.1")
	if !errors.Is(err, checkpoint.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestDelete_UnknownHost_IsNoop(t *testing.T) {
	s, _ := checkpoint.New(tempDir(t))
	if err := s.Delete("no-such-host"); err != nil {
		t.Errorf("Delete unknown host should be noop, got %v", err)
	}
}

// stubScanner is a minimal Scanner for middleware tests.
type stubScanner struct {
	ports []int
	err   error
}

func (s *stubScanner) Scan(_ context.Context, _ string) ([]int, error) {
	return s.ports, s.err
}

func TestCheckpointedScanner_SavesOnSuccess(t *testing.T) {
	cp, _ := checkpoint.New(tempDir(t))
	stub := &stubScanner{ports: []int{22, 443}}
	cs := checkpoint.NewCheckpointedScanner(stub, cp)

	ports, err := cs.Scan(context.Background(), "host-a")
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(ports) != 2 {
		t.Errorf("ports: got %d, want 2", len(ports))
	}

	prev, err := cs.Previous("host-a")
	if err != nil {
		t.Fatalf("Previous: %v", err)
	}
	if len(prev) != 2 {
		t.Errorf("prev ports: got %d, want 2", len(prev))
	}
}

func TestCheckpointedScanner_InnerError_NoSave(t *testing.T) {
	cp, _ := checkpoint.New(tempDir(t))
	stub := &stubScanner{err: errors.New("scan failed")}
	cs := checkpoint.NewCheckpointedScanner(stub, cp)

	_, err := cs.Scan(context.Background(), "host-b")
	if err == nil {
		t.Fatal("expected error from inner scanner")
	}

	prev, _ := cs.Previous("host-b")
	if prev != nil {
		t.Errorf("expected no checkpoint after failed scan, got %v", prev)
	}
}
