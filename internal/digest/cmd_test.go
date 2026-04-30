package digest

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/user/portwatch/internal/store"
)

func tempStore(t *testing.T) *store.Store {
	t.Helper()
	dir := t.TempDir()
	s, err := store.New(filepath.Join(dir, "portwatch.db"))
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestRun_TextFormat(t *testing.T) {
	s := tempStore(t)
	err := s.Append(store.Entry{
		Host:      "web-01",
		Ports:     []int{80, 443},
		ScannedAt: time.Now().Add(-10 * time.Minute),
	})
	if err != nil {
		t.Fatalf("append: %v", err)
	}

	var buf bytes.Buffer
	err = Run(s, RunArgs{
		Host:   "web-01",
		Window: 1 * time.Hour,
		Format: "text",
		Out:    &buf,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(buf.String(), "web-01") {
		t.Errorf("expected host in output, got: %s", buf.String())
	}
}

func TestRun_JSONFormat(t *testing.T) {
	s := tempStore(t)
	_ = s.Append(store.Entry{
		Host:      "db-01",
		Ports:     []int{5432},
		ScannedAt: time.Now().Add(-5 * time.Minute),
	})

	var buf bytes.Buffer
	err := Run(s, RunArgs{Host: "db-01", Window: 1 * time.Hour, Format: "json", Out: &buf})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(buf.String(), "\"Host\"") && !strings.Contains(buf.String(), "db-01") {
		t.Errorf("expected JSON output, got: %s", buf.String())
	}
}

func TestRun_UnknownHost(t *testing.T) {
	s := tempStore(t)
	err := Run(s, RunArgs{Host: "ghost", Window: 1 * time.Hour, Out: os.Discard})
	if err == nil {
		t.Fatal("expected error for unknown host")
	}
}

func TestRun_DefaultsToStdout(t *testing.T) {
	s := tempStore(t)
	_ = s.Append(store.Entry{
		Host:      "svc",
		Ports:     []int{8080},
		ScannedAt: time.Now().Add(-1 * time.Minute),
	})
	// Out is nil — should not panic
	err := Run(s, RunArgs{Host: "svc", Window: 1 * time.Hour})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
