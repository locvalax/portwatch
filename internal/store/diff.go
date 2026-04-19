package store

// Diff describes the changes between two port snapshots.
type Diff struct {
	Host    string
	Opened  []int
	Closed  []int
}

// HasChanges returns true if any ports were opened or closed.
func (d Diff) HasChanges() bool {
	return len(d.Opened) > 0 || len(d.Closed) > 0
}

// Compare computes the difference between a previous and current set of ports.
// prev may be nil (treated as empty).
func Compare(host string, prev *Snapshot, current []int) Diff {
	d := Diff{Host: host}

	prevSet := make(map[int]struct{})
	if prev != nil {
		for _, p := range prev.Ports {
			prevSet[p] = struct{}{}
		}
	}

	currSet := make(map[int]struct{})
	for _, p := range current {
		currSet[p] = struct{}{}
		if _, ok := prevSet[p]; !ok {
			d.Opened = append(d.Opened, p)
		}
	}

	if prev != nil {
		for _, p := range prev.Ports {
			if _, ok := currSet[p]; !ok {
				d.Closed = append(d.Closed, p)
			}
		}
	}

	return d
}
