package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/user/portwatch/internal/store"
)

// Format represents an export format.
type Format string

const (
	FormatCSV  Format = "csv"
	FormatJSON Format = "json"
)

// Exporter writes scan entries to an output stream.
type Exporter struct {
	w      io.Writer
	format Format
}

// New creates a new Exporter.
func New(w io.Writer, format Format) *Exporter {
	return &Exporter{w: w, format: format}
}

// Write exports the given entries in the configured format.
func (e *Exporter) Write(entries []store.Entry) error {
	switch e.format {
	case FormatCSV:
		return e.writeCSV(entries)
	case FormatJSON:
		return e.writeJSON(entries)
	default:
		return fmt.Errorf("unsupported format: %s", e.format)
	}
}

func (e *Exporter) writeCSV(entries []store.Entry) error {
	w := csv.NewWriter(e.w)
	_ = w.Write([]string{"host", "port", "proto", "scanned_at"})
	for _, en := range entries {
		for _, p := range en.Ports {
			_ = w.Write([]string{
				en.Host,
				fmt.Sprintf("%d", p.Port),
				p.Proto,
				en.ScannedAt.Format(time.RFC3339),
			})
		}
	}
	w.Flush()
	return w.Error()
}

func (e *Exporter) writeJSON(entries []store.Entry) error {
	enc := json.NewEncoder(e.w)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}
