package export_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/portwatch/internal/export"
	"github.com/user/portwatch/internal/store"
)

func tempStore(t *testing.T) *store.Store {
	t.Helper()
	dir := t.TempDir()
	s, err := store.New(filepath.Join(dir, "data.db"))
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestRun_ToStdout(t *testing.T) {
	s := tempStore(t)
	for _, e := range sampleEntries() {
		if err := s.Append(e); err != nil {
			t.Fatalf("append: %v", err)
		}
	}
	err := export.Run(s, export.Options{
		Format: export.FormatJSON,
		Output: "-",
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
}

func TestRun_ToFile(t *testing.T) {
	s := tempStore(t)
	for _, e := range sampleEntries() {
		_ = s.Append(e)
	}
	out := filepath.Join(t.TempDir(), "out.csv")
	err := export.Run(s, export.Options{
		Format: export.FormatCSV,
		Output: out,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Errorf("output file not created: %v", err)
	}
}

func TestRun_UnknownHost(t *testing.T) {
	s := tempStore(t)
	err := export.Run(s, export.Options{
		Host:   "ghost",
		Format: export.FormatJSON,
		Output: "-",
	})
	if err == nil {
		t.Error("expected error for unknown host")
	}
}
