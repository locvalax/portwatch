package throttle

import (
	"testing"
	"time"
)

func TestAllow_BurstPermitsMultipleCalls(t *testing.T) {
	th := New(Options{Rate: time.Minute, Burst: 3})
	for i := 0; i < 3; i++ {
		if !th.Allow("host1") {
			t.Fatalf("expected call %d to be allowed", i+1)
		}
	}
}

func TestAllow_ExceedsBurst_Blocked(t *testing.T) {
	th := New(Options{Rate: time.Minute, Burst: 2})
	th.Allow("host1")
	th.Allow("host1")
	if th.Allow("host1") {
		t.Fatal("expected third call to be blocked after burst exhausted")
	}
}

func TestAllow_TokenReplenishesOverTime(t *testing.T) {
	th := New(Options{Rate: 50 * time.Millisecond, Burst: 1})
	if !th.Allow("host1") {
		t.Fatal("first call should be allowed")
	}
	if th.Allow("host1") {
		t.Fatal("immediate second call should be blocked")
	}
	time.Sleep(60 * time.Millisecond)
	if !th.Allow("host1") {
		t.Fatal("call after rate interval should be allowed")
	}
}

func TestAllow_DifferentHosts_Independent(t *testing.T) {
	th := New(Options{Rate: time.Minute, Burst: 1})
	if !th.Allow("host1") {
		t.Fatal("host1 first call should be allowed")
	}
	if !th.Allow("host2") {
		t.Fatal("host2 first call should be allowed independently")
	}
	if th.Allow("host1") {
		t.Fatal("host1 second call should be blocked")
	}
}

func TestReset_RestoresTokens(t *testing.T) {
	th := New(Options{Rate: time.Minute, Burst: 1})
	th.Allow("host1") // consume token
	if th.Allow("host1") {
		t.Fatal("should be blocked before reset")
	}
	th.Reset("host1")
	if !th.Allow("host1") {
		t.Fatal("should be allowed after reset")
	}
}

func TestDefaultOptions_SaneValues(t *testing.T) {
	opts := DefaultOptions()
	if opts.Rate <= 0 {
		t.Error("default rate should be positive")
	}
	if opts.Burst <= 0 {
		t.Error("default burst should be positive")
	}
}
