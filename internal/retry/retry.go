package retry

import (
	"context"
	"errors"
	"time"
)

// Policy defines retry behaviour for a scan attempt.
type Policy struct {
	MaxAttempts int
	Delay       time.Duration
	Multiplier  float64 // backoff multiplier; 1.0 = constant delay
}

// DefaultPolicy returns a sensible out-of-the-box policy.
func DefaultPolicy() Policy {
	return Policy{
		MaxAttempts: 3,
		Delay:       500 * time.Millisecond,
		Multiplier:  2.0,
	}
}

// ErrExhausted is returned when all attempts are consumed.
var ErrExhausted = errors.New("retry: all attempts exhausted")

// Do runs fn up to p.MaxAttempts times, backing off between attempts.
// It returns the first nil error or ErrExhausted wrapping the last error.
func (p Policy) Do(ctx context.Context, fn func() error) error {
	if p.MaxAttempts < 1 {
		p.MaxAttempts = 1
	}
	if p.Multiplier <= 0 {
		p.Multiplier = 1.0
	}

	delay := p.Delay
	var last error

	for attempt := 0; attempt < p.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if last = fn(); last == nil {
			return nil
		}
		if attempt < p.MaxAttempts-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
			delay = time.Duration(float64(delay) * p.Multiplier)
		}
	}
	return errors.Join(ErrExhausted, last)
}
