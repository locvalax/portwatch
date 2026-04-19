package audit

import (
	"os"
	"strings"
	"testing"
)

func TestFileLogger_CreatesAndWrites(t *testing.T) {
	tmp, err := os.CreateTemp("", "audit-*.log")
	if err != nil {
		t.Fatal(err)
	}
	tmp.Close()
	defer os.Remove(tmp.Name())

	l, f, err := FileLogger(tmp.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()

	if err := l.Log("host1", "test", "details"); err != nil {
		t.Fatalf("log error: %v", err)
	}
	f.Close()

	data, err := os.ReadFile(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "host1") {
		t.Error("expected host1 in log file")
	}
}

func TestFileLogger_BadPath_ReturnsError(t *testing.T) {
	_, _, err := FileLogger("/nonexistent/dir/audit.log")
	if err == nil {
		t.Error("expected error for bad path")
	}
}
