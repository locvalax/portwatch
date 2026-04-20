package suppress

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestIsSuppressed_FirstCall_NotSuppressed(t *testing.T) {
	s := New(5 * time.Minute)
	if s.IsSuppressed("192.168.1.1", 80, true) {
		t.Fatal("expected first call to not be suppressed")
	}
}

func TestIsSuppressed_SecondCallWithinWindow_Suppressed(t *testing.T) {
	now := time.Now()
	s := New(5 * time.Minute)
	s.now = fixedNow(now)

	s.IsSuppressed("192.168.1.1", 80, true)
	if !s.IsSuppressed("192.168.1.1", 80, true) {
		t.Fatal("expected second call within window to be suppressed")
	}
}

func TestIsSuppressed_AfterWindowExpires_NotSuppressed(t *testing.T) {
	now := time.Now()
	s := New(5 * time.Minute)
	s.now = fixedNow(now)
	s.IsSuppressed("192.168.1.1", 80, true)

	// advance time past the window
	s.now = fixedNow(now.Add(6 * time.Minute))
	if s.IsSuppressed("192.168.1.1", 80, true) {
		t.Fatal("expected call after window to not be suppressed")
	}
}

func TestIsSuppressed_DifferentDirection_Independent(t *testing.T) {
	s := New(5 * time.Minute)
	s.IsSuppressed("host", 443, true) // opened
	// closed is a different key — should not be suppressed
	if s.IsSuppressed("host", 443, false) {
		t.Fatal("opened and closed should be tracked independently")
	}
}

func TestIsSuppressed_DifferentHosts_Independent(t *testing.T) {
	s := New(5 * time.Minute)
	s.IsSuppressed("host-a", 22, true)
	if s.IsSuppressed("host-b", 22, true) {
		t.Fatal("different hosts should be tracked independently")
	}
}

func TestReset_ClearsHost(t *testing.T) {
	s := New(5 * time.Minute)
	s.IsSuppressed("host-a", 22, true)
	s.IsSuppressed("host-a", 80, false)
	s.IsSuppressed("host-b", 22, true)

	s.Reset("host-a")

	if s.IsSuppressed("host-a", 22, true) {
		t.Fatal("expected host-a/22 to be cleared after Reset")
	}
	// host-b should still be suppressed
	if !s.IsSuppressed("host-b", 22, true) {
		t.Fatal("expected host-b/22 to remain suppressed")
	}
}

func TestFlush_RemovesExpiredEntries(t *testing.T) {
	now := time.Now()
	s := New(1 * time.Minute)
	s.now = fixedNow(now)
	s.IsSuppressed("host", 8080, true)

	// advance past window and flush
	s.now = fixedNow(now.Add(2 * time.Minute))
	s.Flush()

	if len(s.entries) != 0 {
		t.Fatalf("expected 0 entries after flush, got %d", len(s.entries))
	}
}

func TestString_ReturnsDescription(t *testing.T) {
	s := New(10 * time.Minute)
	s.IsSuppressed("host", 80, true)
	out := s.String()
	if out == "" {
		t.Fatal("expected non-empty string description")
	}
}
