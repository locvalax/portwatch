package batch

import (
	"context"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"github.com/user/portwatch/internal/scanner"
	"github.com/user/portwatch/internal/store"
)

// RunCmd performs a batch scan of hosts, prints a summary table, and
// persists each result to st. It is the entry-point called by main.
func RunCmd(ctx context.Context, hosts []string, st *store.Store, opts Options, w io.Writer) error {
	if w == nil {
		w = os.Stdout
	}

	sc := ScanFunc(func(ctx context.Context, host string, o scanner.Options) (store.Entry, error) {
		return scanner.Scan(ctx, host, o)
	})

	results := Run(ctx, hosts, sc, opts)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "HOST\tPORTS\tSTATUS")

	for _, r := range results {
		if r.Err != nil {
			fmt.Fprintf(tw, "%s\t-\tERROR: %v\n", r.Host, r.Err)
			continue
		}
		r.Entry.ScannedAt = time.Now()
		if err := st.Append(r.Host, r.Entry); err != nil {
			fmt.Fprintf(tw, "%s\t-\tSTORE ERROR: %v\n", r.Host, err)
			continue
		}
		fmt.Fprintf(tw, "%s\t%v\tOK\n", r.Host, r.Entry.Ports)
	}

	return tw.Flush()
}
