package report_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/user/portwatch/internal/report"
	"github.com/user/portwatch/internal/store"
)

func sampleEntries() []report.Entry {
	return []report.Entry{
		{
			Host:      "localhost",
			Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Ports: []store.Result{
				{Port: 80, State: "open"},
				{Port: 443, State: "open"},
			},
		},
	}
}

func TestNew_DefaultsToStdout(t *testing.T) {
	r := report.New(nil, "")
	if r == nil {
		t.Fatal("expected non-nil reporter")
	}
}

func TestWrite_TextFormat(t *testing.T) {
	var buf bytes.Buffer
	r := report.New(&buf, report.FormatText)
	if err := r.Write(sampleEntries()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "localhost") {
		t.Error("expected host in output")
	}
	if !strings.Contains(out, ":80") {
		t.Error("expected port 80 in output")
	}
}

func TestWrite_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	r := report.New(&buf, report.FormatJSON)
	if err := r.Write(sampleEntries()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var decoded []report.Entry
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(decoded) != 1 || decoded[0].Host != "localhost" {
		t.Error("unexpected decoded entries")
	}
}

func TestWrite_Empty(t *testing.T) {
	var buf bytes.Buffer
	r := report.New(&buf, report.FormatText)
	if err := r.Write(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Error("expected empty output for no entries")
	}
}
