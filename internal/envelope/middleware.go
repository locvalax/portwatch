package envelope

import (
	"context"

	"github.com/yourorg/portwatch/internal/store"
)

// Scanner is the interface satisfied by port scanners.
type Scanner interface {
	Scan(ctx context.Context, host string) (store.Entry, error)
}

// WrappingScanner wraps scan results in Envelopes and forwards them to a
// registered handler before returning the raw entry to the caller.
type WrappingScanner struct {
	inner   Scanner
	builder *Builder
	sink    func(Envelope)
}

// NewWrappingScanner creates a WrappingScanner that envelopes every successful
// scan result and passes it to sink. The original entry is still returned to
// the caller unchanged.
func NewWrappingScanner(inner Scanner, builder *Builder, sink func(Envelope)) *WrappingScanner {
	if sink == nil {
		sink = func(Envelope) {}
	}
	return &WrappingScanner{inner: inner, builder: builder, sink: sink}
}

// Scan delegates to the inner scanner and, on success, wraps the result and
// calls the sink before returning.
func (w *WrappingScanner) Scan(ctx context.Context, host string) (store.Entry, error) {
	entry, err := w.inner.Scan(ctx, host)
	if err != nil {
		return entry, err
	}
	env := w.builder.Wrap(host, entry)
	w.sink(env)
	return entry, nil
}
