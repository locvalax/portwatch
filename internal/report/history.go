package report

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/user/portwatch/internal/store"
)

// Builder accumulates scan entries and renders historical summaries.
type Builder struct {
	w io.Writer
}

// NewBuilder returns a Builder that writes to w.
// If w is nil, os.Stdout is used.
func NewBuilder(w io.Writer) *Builder {
	if w == nil {
		w = os.Stdout
	}
	return &Builder{w: w}
}

// Summary holds aggregated stats for a single host across all stored scans.
type Summary struct {
	Host      string
	ScanCount int
	FirstSeen time.Time
	LastSeen  time.Time
	// UniquePorts is the union of all ports ever observed open.
	UniquePorts []uint16
}

// Build computes per-host summaries from the provided entries.
func (b *Builder) Build(entries []store.Entry) []Summary {
	type hostData struct {
		count int
		first time.Time
		last  time.Time
		ports map[uint16]struct{}
	}

	m := make(map[string]*hostData)
	for _, e := range entries {
		hd, ok := m[e.Host]
		if !ok {
			hd = &hostData{
				first: e.ScannedAt,
				last:  e.ScannedAt,
				ports: make(map[uint16]struct{}),
			}
			m[e.Host] = hd
		}
		hd.count++
		if e.ScannedAt.Before(hd.first) {
			hd.first = e.ScannedAt
		}
		if e.ScannedAt.After(hd.last) {
			hd.last = e.ScannedAt
		}
		for _, p := range e.Ports {
			hd.ports[p] = struct{}{}
		}
	}

	summaries := make([]Summary, 0, len(m))
	for host, hd := range m {
		ports := make([]uint16, 0, len(hd.ports))
		for p := range hd.ports {
			ports = append(ports, p)
		}
		sort.Slice(ports, func(i, j int) bool { return ports[i] < ports[j] })
		summaries = append(summaries, Summary{
			Host:        host,
			ScanCount:   hd.count,
			FirstSeen:   hd.first,
			LastSeen:    hd.last,
			UniquePorts: ports,
		})
	}
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Host < summaries[j].Host
	})
	return summaries
}

// Write renders the summaries as a human-readable table.
func (b *Builder) Write(summaries []Summary) error {
	if len(summaries) == 0 {
		_, err := fmt.Fprintln(b.w, "no history available")
		return err
	}

	tw := tabwriter.NewWriter(b.w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "HOST\tSCANS\tFIRST SEEN\tLAST SEEN\tUNIQUE PORTS")
	fmt.Fprintln(tw, strings.Repeat("-", 72))
	const timeFmt = "2006-01-02 15:04"
	for _, s := range summaries {
		ports := formatPorts(s.UniquePorts)
		fmt.Fprintf(tw, "%s\t%d\t%s\t%s\t%s\n",
			s.Host,
			s.ScanCount,
			s.FirstSeen.Format(timeFmt),
			s.LastSeen.Format(timeFmt),
			ports,
		)
	}
	return tw.Flush()
}

// formatPorts converts a sorted slice of port numbers to a compact string.
func formatPorts(ports []uint16) string {
	if len(ports) == 0 {
		return "none"
	}
	parts := make([]string, len(ports))
	for i, p := range ports {
		parts[i] = fmt.Sprintf("%d", p)
	}
	return strings.Join(parts, ", ")
}
