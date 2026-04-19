package report

import (
	"fmt"

	"github.com/user/portwatch/internal/store"
)

// Builder collects scan results from the store and builds report entries.
type Builder struct {
	st *store.Store
}

// NewBuilder creates a Builder backed by the given store.
func NewBuilder(st *store.Store) *Builder {
	return &Builder{st: st}
}

// ForHost returns up to limit entries for the given host, newest first.
func (b *Builder) ForHost(host string, limit int) ([]Entry, error) {
	records, err := b.st.History(host, limit)
	if err != nil {
		return nil, fmt.Errorf("report: history for %s: %w", host, err)
	}
	entries := make([]Entry, 0, len(records))
	for _, rec := range records {
		entries = append(entries, Entry{
			Host:      host,
			Timestamp: rec.Timestamp,
			Ports:     rec.Ports,
		})
	}
	return entries, nil
}
