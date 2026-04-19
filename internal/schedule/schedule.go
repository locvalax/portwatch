package schedule

import (
	"context"
	"log"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/scanner"
	"github.com/user/portwatch/internal/store"
)

// Runner periodically scans hosts and alerts on port changes.
type Runner struct {
	hosts    []string
	interval time.Duration
	store    *store.Store
	alerter  *alert.Alerter
	opts     scanner.Options
}

// New creates a new Runner.
func New(hosts []string, interval time.Duration, s *store.Store, a *alert.Alerter, opts scanner.Options) *Runner {
	return &Runner{
		hosts:    hosts,
		interval: interval,
		store:    s,
		alerter:  a,
		opts:     opts,
	}
}

// Run starts the scheduling loop, blocking until ctx is cancelled.
func (r *Runner) Run(ctx context.Context) {
	r.tick()
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			r.tick()
		case <-ctx.Done():
			return
		}
	}
}

func (r *Runner) tick() {
	for _, host := range r.hosts {
		ports, err := scanner.Scan(host, r.opts)
		if err != nil {
			log.Printf("scan error for %s: %v", host, err)
			continue
		}
		prev, _ := r.store.Latest(host)
		if err := r.store.Append(host, ports); err != nil {
			log.Printf("store error for %s: %v", host, err)
			continue
		}
		if err := r.alerter.Notify(host, prev, ports); err != nil {
			log.Printf("alert error for %s: %v", host, err)
		}
	}
}
