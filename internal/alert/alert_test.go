package alert_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/store"
)

func TestNotify_OpenedPorts(t *testing.T) {
	var buf bytes.Buffer
	n := alert.New(&buf)

	d := store.Diff{Opened: []int{8080, 443}}
	events := n.Notify("localhost", d)

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	for _, e := range events {
		if e.Level != alert.LevelAlert {
			t.Errorf("expected ALERT level, got %s", e.Level)
		}
	}
	if !strings.Contains(buf.String(), "OPEN") {
		t.Error("expected output to contain OPEN")
	}
}

func TestNotify_ClosedPorts(t *testing.T) {
	var buf bytes.Buffer
	n := alert.New(&buf)

	d := store.Diff{Closed: []int{22}}
	events := n.Notify("10.0.0.1", d)

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Level != alert.LevelWarn {
		t.Errorf("expected WARN level, got %s", events[0].Level)
	}
	if !strings.Contains(buf.String(), "CLOSED") {
		t.Error("expected output to contain CLOSED")
	}
}

func TestNotify_NoChanges(t *testing.T) {
	var buf bytes.Buffer
	n := alert.New(&buf)

	d := store.Diff{}
	events := n.Notify("host1", d)

	if len(events) != 1 {
		t.Fatalf("expected 1 info event, got %d", len(events))
	}
	if events[0].Level != alert.LevelInfo {
		t.Errorf("expected INFO level, got %s", events[0].Level)
	}
}

func TestNew_DefaultsToStdout(t *testing.T) {
	// Should not panic when nil writer is passed
	n := alert.New(nil)
	if n == nil {
		t.Fatal("expected non-nil notifier")
	}
}
