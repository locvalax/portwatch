package resolve

import (
	"testing"
	"time"
)

func TestCache_ReturnsCachedResult(t *testing.T) {
	r := New(2 * time.Second)
	c := NewCache(r, 10*time.Second)

	first, err := c.Resolve("localhost")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	second, err := c.Resolve("localhost")
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}

	if len(first) == 0 || len(second) == 0 {
		t.Fatal("expected at least one address")
	}
	if first[0] != second[0] {
		t.Errorf("expected same cached address, got %q and %q", first[0], second[0])
	}
}

func TestCache_Invalidate_ForcesRefresh(t *testing.T) {
	r := New(2 * time.Second)
	c := NewCache(r, 10*time.Second)

	if _, err := c.Resolve("localhost"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c.Invalidate("localhost")

	c.mu.Lock()
	_, cached := c.entries["localhost"]
	c.mu.Unlock()

	if cached {
		t.Error("expected entry to be removed after Invalidate")
	}
}

func TestCache_Flush_ClearsAll(t *testing.T) {
	r := New(2 * time.Second)
	c := NewCache(r, 10*time.Second)

	for _, h := range []string{"localhost", "127.0.0.1"} {
		if _, err := c.Resolve(h); err != nil {
			t.Fatalf("resolve %q: %v", h, err)
		}
	}

	c.Flush()

	c.mu.Lock()
	n := len(c.entries)
	c.mu.Unlock()

	if n != 0 {
		t.Errorf("expected 0 entries after Flush, got %d", n)
	}
}

func TestCache_ExpiredEntry_Refreshes(t *testing.T) {
	r := New(2 * time.Second)
	c := NewCache(r, 1*time.Millisecond)

	if _, err := c.Resolve("localhost"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	time.Sleep(5 * time.Millisecond)

	// Entry should be stale; Resolve must succeed by re-resolving.
	addrs, err := c.Resolve("localhost")
	if err != nil {
		t.Fatalf("unexpected error after expiry: %v", err)
	}
	if len(addrs) == 0 {
		t.Error("expected at least one address after cache refresh")
	}
}

func TestCache_Invalidate_UnknownHost_IsNoop(t *testing.T) {
	// Invalidating a host that was never resolved should not panic or error.
	r := New(2 * time.Second)
	c := NewCache(r, 10*time.Second)

	c.Invalidate("nonexistent.example.invalid")

	c.mu.Lock()
	n := len(c.entries)
	c.mu.Unlock()

	if n != 0 {
		t.Errorf("expected 0 entries, got %d", n)
	}
}
