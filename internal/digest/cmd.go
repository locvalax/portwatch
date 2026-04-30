package digest

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/user/portwatch/internal/store"
)

// RunArgs holds parameters for the Digest CLI command.
type RunArgs struct {
	StorePath string
	Host      string
	Window    time.Duration
	Format    string // "text" or "json"
	Out       io.Writer
}

// Run executes the digest command, printing a summary for the given host.
func Run(s *store.Store, args RunArgs) error {
	if args.Out == nil {
		args.Out = os.Stdout
	}
	if args.Window <= 0 {
		args.Window = DefaultOptions().Window
	}

	entries, err := s.All(args.Host)
	if err != nil {
		return fmt.Errorf("digest: load entries: %w", err)
	}

	d := New(Options{Window: args.Window})
	sum, err := d.Summarise(args.Host, entries)
	if err != nil {
		return err
	}

	switch args.Format {
	case "json":
		return json.NewEncoder(args.Out).Encode(sum)
	default:
		fmt.Fprintf(args.Out, "host:  %s\n", sum.Host)
		fmt.Fprintf(args.Out, "hash:  %s\n", sum.Hash)
		fmt.Fprintf(args.Out, "ports: %v\n", sum.Ports)
		fmt.Fprintf(args.Out, "at:    %s\n", sum.CreatedAt.Format(time.RFC3339))
		return nil
	}
}
