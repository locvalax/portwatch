// Package sampler provides probabilistic scan sampling to reduce load
// by skipping a configurable fraction of scheduled scans.
package sampler

import (
	"fmt"
	"math/rand"
	"sync"
)

// Options configures the Sampler.
type Options struct {
	// Rate is the probability [0.0, 1.0] that a scan is allowed through.
	// 1.0 means every scan runs; 0.0 means no scans run.
	Rate float64
	// Seed is the random seed. 0 uses a default source.
	Seed int64
}

// DefaultOptions returns an Options that allows every scan.
func DefaultOptions() Options {
	return Options{Rate: 1.0}
}

// Sampler decides probabilistically whether a scan should proceed.
type Sampler struct {
	mu   sync.Mutex
	rate float64
	rng  *rand.Rand
}

// New creates a Sampler from opts. Returns an error if Rate is outside [0, 1].
func New(opts Options) (*Sampler, error) {
	if opts.Rate < 0 || opts.Rate > 1 {
		return nil, fmt.Errorf("sampler: rate must be in [0.0, 1.0], got %v", opts.Rate)
	}
	//nolint:gosec
	src := rand.NewSource(opts.Seed)
	return &Sampler{
		rate: opts.Rate,
		rng:  rand.New(src),
	}, nil
}

// Allow returns true if the scan for host should proceed given the configured rate.
func (s *Sampler) Allow(_ string) bool {
	if s.rate >= 1.0 {
		return true
	}
	if s.rate <= 0.0 {
		return false
	}
	s.mu.Lock()
	v := s.rng.Float64()
	s.mu.Unlock()
	return v < s.rate
}

// Rate returns the configured sampling rate.
func (s *Sampler) Rate() float64 {
	return s.rate
}
