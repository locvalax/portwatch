package export_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/user/portwatch/internal/export"
	"github.com/user/portwatch/internal/store"
)

func sampleEntries() []store.Entry {
	return []store.Entry{
		{
			Host:      "192.168.1.1",
			ScannedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Ports: []store.PortResult{
				{Port: 80, Proto: "tcp"},
				{Port: 443, Proto: "tcp"},
			},
		},
	}
}

func TestWrite_CSVFormat(t *testing.T) {
	var buf bytes.Buffer
	ex := export.New(&buf, export.FormatCSV)
	if err := ex.Write(sampleEntries()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "host") || !strings.Contains(out, "80") {
		t.Errorf("unexpected CSV output: %s", out)
	}
}

func TestWrite_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	ex := export.New(&buf, export.FormatJSON)
	if err := ex.Write(sampleEntries()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var out []store.Entry
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out) != 1 || out[0].Host != "192.168.1.1" {
		t.Errorf("unexpected JSON output: %+v", out)
	}
}

func TestWrite_UnknownFormat(t *testing.T) {
	var buf bytes.Buffer
	ex := export.New(&buf, export.Format("xml"))
	if err := ex.Write(sampleEntries()); err == nil {
		t.Error("expected error for unknown format")
	}
}

func TestWrite_Empty(t *testing.T) {
	var buf bytes.Buffer
	ex := export.New(&buf, export.FormatCSV)
	if err := ex.Write(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
