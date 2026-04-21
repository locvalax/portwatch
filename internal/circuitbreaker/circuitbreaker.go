package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// State represents the circuit breaker state.
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

// Options configures a Breaker.
type Options struct {
	MaxFailures int
	OpenDuration time.Duration
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		MaxFailures:  3,
		OpenDuration: 30 * time.Second,
	}
}

// Breaker is a per-host circuit breaker.
type Breaker struct {
	mu       sync.Mutex
	opts     Options
	hosts    map[string]*hostState
}

type hostState struct {
	failures  int
	state     State
	openedAt  time.Time
}

// New creates a Breaker with the given options.
func New(opts Options) *Breaker {
	return &Breaker{
		opts:  opts,
		hosts: make(map[string]*hostState),
	}
}

// Allow returns nil if the host is allowed through, or ErrCircuitOpen.
func (b *Breaker) Allow(host string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	hs := b.ensureHost(host)
	switch hs.state {
	case StateClosed:
		return nil
	case StateOpen:
		if time.Since(hs.openedAt) >= b.opts.OpenDuration {
			hs.state = StateHalfOpen
			return nil
		}
		return ErrCircuitOpen
	case StateHalfOpen:
		return nil
	}
	return nil
}

// RecordSuccess resets failure count for a host.
func (b *Breaker) RecordSuccess(host string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	hs := b.ensureHost(host)
	hs.failures = 0
	hs.state = StateClosed
}

// RecordFailure increments failure count and may open the circuit.
func (b *Breaker) RecordFailure(host string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	hs := b.ensureHost(host)
	hs.failures++
	if hs.failures >= b.opts.MaxFailures {
		hs.state = StateOpen
		hs.openedAt = time.Now()
	}
}

// State returns the current state for a host.
func (b *Breaker) State(host string) State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.ensureHost(host).state
}

func (b *Breaker) ensureHost(host string) *hostState {
	if _, ok := b.hosts[host]; !ok {
		b.hosts[host] = &hostState{state: StateClosed}
	}
	return b.hosts[host]
}
