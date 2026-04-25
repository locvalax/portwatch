package trend

import (
	"testing"
	"time"
)

func TestSummarize_InsufficientData(t *testing.T) {
	a := New(24 * time.Hour)
	a.Record("host1", 5, time.Now())

	_, ok := a.Summarize("host1")
	if ok {
		t.Fatal("expected ok=false with only one point")
	}
}

func TestSummarize_UnknownHost(t *testing.T) {
	a := New(24 * time.Hour)
	_, ok := a.Summarize("ghost")
	if ok {
		t.Fatal("expected ok=false for unknown host")
	}
}

func TestSummarize_GrowingTrend(t *testing.T) {
	a := New(24 * time.Hour)
	base := time.Now()
	for i := 0; i < 5; i++ {
		a.Record("host1", i*10, base.Add(time.Duration(i)*time.Hour))
	}

	s, ok := a.Summarize("host1")
	if !ok {
		t.Fatal("expected ok=true")
	}
	if s.Direction != Growing {
		t.Errorf("expected Growing, got %s (slope=%.2f)", s.Direction, s.Slope)
	}
	if s.Slope <= 0 {
		t.Errorf("expected positive slope, got %.2f", s.Slope)
	}
}

func TestSummarize_ShrinkingTrend(t *testing.T) {
	a := New(24 * time.Hour)
	base := time.Now()
	for i := 0; i < 5; i++ {
		a.Record("host2", 50-i*10, base.Add(time.Duration(i)*time.Hour))
	}

	s, ok := a.Summarize("host2")
	if !ok {
		t.Fatal("expected ok=true")
	}
	if s.Direction != Shrinking {
		t.Errorf("expected Shrinking, got %s (slope=%.2f)", s.Direction, s.Slope)
	}
}

func TestSummarize_StableTrend(t *testing.T) {
	a := New(24 * time.Hour)
	base := time.Now()
	for i := 0; i < 4; i++ {
		a.Record("host3", 8, base.Add(time.Duration(i)*time.Hour))
	}

	s, ok := a.Summarize("host3")
	if !ok {
		t.Fatal("expected ok=true")
	}
	if s.Direction != Stable {
		t.Errorf("expected Stable, got %s", s.Direction)
	}
	if s.Slope != 0 {
		t.Errorf("expected zero slope, got %.2f", s.Slope)
	}
}

func TestRecord_PrunesOldPoints(t *testing.T) {
	window := 2 * time.Hour
	a := New(window)
	base := time.Now()

	// old point outside window
	a.Record("h", 1, base.Add(-3*time.Hour))
	// recent points
	a.Record("h", 5, base.Add(-1*time.Hour))
	a.Record("h", 9, base)

	s, ok := a.Summarize("h")
	if !ok {
		t.Fatal("expected ok=true")
	}
	if s.Points != 2 {
		t.Errorf("expected 2 points after pruning, got %d", s.Points)
	}
}

func TestSummarize_DifferentHosts_Independent(t *testing.T) {
	a := New(24 * time.Hour)
	base := time.Now()

	for i := 0; i < 3; i++ {
		a.Record("alpha", i*5, base.Add(time.Duration(i)*time.Hour))
		a.Record("beta", 20-i*5, base.Add(time.Duration(i)*time.Hour))
	}

	sa, _ := a.Summarize("alpha")
	sb, _ := a.Summarize("beta")

	if sa.Direction != Growing {
		t.Errorf("alpha: expected Growing, got %s", sa.Direction)
	}
	if sb.Direction != Shrinking {
		t.Errorf("beta: expected Shrinking, got %s", sb.Direction)
	}
}
