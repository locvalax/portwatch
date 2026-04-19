package baseline

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "baseline.json")
}

func TestSaveAndLoad(t *testing.T) {
	m := New(tempPath(t))
	snap := Snapshot{Host: "localhost", Ports: []int{80, 443}}
	if err := m.Save(snap); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := m.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.Host != snap.Host {
		t.Errorf("host: got %s want %s", got.Host, snap.Host)
	}
	if len(got.Ports) != len(snap.Ports) {
		t.Errorf("ports len: got %d want %d", len(got.Ports), len(snap.Ports))
	}
	if got.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestExists(t *testing.T) {
	p := tempPath(t)
	m := New(p)
	if m.Exists() {
		t.Error("expected false before save")
	}
	_ = m.Save(Snapshot{Host: "h", Ports: nil, CreatedAt: time.Now()})
	if !m.Exists() {
		t.Error("expected true after save")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	m := New(filepath.Join(t.TempDir(), "nope.json"))
	_, err := m.Load()
	if !os.IsNotExist(err) {
		t.Errorf("expected not-exist error, got %v", err)
	}
}
