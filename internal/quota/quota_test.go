package quota

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestAllow_FirstCall_Permitted(t *testing.T) {
	l := New(DefaultOptions())
	if err := l.Allow("host-a"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestAllow_ExceedsMax_ReturnsError(t *testing.T) {
	opts := Options{MaxScans: 3, Window: time.Hour}
	l := New(opts)
	for i := 0; i < 3; i++ {
		if err := l.Allow("host-a"); err != nil {
			t.Fatalf("unexpected error on call %d: %v", i+1, err)
		}
	}
	if err := l.Allow("host-a"); err != ErrQuotaExceeded {
		t.Fatalf("expected ErrQuotaExceeded, got %v", err)
	}
}

func TestAllow_AfterWindowExpires_Resets(t *testing.T) {
	now := time.Now()
	opts := Options{MaxScans: 1, Window: time.Millisecond}
	l := New(opts)
	l.now = fixedClock(now)

	if err := l.Allow("host-a"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := l.Allow("host-a"); err != ErrQuotaExceeded {
		t.Fatalf("expected exceeded, got %v", err)
	}

	// Advance past window
	l.now = fixedClock(now.Add(2 * time.Millisecond))
	if err := l.Allow("host-a"); err != nil {
		t.Fatalf("expected reset, got %v", err)
	}
}

func TestAllow_DifferentHosts_Independent(t *testing.T) {
	opts := Options{MaxScans: 1, Window: time.Hour}
	l := New(opts)
	l.Allow("host-a")
	if err := l.Allow("host-b"); err != nil {
		t.Fatalf("host-b should not be limited: %v", err)
	}
}

func TestRemaining_DecreasesWithUse(t *testing.T) {
	opts := Options{MaxScans: 5, Window: time.Hour}
	l := New(opts)
	if r := l.Remaining("host-a"); r != 5 {
		t.Fatalf("expected 5, got %d", r)
	}
	l.Allow("host-a")
	l.Allow("host-a")
	if r := l.Remaining("host-a"); r != 3 {
		t.Fatalf("expected 3, got %d", r)
	}
}

func TestReset_ClearsHost(t *testing.T) {
	opts := Options{MaxScans: 1, Window: time.Hour}
	l := New(opts)
	l.Allow("host-a")
	l.Reset("host-a")
	if err := l.Allow("host-a"); err != nil {
		t.Fatalf("expected nil after reset, got %v", err)
	}
}

func TestDefaultOptions_Values(t *testing.T) {
	o := DefaultOptions()
	if o.MaxScans != 10 {
		t.Fatalf("expected MaxScans=10, got %d", o.MaxScans)
	}
	if o.Window != time.Hour {
		t.Fatalf("expected Window=1h, got %v", o.Window)
	}
}
