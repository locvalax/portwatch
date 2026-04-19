package store

// Diff holds the result of comparing two port snapshots.
type Diff struct {
	Opened []int
	Closed []int
}

// Compare returns a Diff between prev and curr port lists.
// Ports in curr but not prev are Opened; ports in prev but not curr are Closed.
func Compare(prev, curr []int) Diff {
	prevSet := make(map[int]struct{}, len(prev))
	for _, p := range prev {
		prevSet[p] = struct{}{}
	}

	currSet := make(map[int]struct{}, len(curr))
	for _, p := range curr {
		currSet[p] = struct{}{}
	}

	var d Diff
	for _, p := range curr {
		if _, ok := prevSet[p]; !ok {
			d.Opened = append(d.Opened, p)
		}
	}
	for _, p := range prev {
		if _, ok := currSet[p]; !ok {
			d.Closed = append(d.Closed, p)
		}
	}
	return d
}
