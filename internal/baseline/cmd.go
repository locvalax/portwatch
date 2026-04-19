package baseline

import (
	"fmt"
	"io"
	"os"

	"github.com/user/portwatch/internal/scanner"
)

// CaptureOptions configures a baseline capture run.
type CaptureOptions struct {
	Host    string
	Path    string
	Out     io.Writer
	ScanOpt scanner.Options
}

// Capture scans the host and saves the result as a baseline.
func Capture(opts CaptureOptions) error {
	if opts.Out == nil {
		opts.Out = os.Stdout
	}
	ports, err := scanner.Scan(opts.Host, opts.ScanOpt)
	if err != nil {
		return fmt.Errorf("scan: %w", err)
	}
	m := New(opts.Path)
	snap := Snapshot{Host: opts.Host, Ports: ports}
	if err := m.Save(snap); err != nil {
		return fmt.Errorf("save baseline: %w", err)
	}
	fmt.Fprintf(opts.Out, "baseline saved: %s — %d port(s) open\n", opts.Host, len(ports))
	return nil
}

// Check loads the baseline and compares it against a fresh scan.
func Check(opts CaptureOptions) (Diff, error) {
	m := New(opts.Path)
	snap, err := m.Load()
	if err != nil {
		return Diff{}, fmt.Errorf("load baseline: %w", err)
	}
	current, err := scanner.Scan(opts.Host, opts.ScanOpt)
	if err != nil {
		return Diff{}, fmt.Errorf("scan: %w", err)
	}
	return Compare(snap, current), nil
}
