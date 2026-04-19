package filter

import "strconv"

// Rule defines inclusion/exclusion criteria for ports.
type Rule struct {
	IncludePorts []int
	ExcludePorts []int
	MinPort      int
	MaxPort      int
}

// Filter applies rules to a list of ports and returns the filtered result.
type Filter struct {
	rule Rule
}

// New creates a Filter from the given Rule.
func New(r Rule) *Filter {
	if r.MaxPort == 0 {
		r.MaxPort = 65535
	}
	return &Filter{rule: r}
}

// Apply returns only the ports that pass the filter rules.
func (f *Filter) Apply(ports []int) []int {
	exclude := toSet(f.rule.ExcludePorts)
	include := toSet(f.rule.IncludePorts)

	var out []int
	for _, p := range ports {
		if p < f.rule.MinPort || p > f.rule.MaxPort {
			continue
		}
		if exclude[p] {
			continue
		}
		if len(include) > 0 && !include[p] {
			continue
		}
		out = append(out, p)
	}
	return out
}

// ParsePorts converts a slice of string port numbers to ints, ignoring invalid entries.
func ParsePorts(strs []string) []int {
	var ports []int
	for _, s := range strs {
		if p, err := strconv.Atoi(s); err == nil && p > 0 && p <= 65535 {
			ports = append(ports, p)
		}
	}
	return ports
}

func toSet(ports []int) map[int]bool {
	m := make(map[int]bool, len(ports))
	for _, p := range ports {
		m[p] = true
	}
	return m
}
