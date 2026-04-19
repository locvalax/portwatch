package report

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/user/portwatch/internal/store"
)

// Builder constructs a historical report from store entries.
type Builder struct {
	w      io.Writer
	entries []store.Entry
}

// NewBuilder creates a new history report builder.
func NewBuilder(w io.Writer, entries []store.Entry) *Builder {
	return &Builder{w: w, entries: entries}
}

// WriteSummary writes a human-readable summary of scan history.
func (b *Builder) WriteSummary() error {
	if len(b.entries) == 0 {
		_, err := fmt.Fprintln(b.w, "No scan history available.")
		return err
	}

	// Group entries by host
	byHost := make(map[string][]store.Entry)
	for _, e := range b.entries {
		byHost[e.Host] = append(byHost[e.Host], e)
	}

	hosts := make([]string, 0, len(byHost))
	for h := range byHost {
		hosts = append(hosts, h)
	}
	sort.Strings(hosts)

	for _, host := range hosts {
		entries := byHost[host]
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Timestamp.Before(entries[j].Timestamp)
		})
		fmt.Fprintf(b.w, "Host: %s (%d scans)\n", host, len(entries))
		for _, e := range entries {
			fmt.Fprintf(b.w, "  [%s] ports: %s\n",
				e.Timestamp.Format(time.RFC3339),
				formatPorts(e.Ports),
			)
		}
		fmt.Fprintln(b.w)
	}
	return nil
}

// WriteChangelog writes only entries where ports changed between scans.
func (b *Builder) WriteChangelog() error {
	if len(b.entries) == 0 {
		_, err := fmt.Fprintln(b.w, "No scan history available.")
		return err
	}

	byHost := make(map[string][]store.Entry)
	for _, e := range b.entries {
		byHost[e.Host] = append(byHost[e.Host], e)
	}

	hosts := make([]string, 0, len(byHost))
	for h := range byHost {
		hosts = append(hosts, h)
	}
	sort.Strings(hosts)

	for _, host := range hosts {
		entries := byHost[host]
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Timestamp.Before(entries[j].Timestamp)
		})
		printed := false
		for i := 1; i < len(entries); i++ {
			prev := toSet(entries[i-1].Ports)
			curr := toSet(entries[i].Ports)
			if setsEqual(prev, curr) {
				continue
			}
			if !printed {
				fmt.Fprintf(b.w, "Host: %s\n", host)
				printed = true
			}
			fmt.Fprintf(b.w, "  [%s] %s -> %s\n",
				entries[i].Timestamp.Format(time.RFC3339),
				formatPorts(entries[i-1].Ports),
				formatPorts(entries[i].Ports),
			)
		}
		if printed {
			fmt.Fprintln(b.w)
		}
	}
	return nil
}

func toSet(ports []int) map[int]struct{} {
	s := make(map[int]struct{}, len(ports))
	for _, p := range ports {
		s[p] = struct{}{}
	}
	return s
}

func setsEqual(a, b map[int]struct{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}

func formatPorts(ports []int) string {
	if len(ports) == 0 {
		return "(none)"
	}
	parts := make([]string, len(ports))
	for i, p := range ports {
		parts[i] = fmt.Sprintf("%d", p)
	}
	return strings.Join(parts, ", ")
}
