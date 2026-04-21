package cooldown_test

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/cooldown"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestAllow_FirstCall_Permitted(t *testing.T) {
	cd := cooldown.New(cooldown.Options{
		Period: time.Minute,
		Now:    fixedClock(time.Now()),
	})
	if !cd.Allow("host1") {
		t.Fatal("expected first call to be allowed")
	}
}

func TestAllow_SecondCallWithinWindow_Blocked(t *testing.T) {
	now := time.Now()
	cd := cooldown.New(cooldown.Options{
		Period: time.Minute,
		Now:    fixedClock(now),
	})
	cd.Allow("host1") // record timestamp
	if cd.Allow("host1") {
		t.Fatal("expected second call within window to be blocked")
	}
}

func TestAllow_AfterWindowExpires_Permitted(t *testing.T) {
	now := time.Now()
	current := now
	cd := cooldown.New(cooldown.Options{
		Period: time.Minute,
		Now:    func() time.Time { return current },
	})
	cd.Allow("host1")
	current = now.Add(2 * time.Minute)
	if !cd.Allow("host1") {
		t.Fatal("expected call after window to be permitted")
	}
}

func TestAllow_DifferentHosts_Independent(t *testing.T) {
	now := time.Now()
	cd := cooldown.New(cooldown.Options{
		Period: time.Minute,
		Now:    fixedClock(now),
	})
	cd.Allow("host1")
	if !cd.Allow("host2") {
		t.Fatal("expected different host to be independent")
	}
}

func TestReset_ClearsHost(t *testing.T) {
	now := time.Now()
	cd := cooldown.New(cooldown.Options{
		Period: time.Minute,
		Now:    fixedClock(now),
	})
	cd.Allow("host1")
	cd.Reset("host1")
	if !cd.Allow("host1") {
		t.Fatal("expected Allow to pass after Reset")
	}
}

func TestFlush_ClearsAll(t *testing.T) {
	now := time.Now()
	cd := cooldown.New(cooldown.Options{
		Period: time.Minute,
		Now:    fixedClock(now),
	})
	cd.Allow("host1")
	cd.Allow("host2")
	cd.Flush()
	for _, h := range []string{"host1", "host2"} {
		if !cd.Allow(h) {
			t.Fatalf("expected %q to be allowed after Flush", h)
		}
	}
}

func TestDefaultOptions_Period(t *testing.T) {
	opts := cooldown.DefaultOptions()
	if opts.Period != 5*time.Minute {
		t.Fatalf("expected default period 5m, got %v", opts.Period)
	}
}
