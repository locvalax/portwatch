// Package pipeline wires together the scanning, filtering, deduplication,
// alerting, and notification stages into a single reusable execution unit.
package pipeline

import (
	"context"
	"fmt"
	"log"

	"github.com/yourorg/portwatch/internal/alert"
	"github.com/yourorg/portwatch/internal/audit"
	"github.com/yourorg/portwatch/internal/dedupe"
	"github.com/yourorg/portwatch/internal/filter"
	"github.com/yourorg/portwatch/internal/metrics"
	"github.com/yourorg/portwatch/internal/notify"
	"github.com/yourorg/portwatch/internal/scanner"
	"github.com/yourorg/portwatch/internal/store"
	"github.com/yourorg/portwatch/internal/suppress"
)

// Scanner is the interface satisfied by scanner.Scan and any middleware wrapping it.
type Scanner interface {
	Scan(ctx context.Context, host string) ([]int, error)
}

// Options configures the pipeline behaviour.
type Options struct {
	Scanner  Scanner
	Store    *store.Store
	Filter   *filter.Filter
	Dedupe   *dedupe.Deduper
	Suppress *suppress.Suppressor
	Alerter  *alert.Alerter
	Notifier *notify.Notifier
	Auditor  *audit.Auditor
	Metrics  *metrics.Metrics
}

// Pipeline executes a full scan-to-alert cycle for a single host.
type Pipeline struct {
	opts Options
}

// New creates a Pipeline from the provided options.
// All fields in opts are optional; nil values cause that stage to be skipped.
func New(opts Options) *Pipeline {
	return &Pipeline{opts: opts}
}

// Run scans host, applies filtering and deduplication, persists the result,
// computes the diff against the previous snapshot, and dispatches alerts and
// notifications. It returns the first error encountered, if any.
func (p *Pipeline) Run(ctx context.Context, host string) error {
	// --- scan ---
	ports, err := p.opts.Scanner.Scan(ctx, host)
	if err != nil {
		if p.opts.Metrics != nil {
			p.opts.Metrics.RecordError(host)
		}
		return fmt.Errorf("scan %s: %w", host, err)
	}

	// --- filter ---
	if p.opts.Filter != nil {
		ports = p.opts.Filter.Apply(ports)
	}

	// --- build entry and persist ---
	entry := store.Entry{Host: host, Ports: ports}

	if p.opts.Store != nil {
		if err := p.opts.Store.Append(entry); err != nil {
			return fmt.Errorf("store append %s: %w", host, err)
		}
	}

	if p.opts.Metrics != nil {
		p.opts.Metrics.RecordScan(host)
	}

	// --- deduplicate ---
	if p.opts.Dedupe != nil && p.opts.Dedupe.IsDuplicate(entry) {
		log.Printf("pipeline: %s ports unchanged, skipping alert", host)
		return nil
	}

	// --- diff against previous ---
	var diff store.Diff
	if p.opts.Store != nil {
		prev, err := p.opts.Store.Previous(host)
		if err == nil {
			diff = store.Compare(prev, entry)
		}
	}

	// --- suppression ---
	if p.opts.Suppress != nil && p.opts.Suppress.IsSuppressed(host, diff) {
		log.Printf("pipeline: %s diff suppressed", host)
		return nil
	}

	// --- alert ---
	if p.opts.Alerter != nil {
		if err := p.opts.Alerter.Notify(host, diff); err != nil {
			log.Printf("pipeline: alert error for %s: %v", host, err)
		}
		if p.opts.Metrics != nil {
			p.opts.Metrics.RecordAlert(host)
		}
	}

	// --- audit ---
	if p.opts.Auditor != nil {
		if err := p.opts.Auditor.LogPortChange(host, diff); err != nil {
			log.Printf("pipeline: audit error for %s: %v", host, err)
		}
	}

	// --- notify (webhooks / slack) ---
	if p.opts.Notifier != nil {
		if err := p.opts.Notifier.Dispatch(ctx, host, diff); err != nil {
			log.Printf("pipeline: notify error for %s: %v", host, err)
		}
	}

	return nil
}

// RunAll executes Run for every host in hosts, collecting all errors.
// Execution continues even when individual hosts fail.
func (p *Pipeline) RunAll(ctx context.Context, hosts []string) []error {
	var errs []error
	for _, h := range hosts {
		if err := p.Run(ctx, h); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// DefaultScanner wraps scanner.Scan so it satisfies the Scanner interface.
type DefaultScanner struct {
	Opts scanner.Options
}

// Scan implements Scanner using the package-level scanner.Scan function.
func (d DefaultScanner) Scan(ctx context.Context, host string) ([]int, error) {
	return scanner.Scan(ctx, host, d.Opts)
}
