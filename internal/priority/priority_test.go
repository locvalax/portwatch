package priority

import (
	"testing"
)

func TestDefaultOptions_Fields(t *testing.T) {
	opts := DefaultOptions()
	if len(opts.HighPorts) == 0 {
		t.Fatal("expected non-empty HighPorts")
	}
	if len(opts.MediumPorts) == 0 {
		t.Fatal("expected non-empty MediumPorts")
	}
}

func TestRank_UnknownPort_ReturnsLow(t *testing.T) {
	r := New(DefaultOptions())
	if got := r.Rank([]int{9999}); got != Low {
		t.Fatalf("expected Low, got %s", got)
	}
}

func TestRank_HighPort_ReturnsHigh(t *testing.T) {
	r := New(DefaultOptions())
	if got := r.Rank([]int{22}); got != High {
		t.Fatalf("expected High, got %s", got)
	}
}

func TestRank_MediumPort_ReturnsMedium(t *testing.T) {
	r := New(DefaultOptions())
	if got := r.Rank([]int{80}); got != Medium {
		t.Fatalf("expected Medium, got %s", got)
	}
}

func TestRank_CriticalBeatsHigh(t *testing.T) {
	opts := DefaultOptions()
	opts.CriticalPorts = []int{4444}
	r := New(opts)
	if got := r.Rank([]int{22, 4444}); got != Critical {
		t.Fatalf("expected Critical, got %s", got)
	}
}

func TestRank_EmptyPorts_ReturnsLow(t *testing.T) {
	r := New(DefaultOptions())
	if got := r.Rank(nil); got != Low {
		t.Fatalf("expected Low, got %s", got)
	}
}

func TestSort_HighPortsFirst(t *testing.T) {
	r := New(DefaultOptions())
	sorted := r.Sort([]int{9000, 80, 22})
	if sorted[0] != 22 {
		t.Fatalf("expected 22 first (High), got %d", sorted[0])
	}
	if sorted[1] != 80 {
		t.Fatalf("expected 80 second (Medium), got %d", sorted[1])
	}
	if sorted[2] != 9000 {
		t.Fatalf("expected 9000 last (Low), got %d", sorted[2])
	}
}

func TestSort_StableWithinSameLevel(t *testing.T) {
	r := New(DefaultOptions())
	// 80 and 8080 are both Medium; lower port should come first
	sorted := r.Sort([]int{8080, 80})
	if sorted[0] != 80 {
		t.Fatalf("expected 80 before 8080, got %d", sorted[0])
	}
}

func TestLevelString(t *testing.T) {
	cases := []struct {
		lvl  Level
		want string
	}{
		{Low, "low"},
		{Medium, "medium"},
		{High, "high"},
		{Critical, "critical"},
	}
	for _, c := range cases {
		if got := c.lvl.String(); got != c.want {
			t.Errorf("Level(%d).String() = %q, want %q", c.lvl, got, c.want)
		}
	}
}
