package snapshot

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"github.com/user/portwatch/internal/store"
)

// CaptureOptions controls the Capture command.
type CaptureOptions struct {
	Host    string
	Dir     string
	Store   *store.Store
	Out     io.Writer
}

// Capture saves a snapshot for the given host and prints confirmation.
func Capture(opts CaptureOptions) error {
	if opts.Out == nil {
		opts.Out = os.Stdout
	}

	m, err := New(opts.Dir)
	if err != nil {
		return err
	}

	if err := m.Save(opts.Host, opts.Store); err != nil {
		return err
	}

	fmt.Fprintf(opts.Out, "snapshot saved for %s at %s\n", opts.Host, time.Now().Format(time.RFC3339))
	return nil
}

// ShowOptions controls the Show command.
type ShowOptions struct {
	Host string
	Dir  string
	Out  io.Writer
}

// Show prints the stored snapshot for host in a human-readable table.
func Show(opts ShowOptions) error {
	if opts.Out == nil {
		opts.Out = os.Stdout
	}

	m, err := New(opts.Dir)
	if err != nil {
		return err
	}

	snap, err := m.Load(opts.Host)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(opts.Out, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Host:\t%s\n", snap.Host)
	fmt.Fprintf(w, "Captured:\t%s\n", snap.CapturedAt.Format(time.RFC3339))
	fmt.Fprintf(w, "Open ports:\t%d\n", len(snap.Ports))
	for _, p := range snap.Ports {
		fmt.Fprintf(w, "  -\t%d\n", p)
	}
	w.Flush()
	return nil
}
