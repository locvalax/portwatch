package checkpoint

import (
	"context"
	"errors"
	"fmt"

	"github.com/user/portwatch/internal/store"
)

// Scanner is the interface satisfied by scanner implementations.
type Scanner interface {
	Scan(ctx context.Context, host string) ([]int, error)
}

// CheckpointedScanner wraps an inner Scanner and automatically saves a
// checkpoint after every successful scan. On the next scan for the same host
// the previous checkpoint is available via Load so callers can diff without
// consulting the full history store.
type CheckpointedScanner struct {
	inner Scanner
	cp    *Store
}

// NewCheckpointedScanner returns a CheckpointedScanner that delegates to inner
// and persists results in cp.
func NewCheckpointedScanner(inner Scanner, cp *Store) *CheckpointedScanner {
	return &CheckpointedScanner{inner: inner, cp: cp}
}

// Scan runs the inner scanner and, on success, saves the checkpoint for host.
// The scan result is returned unchanged; checkpoint failures are surfaced as
// wrapped errors so the caller can decide whether to treat them as fatal.
func (c *CheckpointedScanner) Scan(ctx context.Context, host string) ([]int, error) {
	ports, err := c.inner.Scan(ctx, host)
	if err != nil {
		return nil, err
	}
	if saveErr := c.cp.Save(Entry{Host: host, Ports: ports}); saveErr != nil {
		return ports, fmt.Errorf("checkpoint save: %w", saveErr)
	}
	return ports, nil
}

// Previous returns the last checkpointed port list for host, or an empty slice
// if no checkpoint exists yet (ErrNotFound is silently swallowed).
func (c *CheckpointedScanner) Previous(host string) ([]int, error) {
	e, err := c.cp.Load(host)
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return e.Ports, nil
}

// ensure CheckpointedScanner satisfies Scanner at compile time.
var _ Scanner = (*CheckpointedScanner)(nil)

// storeEntry is a local alias kept for future store integration.
type storeEntry = store.Entry
