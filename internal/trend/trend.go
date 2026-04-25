package trend

import (
	"math"
	"sort"
	"sync"
	"time"
)

// Direction indicates whether a port count is growing, shrinking, or stable.
type Direction string

const (
	Growing  Direction = "growing"
	Shrinking Direction = "shrinking"
	Stable   Direction = "stable"
)

// Point is a single observation of open-port count for a host.
type Point struct {
	At    time.Time
	Count int
}

// Summary describes the trend for a single host.
type Summary struct {
	Host      string
	Direction Direction
	Slope     float64 // ports per hour
	Points    int
}

// Analyzer tracks port-count history and computes linear trends.
type Analyzer struct {
	mu      sync.Mutex
	window  time.Duration
	history map[string][]Point
}

// New returns an Analyzer that retains observations within window.
func New(window time.Duration) *Analyzer {
	if window <= 0 {
		window = 24 * time.Hour
	}
	return &Analyzer{
		window:  window,
		history: make(map[string][]Point),
	}
}

// Record adds a port-count observation for host at the given time.
func (a *Analyzer) Record(host string, count int, at time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()

	cutoff := at.Add(-a.window)
	pts := append(a.history[host], Point{At: at, Count: count})

	// prune old points
	keep := pts[:0]
	for _, p := range pts {
		if !p.At.Before(cutoff) {
			keep = append(keep, p)
		}
	}
	sort.Slice(keep, func(i, j int) bool { return keep[i].At.Before(keep[j].At) })
	a.history[host] = keep
}

// Summarize returns a trend Summary for host, or ok=false if insufficient data.
func (a *Analyzer) Summarize(host string) (Summary, bool) {
	a.mu.Lock()
	pts := a.history[host]
	a.mu.Unlock()

	if len(pts) < 2 {
		return Summary{}, false
	}

	slope := leastSquaresSlope(pts) // ports per nanosecond
	slopePerHour := slope * float64(time.Hour)

	dir := Stable
	switch {
	case slopePerHour > 0.5:
		dir = Growing
	case slopePerHour < -0.5:
		dir = Shrinking
	}

	return Summary{
		Host:      host,
		Direction: dir,
		Slope:     math.Round(slopePerHour*100) / 100,
		Points:    len(pts),
	}, true
}

func leastSquaresSlope(pts []Point) float64 {
	n := float64(len(pts))
	t0 := pts[0].At.UnixNano()
	var sumX, sumY, sumXY, sumX2 float64
	for _, p := range pts {
		x := float64(p.At.UnixNano() - t0)
		y := float64(p.Count)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	denom := n*sumX2 - sumX*sumX
	if denom == 0 {
		return 0
	}
	return (n*sumXY - sumX*sumY) / denom
}
