package backoff

import (
	"testing"
	"time"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if opts.InitialInterval <= 0 {
		t.Fatal("expected positive InitialInterval")
	}
	if opts.MaxInterval < opts.InitialInterval {
		t.Fatal("MaxInterval must be >= InitialInterval")
	}
	if opts.Multiplier <= 1 {
		t.Fatal("expected Multiplier > 1")
	}
	if opts.MaxAttempts <= 0 {
		t.Fatal("expected positive MaxAttempts")
	}
}

func TestDuration_ZeroAttempt_ReturnsInitial(t *testing.T) {
	b := New(DefaultOptions())
	d, ok := b.Duration(0)
	if !ok {
		t.Fatal("expected ok=true for attempt 0")
	}
	if d != DefaultOptions().InitialInterval {
		t.Fatalf("expected %v, got %v", DefaultOptions().InitialInterval, d)
	}
}

func TestDuration_GrowsExponentially(t *testing.T) {
	opts := Options{
		InitialInterval: 100 * time.Millisecond,
		MaxInterval:     10 * time.Second,
		Multiplier:      2.0,
		MaxAttempts:     10,
	}
	b := New(opts)
	d0, _ := b.Duration(0)
	d1, _ := b.Duration(1)
	d2, _ := b.Duration(2)
	if d1 != 2*d0 {
		t.Fatalf("expected %v, got %v", 2*d0, d1)
	}
	if d2 != 2*d1 {
		t.Fatalf("expected %v, got %v", 2*d1, d2)
	}
}

func TestDuration_CappedAtMaxInterval(t *testing.T) {
	opts := Options{
		InitialInterval: 1 * time.Second,
		MaxInterval:     2 * time.Second,
		Multiplier:      10.0,
		MaxAttempts:     5,
	}
	b := New(opts)
	for i := 0; i < 5; i++ {
		d, _ := b.Duration(i)
		if d > opts.MaxInterval {
			t.Fatalf("attempt %d: duration %v exceeds MaxInterval %v", i, d, opts.MaxInterval)
		}
	}
}

func TestDuration_ExceedsMaxAttempts_ReturnsFalse(t *testing.T) {
	opts := DefaultOptions()
	b := New(opts)
	_, ok := b.Duration(opts.MaxAttempts)
	if ok {
		t.Fatal("expected ok=false when attempt >= MaxAttempts")
	}
}

func TestSequence_LengthMatchesMaxAttempts(t *testing.T) {
	opts := DefaultOptions()
	b := New(opts)
	seq := b.Sequence()
	if len(seq) != opts.MaxAttempts {
		t.Fatalf("expected %d durations, got %d", opts.MaxAttempts, len(seq))
	}
}

func TestFromConfig_OverridesDefaults(t *testing.T) {
	cfg := Config{
		InitialIntervalMs: 50,
		MaxIntervalMs:     5000,
		Multiplier:        3.0,
		MaxAttempts:       7,
	}
	opts := FromConfig(cfg)
	if opts.InitialInterval != 50*time.Millisecond {
		t.Fatalf("unexpected InitialInterval: %v", opts.InitialInterval)
	}
	if opts.MaxInterval != 5*time.Second {
		t.Fatalf("unexpected MaxInterval: %v", opts.MaxInterval)
	}
	if opts.Multiplier != 3.0 {
		t.Fatalf("unexpected Multiplier: %v", opts.Multiplier)
	}
	if opts.MaxAttempts != 7 {
		t.Fatalf("unexpected MaxAttempts: %v", opts.MaxAttempts)
	}
}

func TestFromConfig_ZeroValues_UsesDefaults(t *testing.T) {
	opts := FromConfig(Config{})
	defaults := DefaultOptions()
	if opts.InitialInterval != defaults.InitialInterval {
		t.Fatalf("expected default InitialInterval, got %v", opts.InitialInterval)
	}
	if opts.MaxInterval != defaults.MaxInterval {
		t.Fatalf("expected default MaxInterval, got %v", opts.MaxInterval)
	}
}
