package decay

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestObserve_FirstCall_SetsScore(t *testing.T) {
	s := New(DefaultOptions())
	s.Observe("host:80", 1.0)
	if got := s.Score("host:80"); got != 1.0 {
		t.Fatalf("expected 1.0, got %f", got)
	}
}

func TestScore_Unknown_ReturnsZero(t *testing.T) {
	s := New(DefaultOptions())
	if got := s.Score("host:9999"); got != 0 {
		t.Fatalf("expected 0, got %f", got)
	}
}

func TestScore_DecaysOverTime(t *testing.T) {
	base := time.Now()
	s := New(Options{HalfLife: 10 * time.Second, Floor: 0.001})
	s.now = fixedClock(base)
	s.Observe("h:80", 1.0)

	// Advance by one half-life; score should be ~0.5.
	s.now = fixedClock(base.Add(10 * time.Second))
	got := s.Score("h:80")
	if got < 0.49 || got > 0.51 {
		t.Fatalf("expected ~0.5 after one half-life, got %f", got)
	}
}

func TestScore_BelowFloor_ReturnsZero(t *testing.T) {
	base := time.Now()
	s := New(Options{HalfLife: 1 * time.Second, Floor: 0.1})
	s.now = fixedClock(base)
	s.Observe("h:443", 1.0)

	// Advance far enough that the score drops below the floor.
	s.now = fixedClock(base.Add(10 * time.Second))
	if got := s.Score("h:443"); got != 0 {
		t.Fatalf("expected 0 below floor, got %f", got)
	}
}

func TestObserve_Accumulates(t *testing.T) {
	s := New(DefaultOptions())
	s.Observe("h:22", 1.0)
	s.Observe("h:22", 1.0)
	if got := s.Score("h:22"); got < 1.9 {
		t.Fatalf("expected accumulated score >= 1.9, got %f", got)
	}
}

func TestReset_ClearsEntry(t *testing.T) {
	s := New(DefaultOptions())
	s.Observe("h:80", 1.0)
	s.Reset("h:80")
	if got := s.Score("h:80"); got != 0 {
		t.Fatalf("expected 0 after reset, got %f", got)
	}
}

func TestDifferentKeys_Independent(t *testing.T) {
	s := New(DefaultOptions())
	s.Observe("host-a:80", 2.0)
	s.Observe("host-b:80", 0.5)

	if s.Score("host-a:80") <= s.Score("host-b:80") {
		t.Fatal("host-a score should be greater than host-b score")
	}
}
