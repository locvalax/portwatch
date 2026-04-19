package export

import (
	"fmt"
	"io"
	"os"

	"github.com/user/portwatch/internal/store"
)

// Options holds CLI-level export options.
type Options struct {
	Host   string
	Format Format
	Output string // file path or "-" for stdout
}

// Run executes an export using the given store and options.
func Run(s *store.Store, opts Options) error {
	var w io.Writer
	if opts.Output == "" || opts.Output == "-" {
		w = os.Stdout
	} else {
		f, err := os.Create(opts.Output)
		if err != nil {
			return fmt.Errorf("open output file: %w", err)
		}
		defer f.Close()
		w = f
	}

	var entries []store.Entry
	var err error
	if opts.Host != "" {
		var e store.Entry
		e, err = s.Latest(opts.Host)
		if err == nil {
			entries = []store.Entry{e}
		}
	} else {
		entries, err = s.All()
	}
	if err != nil {
		return fmt.Errorf("fetch entries: %w", err)
	}

	ex := New(w, opts.Format)
	return ex.Write(entries)
}
