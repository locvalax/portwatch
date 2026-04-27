package cache

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestGet_MissingEntry_ReturnsFalse(t *testing.T) {
	c := New(DefaultOptions())
	_, ok := c.Get("host1")
	if ok {
		t.Fatal("expected false for unknown host")
	}
}

func TestSetAndGet_ReturnsStoredPorts(t *testing.T) {
	c := New(DefaultOptions())
	c.Set("host1", []uint16{80, 443})
	ports, ok := c.Get("host1")
	if !ok {
		t.Fatal("expected cached entry to be found")
	}
	if len(ports) != 2 {
		t.Fatalf("expected 2 ports, got %d", len(ports))
	}
}

func TestGet_ExpiredEntry_ReturnsFalse(t *testing.T) {
	now := time.Now()
	c := New(Options{TTL: time.Second, MaxSize: 10})
	c.nowFunc = fixedClock(now)
	c.Set("host1", []uint16{22})
	c.nowFunc = fixedClock(now.Add(2 * time.Second))
	_, ok := c.Get("host1")
	if ok {
		t.Fatal("expected expired entry to be absent")
	}
}

func TestInvalidate_RemovesEntry(t *testing.T) {
	c := New(DefaultOptions())
	c.Set("host1", []uint16{8080})
	c.Invalidate("host1")
	_, ok := c.Get("host1")
	if ok {
		t.Fatal("expected entry to be removed after invalidation")
	}
}

func TestFlush_ClearsAll(t *testing.T) {
	c := New(DefaultOptions())
	c.Set("host1", []uint16{80})
	c.Set("host2", []uint16{443})
	c.Flush()
	for _, h := range []string{"host1", "host2"} {
		if _, ok := c.Get(h); ok {
			t.Fatalf("expected %s to be absent after flush", h)
		}
	}
}

func TestSet_EvictsOldestWhenFull(t *testing.T) {
	now := time.Now()
	c := New(Options{TTL: time.Minute, MaxSize: 2})
	c.nowFunc = fixedClock(now)
	c.Set("host1", []uint16{80})
	c.nowFunc = fixedClock(now.Add(time.Second))
	c.Set("host2", []uint16{443})
	c.nowFunc = fixedClock(now.Add(2 * time.Second))
	c.Set("host3", []uint16{8080})
	// host1 was oldest; it should have been evicted
	if _, ok := c.Get("host1"); ok {
		t.Fatal("expected host1 to be evicted")
	}
	if _, ok := c.Get("host3"); !ok {
		t.Fatal("expected host3 to be present")
	}
}

func TestDefaultOptions_Fields(t *testing.T) {
	opts := DefaultOptions()
	if opts.TTL != 5*time.Minute {
		t.Fatalf("unexpected TTL: %v", opts.TTL)
	}
	if opts.MaxSize != 256 {
		t.Fatalf("unexpected MaxSize: %d", opts.MaxSize)
	}
}
