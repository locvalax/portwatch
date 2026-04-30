package redact

import (
	"context"

	"github.com/sgreben/portwatch/internal/store"
)

// Scanner is the interface satisfied by scanner implementations.
type Scanner interface {
	Scan(ctx context.Context, host string) (store.Entry, error)
}

// RedactedScanner wraps a Scanner and masks host names in returned entries.
type RedactedScanner struct {
	inner   Scanner
	redactor *Redactor
}

// NewRedactedScanner returns a Scanner that redacts host fields in results.
func NewRedactedScanner(inner Scanner, r *Redactor) *RedactedScanner {
	return &RedactedScanner{inner: inner, redactor: r}
}

// Scan delegates to the inner scanner and replaces the host in the returned entry.
func (s *RedactedScanner) Scan(ctx context.Context, host string) (store.Entry, error) {
	entry, err := s.inner.Scan(ctx, host)
	if err != nil {
		return store.Entry{}, err
	}
	entry.Host = s.redactor.Host(entry.Host)
	return entry, nil
}
