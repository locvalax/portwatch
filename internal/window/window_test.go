package window_test

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/window"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestRecord_FirstCall_Permitted(t *testing.T) {
	w := window.New(window.Options{
		Size:     time.Minute,
		MaxCount: 3,
		Clock:    fixedClock(time.Now()),
	})
	count, ok := w.Record("host1")
	if !ok {
		t.Fatalf("expected first call to be permitted")
	}
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}
}

func TestRecord_ExceedsMax_Blocked(t *testing.T) {
	now := time.Now()
	w := window.New(window.Options{
		Size:     time.Minute,
		MaxCount: 2,
		Clock:    fixedClock(now),
	})
	w.Record("host1")
	w.Record("host1")
	count, ok := w.Record("host1")
	if ok {
		t.Fatalf("expected third call to be blocked")
	}
	if count != 3 {
		t.Fatalf("expected count 3, got %d", count)
	}
}

func TestRecord_AfterWindowExpires_Resets(t *testing.T) {
	base := time.Now()
	clock := base
	w := window.New(window.Options{
		Size:     time.Minute,
		MaxCount: 1,
		Clock:    func() time.Time { return clock },
	})
	w.Record("host1")
	// advance past window
	clock = base.Add(2 * time.Minute)
	count, ok := w.Record("host1")
	if !ok {
		t.Fatalf("expected call after window expiry to be permitted")
	}
	if count != 1 {
		t.Fatalf("expected count 1 after reset, got %d", count)
	}
}

func TestRecord_DifferentHosts_Independent(t *testing.T) {
	now := time.Now()
	w := window.New(window.Options{
		Size:     time.Minute,
		MaxCount: 1,
		Clock:    fixedClock(now),
	})
	w.Record("hostA")
	_, ok := w.Record("hostB")
	if !ok {
		t.Fatalf("expected hostB to be independent of hostA")
	}
}

func TestCount_ReflectsWindow(t *testing.T) {
	now := time.Now()
	w := window.New(window.Options{
		Size:     time.Minute,
		MaxCount: 10,
		Clock:    fixedClock(now),
	})
	w.Record("host1")
	w.Record("host1")
	if c := w.Count("host1"); c != 2 {
		t.Fatalf("expected count 2, got %d", c)
	}
}

func TestReset_ClearsHost(t *testing.T) {
	now := time.Now()
	w := window.New(window.Options{
		Size:     time.Minute,
		MaxCount: 10,
		Clock:    fixedClock(now),
	})
	w.Record("host1")
	w.Record("host1")
	w.Reset("host1")
	if c := w.Count("host1"); c != 0 {
		t.Fatalf("expected count 0 after reset, got %d", c)
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := window.DefaultOptions()
	if opts.Size != 5*time.Minute {
		t.Fatalf("unexpected default size: %v", opts.Size)
	}
	if opts.MaxCount != 100 {
		t.Fatalf("unexpected default max count: %d", opts.MaxCount)
	}
}
