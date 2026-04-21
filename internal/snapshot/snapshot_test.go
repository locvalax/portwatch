package snapshot_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/user/portwatch/internal/snapshot"
	"github.com/user/portwatch/internal/store"
)

func tempDir(t *testing.T) string {
	t.Helper()
	d, err := os.MkdirTemp("", "snapshot-test-*")
	if err != nil {
		t.Fatalf("tempDir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(d) })
	return d
}

func tempStore(t *testing.T, host string, ports []int) *store.Store {
	t.Helper()
	path := filepath.Join(tempDir(t), "store.json")
	st, err := store.New(path)
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	if err := st.Append(host, ports); err != nil {
		t.Fatalf("Append: %v", err)
	}
	return st
}

func TestSaveAndLoad(t *testing.T) {
	dir := tempDir(t)
	st := tempStore(t, "localhost", []int{22, 80, 443})

	m, err := snapshot.New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := m.Save("localhost", st); err != nil {
		t.Fatalf("Save: %v", err)
	}

	snap, err := m.Load("localhost")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if snap.Host != "localhost" {
		t.Errorf("host = %q, want %q", snap.Host, "localhost")
	}
	if len(snap.Ports) != 3 {
		t.Errorf("ports len = %d, want 3", len(snap.Ports))
	}
}

func TestExists(t *testing.T) {
	dir := tempDir(t)
	st := tempStore(t, "host1", []int{8080})

	m, _ := snapshot.New(dir)

	if m.Exists("host1") {
		t.Error("Exists returned true before Save")
	}
	if err := m.Save("host1", st); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if !m.Exists("host1") {
		t.Error("Exists returned false after Save")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	m, _ := snapshot.New(tempDir(t))
	_, err := m.Load("ghost")
	if err == nil {
		t.Error("expected error for missing snapshot, got nil")
	}
}

func TestShow_PrintsTable(t *testing.T) {
	dir := tempDir(t)
	st := tempStore(t, "myhost", []int{22, 443})
	m, _ := snapshot.New(dir)
	_ = m.Save("myhost", st)

	var buf bytes.Buffer
	err := snapshot.Show(snapshot.ShowOptions{
		Host: "myhost",
		Dir:  dir,
		Out:  &buf,
	})
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	out := buf.String()
	if !bytes.Contains([]byte(out), []byte("myhost")) {
		t.Errorf("output missing host name: %s", out)
	}
}
