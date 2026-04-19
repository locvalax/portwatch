package report

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/user/portwatch/internal/store"
)

func makeEntries() []store.Entry {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	return []store.Entry{
		{Host: "host-a", Ports: []int{80, 443}, Timestamp: base},
		{Host: "host-a", Ports: []int{80, 443, 8080}, Timestamp: base.Add(time.Hour)},
		{Host: "host-b", Ports: []int{22}, Timestamp: base},
		{Host: "host-b", Ports: []int{22}, Timestamp: base.Add(time.Hour)},
	}
}

func TestWriteSummary_ListsAllHosts(t *testing.T) {
	var buf bytes.Buffer
	b := NewBuilder(&buf, makeEntries())
	if err := b.WriteSummary(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "host-a (2 scans)") {
		t.Errorf("expected host-a scan count, got:\n%s", out)
	}
	if !strings.Contains(out, "host-b (2 scans)") {
		t.Errorf("expected host-b scan count, got:\n%s", out)
	}
	if !strings.Contains(out, "80, 443, 8080") {
		t.Errorf("expected port list, got:\n%s", out)
	}
}

func TestWriteSummary_Empty(t *testing.T) {
	var buf bytes.Buffer
	b := NewBuilder(&buf, nil)
	if err := b.WriteSummary(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "No scan history") {
		t.Errorf("expected empty message, got: %s", buf.String())
	}
}

func TestWriteChangelog_OnlyShowsChanges(t *testing.T) {
	var buf bytes.Buffer
	b := NewBuilder(&buf, makeEntries())
	if err := b.WriteChangelog(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	// host-a changed, host-b did not
	if !strings.Contains(out, "host-a") {
		t.Errorf("expected host-a in changelog, got:\n%s", out)
	}
	if strings.Contains(out, "host-b") {
		t.Errorf("host-b should not appear (no changes), got:\n%s", out)
	}
	if !strings.Contains(out, "80, 443 ->") {
		t.Errorf("expected port transition arrow, got:\n%s", out)
	}
}

func TestWriteChangelog_Empty(t *testing.T) {
	var buf bytes.Buffer
	b := NewBuilder(&buf, []store.Entry{})
	if err := b.WriteChangelog(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "No scan history") {
		t.Errorf("expected empty message")
	}
}
