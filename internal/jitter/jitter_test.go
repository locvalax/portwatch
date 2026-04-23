package jitter_test

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/user/portwatch/internal/jitter"
)

func TestDefaultOptions_Factor(t *testing.T) {
	opts := jitter.DefaultOptions()
	if opts.Factor != 0.25 {
		t.Fatalf("expected factor 0.25, got %v", opts.Factor)
	}
}

func TestApply_InRange(t *testing.T) {
	// Use a fixed seed for determinism.
	rng := rand.New(rand.NewSource(42))
	j := jitter.New(jitter.Options{Factor: 0.5, Rand: rng})

	base := 100 * time.Millisecond
	for i := 0; i < 50; i++ {
		got := j.Apply(base)
		if got < base {
			t.Fatalf("jittered value %v is less than base %v", got, base)
		}
		max := base + time.Duration(float64(base)*0.5)
		if got > max {
			t.Fatalf("jittered value %v exceeds max %v", got, max)
		}
	}
}

func TestApply_InvalidFactor_Clamped(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	j := jitter.New(jitter.Options{Factor: -1, Rand: rng})

	base := 200 * time.Millisecond
	got := j.Apply(base)
	if got < base {
		t.Fatalf("expected at least base duration, got %v", got)
	}
}

func TestSleep_CompletesWithinBound(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	j := jitter.New(jitter.Options{Factor: 0.1, Rand: rng})

	base := 20 * time.Millisecond
	start := time.Now()
	if err := j.Sleep(context.Background(), base); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	elapsed := time.Since(start)
	if elapsed < base {
		t.Fatalf("sleep returned too early: %v", elapsed)
	}
	max := base + time.Duration(float64(base)*0.1) + 5*time.Millisecond
	if elapsed > max {
		t.Fatalf("sleep took too long: %v (max %v)", elapsed, max)
	}
}

func TestSleep_ContextCancel_ReturnsError(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	j := jitter.New(jitter.Options{Factor: 0.25, Rand: rng})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := j.Sleep(ctx, 10*time.Second)
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
}
