package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

var errTemp = errors.New("temporary failure")

func TestDo_SucceedsOnFirstAttempt(t *testing.T) {
	p := Policy{MaxAttempts: 3, Delay: 0, Multiplier: 1}
	calls := 0
	err := p.Do(context.Background(), func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestDo_RetriesOnFailure(t *testing.T) {
	p := Policy{MaxAttempts: 3, Delay: 0, Multiplier: 1}
	calls := 0
	err := p.Do(context.Background(), func() error {
		calls++
		if calls < 3 {
			return errTemp
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil after retries, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_ExhaustsAttempts(t *testing.T) {
	p := Policy{MaxAttempts: 3, Delay: 0, Multiplier: 1}
	calls := 0
	err := p.Do(context.Background(), func() error {
		calls++
		return errTemp
	})
	if !errors.Is(err, ErrExhausted) {
		t.Fatalf("expected ErrExhausted, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_ContextCancelled(t *testing.T) {
	p := Policy{MaxAttempts: 5, Delay: 100 * time.Millisecond, Multiplier: 1}
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	err := p.Do(ctx, func() error {
		calls++
		if calls == 1 {
			cancel()
		}
		return errTemp
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestDefaultPolicy_Values(t *testing.T) {
	p := DefaultPolicy()
	if p.MaxAttempts != 3 {
		t.Errorf("MaxAttempts: want 3, got %d", p.MaxAttempts)
	}
	if p.Multiplier != 2.0 {
		t.Errorf("Multiplier: want 2.0, got %f", p.Multiplier)
	}
}
