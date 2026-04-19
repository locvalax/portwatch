package baseline

import (
	"sort"
	"testing"
)

func sorted(s []int) []int {
	sort.Ints(s)
	return s
}

func TestCompare_Opened(t *testing.T) {
	snap := Snapshot{Ports: []int{80}}
	diff := Compare(snap, []int{80, 443})
	if len(diff.Closed) != 0 {
		t.Errorf("unexpected closed: %v", diff.Closed)
	}
	if got := sorted(diff.Opened); len(got) != 1 || got[0] != 443 {
		t.Errorf("opened: got %v want [443]", got)
	}
}

func TestCompare_Closed(t *testing.T) {
	snap := Snapshot{Ports: []int{80, 8080}}
	diff := Compare(snap, []int{80})
	if len(diff.Opened) != 0 {
		t.Errorf("unexpected opened: %v", diff.Opened)
	}
	if got := sorted(diff.Closed); len(got) != 1 || got[0] != 8080 {
		t.Errorf("closed: got %v want [8080]", got)
	}
}

func TestCompare_NoChanges(t *testing.T) {
	snap := Snapshot{Ports: []int{22, 80}}
	diff := Compare(snap, []int{22, 80})
	if diff.HasChanges() {
		t.Errorf("expected no changes, got %+v", diff)
	}
}

func TestCompare_EmptyBaseline(t *testing.T) {
	snap := Snapshot{Ports: nil}
	diff := Compare(snap, []int{9000})
	if got := sorted(diff.Opened); len(got) != 1 || got[0] != 9000 {
		t.Errorf("opened: got %v", got)
	}
}
