package audit

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestLog_WritesJSON(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	if err := l.Log("192.168.1.1", "scan", "completed"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var e Event
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &e); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if e.Host != "192.168.1.1" {
		t.Errorf("expected host 192.168.1.1, got %s", e.Host)
	}
	if e.Action != "scan" {
		t.Errorf("expected action scan, got %s", e.Action)
	}
	if e.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestLog_DefaultsToStdout(t *testing.T) {
	l := New(nil)
	if l.w == nil {
		t.Error("expected non-nil writer")
	}
}

func TestLogPortChange_Opened(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	if err := l.LogPortChange("10.0.0.1", []uint16{80, 443}, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "ports_opened") {
		t.Error("expected ports_opened action in output")
	}
}

func TestLogPortChange_Closed(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	if err := l.LogPortChange("10.0.0.1", nil, []uint16{22}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "ports_closed") {
		t.Error("expected ports_closed action in output")
	}
}

func TestLogPortChange_NoChanges_NoOutput(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	if err := l.LogPortChange("10.0.0.1", nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Error("expected no output when no changes")
	}
}
