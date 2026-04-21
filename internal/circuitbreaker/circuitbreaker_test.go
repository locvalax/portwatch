package circuitbreaker_test

import (
	"testing"
	"time"

	"github.com/yourusername/portwatch/internal/circuitbreaker"
)

func TestAllow_InitiallyPermits(t *testing.T) {
	b := circuitbreaker.New(circuitbreaker.DefaultOptions())
	if err := b.Allow("host1"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestAllow_OpensAfterMaxFailures(t *testing.T) {
	opts := circuitbreaker.Options{MaxFailures: 2, OpenDuration: 10 * time.Second}
	b := circuitbreaker.New(opts)

	b.RecordFailure("host1")
	b.RecordFailure("host1")

	if err := b.Allow("host1"); err != circuitbreaker.ErrCircuitOpen {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestAllow_HalfOpenAfterDuration(t *testing.T) {
	opts := circuitbreaker.Options{MaxFailures: 1, OpenDuration: 10 * time.Millisecond}
	b := circuitbreaker.New(opts)

	b.RecordFailure("host1")
	time.Sleep(20 * time.Millisecond)

	if err := b.Allow("host1"); err != nil {
		t.Fatalf("expected circuit to be half-open, got %v", err)
	}
	if b.State("host1") != circuitbreaker.StateHalfOpen {
		t.Fatalf("expected StateHalfOpen")
	}
}

func TestRecordSuccess_ResetsClosed(t *testing.T) {
	opts := circuitbreaker.Options{MaxFailures: 1, OpenDuration: time.Hour}
	b := circuitbreaker.New(opts)

	b.RecordFailure("host1")
	b.RecordSuccess("host1")

	if b.State("host1") != circuitbreaker.StateClosed {
		t.Fatalf("expected StateClosed after success")
	}
	if err := b.Allow("host1"); err != nil {
		t.Fatalf("expected nil after reset, got %v", err)
	}
}

func TestAllow_DifferentHosts_Independent(t *testing.T) {
	opts := circuitbreaker.Options{MaxFailures: 1, OpenDuration: time.Hour}
	b := circuitbreaker.New(opts)

	b.RecordFailure("host1")

	if err := b.Allow("host2"); err != nil {
		t.Fatalf("host2 should not be affected by host1 failures: %v", err)
	}
}

func TestDefaultOptions_Values(t *testing.T) {
	opts := circuitbreaker.DefaultOptions()
	if opts.MaxFailures <= 0 {
		t.Fatal("MaxFailures should be positive")
	}
	if opts.OpenDuration <= 0 {
		t.Fatal("OpenDuration should be positive")
	}
}
