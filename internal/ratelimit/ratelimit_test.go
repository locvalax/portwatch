package ratelimit_test

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/ratelimit"
)

func TestAllow_FirstCallPermitted(t *testing.T) {
	l := ratelimit.New(100 * time.Millisecond)
	if !l.Allow("host1") {
		t.Fatal("expected first call to be allowed")
	}
}

func TestAllow_SecondCallBlocked(t *testing.T) {
	l := ratelimit.New(100 * time.Millisecond)
	l.Allow("host1")
	if l.Allow("host1") {
		t.Fatal("expected second immediate call to be blocked")
	}
}

func TestAllow_AfterInterval_Permitted(t *testing.T) {
	l := ratelimit.New(20 * time.Millisecond)
	l.Allow("host1")
	time.Sleep(30 * time.Millisecond)
	if !l.Allow("host1") {
		t.Fatal("expected call after interval to be allowed")
	}
}

func TestAllow_DifferentHosts_Independent(t *testing.T) {
	l := ratelimit.New(100 * time.Millisecond)
	l.Allow("host1")
	if !l.Allow("host2") {
		t.Fatal("expected different host to be allowed independently")
	}
}

func TestReset_ClearsHost(t *testing.T) {
	l := ratelimit.New(100 * time.Millisecond)
	l.Allow("host1")
	l.Reset("host1")
	if !l.Allow("host1") {
		t.Fatal("expected host to be allowed after reset")
	}
}

func TestResetAll_ClearsAllHosts(t *testing.T) {
	l := ratelimit.New(100 * time.Millisecond)
	l.Allow("host1")
	l.Allow("host2")
	l.ResetAll()
	if !l.Allow("host1") || !l.Allow("host2") {
		t.Fatal("expected all hosts to be allowed after ResetAll")
	}
}
