// Package priority assigns severity levels to port change events
// based on configurable port-to-priority mappings.
package priority

import "sort"

// Level represents the severity of a port change event.
type Level int

const (
	Low    Level = iota // default for unmapped ports
	Medium              // notable service ports
	High                // critical / well-known service ports
	Critical            // explicitly flagged ports
)

func (l Level) String() string {
	switch l {
	case Critical:
		return "critical"
	case High:
		return "high"
	case Medium:
		return "medium"
	default:
		return "low"
	}
}

// Options controls how priorities are assigned.
type Options struct {
	// Critical ports always surface immediately.
	CriticalPorts []int
	// High-priority ports (e.g. 22, 443, 3306).
	HighPorts []int
	// Medium-priority ports.
	MediumPorts []int
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		CriticalPorts: []int{},
		HighPorts:     []int{22, 23, 3389, 5900},
		MediumPorts:   []int{80, 443, 8080, 8443, 3306, 5432, 6379, 27017},
	}
}

// Ranker maps ports to priority levels.
type Ranker struct {
	index map[int]Level
}

// New builds a Ranker from the provided options.
func New(opts Options) *Ranker {
	idx := make(map[int]Level)
	for _, p := range opts.MediumPorts {
		idx[p] = Medium
	}
	for _, p := range opts.HighPorts {
		idx[p] = High
	}
	for _, p := range opts.CriticalPorts {
		idx[p] = Critical
	}
	return &Ranker{index: idx}
}

// Rank returns the highest priority level found among the given ports.
func (r *Ranker) Rank(ports []int) Level {
	max := Low
	for _, p := range ports {
		if lvl, ok := r.index[p]; ok && lvl > max {
			max = lvl
		}
	}
	return max
}

// Sort orders ports by their assigned priority level descending,
// then by port number ascending for stability.
func (r *Ranker) Sort(ports []int) []int {
	out := make([]int, len(ports))
	copy(out, ports)
	sort.SliceStable(out, func(i, j int) bool {
		li := r.index[out[i]]
		lj := r.index[out[j]]
		if li != lj {
			return li > lj
		}
		return out[i] < out[j]
	})
	return out
}
