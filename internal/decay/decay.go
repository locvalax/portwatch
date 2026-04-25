// Package decay implements an exponential decay scorer for port scan results.
// Older observations contribute less weight to the current score, allowing
// the system to naturally "forget" stale port states over time.
package decay

import (
	"math"
	"sync"
	"time"
)

// DefaultOptions returns sensible defaults for the decay scorer.
func DefaultOptions() Options {
	return Options{
		HalfLife: 5 * time.Minute,
		Floor:    0.01,
	}
}

// Options configures the decay scorer.
type Options struct {
	// HalfLife is the duration after which a score decays to half its value.
	HalfLife time.Duration
	// Floor is the minimum score below which an entry is considered expired.
	Floor float64
}

// entry holds the current score and last update time for a key.
type entry struct {
	score     float64
	updatedAt time.Time
}

// Scorer tracks exponentially decaying scores keyed by host+port pairs.
type Scorer struct {
	mu   sync.Mutex
	opts Options
	data map[string]*entry
	now  func() time.Time
}

// New creates a new Scorer with the given options.
func New(opts Options) *Scorer {
	if opts.HalfLife <= 0 {
		opts.HalfLife = DefaultOptions().HalfLife
	}
	if opts.Floor <= 0 {
		opts.Floor = DefaultOptions().Floor
	}
	return &Scorer{
		opts: opts,
		data: make(map[string]*entry),
		now:  time.Now,
	}
}

// Observe records a new observation for the given key, boosting its score by delta.
func (s *Scorer) Observe(key string, delta float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	e, ok := s.data[key]
	if !ok {
		s.data[key] = &entry{score: delta, updatedAt: now}
		return
	}
	e.score = s.decayed(e.score, e.updatedAt, now) + delta
	e.updatedAt = now
}

// Score returns the current decayed score for the given key.
// Returns 0 if the key is unknown or its score has fallen below the floor.
func (s *Scorer) Score(key string) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.data[key]
	if !ok {
		return 0
	}
	v := s.decayed(e.score, e.updatedAt, s.now())
	if v < s.opts.Floor {
		delete(s.data, key)
		return 0
	}
	return v
}

// Reset removes the score entry for the given key.
func (s *Scorer) Reset(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

// decayed applies exponential decay: score * 2^(-elapsed/halfLife).
func (s *Scorer) decayed(score float64, from, to time.Time) float64 {
	elapsed := to.Sub(from).Seconds()
	half := s.opts.HalfLife.Seconds()
	return score * math.Pow(2, -elapsed/half)
}
