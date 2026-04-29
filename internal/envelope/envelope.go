// Package envelope wraps scan results with metadata for downstream processing.
package envelope

import (
	"time"

	"github.com/yourorg/portwatch/internal/store"
)

// Envelope wraps a store.Entry with additional routing and provenance metadata.
type Envelope struct {
	Entry     store.Entry
	Host      string
	ScannedAt time.Time
	Source    string
	Labels    map[string]string
	Seq       uint64
}

// Options configures the Builder.
type Options struct {
	Source string
	Labels map[string]string
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		Source: "portwatch",
		Labels: map[string]string{},
	}
}

// Builder constructs Envelopes from scan entries.
type Builder struct {
	opts Options
	seq  uint64
}

// New creates a Builder with the given options.
func New(opts Options) *Builder {
	if opts.Source == "" {
		opts.Source = DefaultOptions().Source
	}
	if opts.Labels == nil {
		opts.Labels = map[string]string{}
	}
	return &Builder{opts: opts}
}

// Wrap creates an Envelope from the given host and entry, stamping it with
// the current time and an auto-incremented sequence number.
func (b *Builder) Wrap(host string, entry store.Entry) Envelope {
	b.seq++
	labels := make(map[string]string, len(b.opts.Labels))
	for k, v := range b.opts.Labels {
		labels[k] = v
	}
	return Envelope{
		Entry:     entry,
		Host:      host,
		ScannedAt: time.Now().UTC(),
		Source:    b.opts.Source,
		Labels:    labels,
		Seq:       b.seq,
	}
}

// WithLabel returns a copy of the Envelope with an additional label applied.
func WithLabel(e Envelope, key, value string) Envelope {
	newLabels := make(map[string]string, len(e.Labels)+1)
	for k, v := range e.Labels {
		newLabels[k] = v
	}
	newLabels[key] = value
	e.Labels = newLabels
	return e
}
