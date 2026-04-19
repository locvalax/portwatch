package report

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/user/portwatch/internal/store"
)

// Format represents the output format for reports.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// Entry holds a single report record.
type Entry struct {
	Host      string         `json:"host"`
	Timestamp time.Time      `json:"timestamp"`
	Ports     []store.Result `json:"ports"`
}

// Reporter writes scan history in a chosen format.
type Reporter struct {
	out    io.Writer
	format Format
}

// New creates a Reporter. If out is nil it defaults to os.Stdout.
func New(out io.Writer, format Format) *Reporter {
	if out == nil {
		out = os.Stdout
	}
	if format == "" {
		format = FormatText
	}
	return &Reporter{out: out, format: format}
}

// Write renders entries to the configured writer.
func (r *Reporter) Write(entries []Entry) error {
	switch r.format {
	case FormatJSON:
		return r.writeJSON(entries)
	default:
		return r.writeText(entries)
	}
}

func (r *Reporter) writeText(entries []Entry) error {
	for _, e := range entries {
		fmt.Fprintf(r.out, "[%s] %s\n", e.Timestamp.Format(time.RFC3339), e.Host)
		for _, p := range e.Ports {
			fmt.Fprintf(r.out, "  :%d %s\n", p.Port, p.State)
		}
	}
	return nil
}

func (r *Reporter) writeJSON(entries []Entry) error {
	enc := json.NewEncoder(r.out)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}
