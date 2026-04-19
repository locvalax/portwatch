package baseline

// Diff holds ports that appeared or disappeared relative to a baseline.
type Diff struct {
	Opened []int
	Closed []int
}

// HasChanges returns true when any port changed.
func (d Diff) HasChanges() bool {
	return len(d.Opened) > 0 || len(d.Closed) > 0
}

// Compare returns the diff between a saved baseline and a current port list.
func Compare(baseline Snapshot, current []int) Diff {
	base := toSet(baseline.Ports)
	curr := toSet(current)

	var opened, closed []int
	for p := range curr {
		if !base[p] {
			opened = append(opened, p)
		}
	}
	for p := range base {
		if !curr[p] {
			closed = append(closed, p)
		}
	}
	return Diff{Opened: opened, Closed: closed}
}

func toSet(ports []int) map[int]bool {
	s := make(map[int]bool, len(ports))
	for _, p := range ports {
		s[p] = true
	}
	return s
}
