// Package checkpoint persists the last-known scan state for each host so that
// portwatch can resume comparisons across process restarts without requiring a
// full history replay.
package checkpoint

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ErrNotFound is returned when no checkpoint exists for the requested host.
var ErrNotFound = errors.New("checkpoint: no entry for host")

// Entry holds the persisted state for a single host.
type Entry struct {
	Host      string    `json:"host"`
	Ports     []int     `json:"ports"`
	RecordedAt time.Time `json:"recorded_at"`
}

// Store manages checkpoint files on disk.
type Store struct {
	mu  sync.RWMutex
	dir string
}

// New returns a Store that persists checkpoints under dir.
// The directory is created if it does not exist.
func New(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("checkpoint: create dir: %w", err)
	}
	return &Store{dir: dir}, nil
}

// Save writes the entry for the given host to disk, overwriting any previous
// checkpoint for that host.
func (s *Store) Save(e Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	e.RecordedAt = time.Now().UTC()
	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("checkpoint: marshal: %w", err)
	}
	return os.WriteFile(s.path(e.Host), data, 0o644)
}

// Load retrieves the last checkpoint for host. Returns ErrNotFound if none
// exists yet.
func (s *Store) Load(host string) (Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.path(host))
	if errors.Is(err, os.ErrNotExist) {
		return Entry{}, ErrNotFound
	}
	if err != nil {
		return Entry{}, fmt.Errorf("checkpoint: read: %w", err)
	}
	var e Entry
	if err := json.Unmarshal(data, &e); err != nil {
		return Entry{}, fmt.Errorf("checkpoint: unmarshal: %w", err)
	}
	return e, nil
}

// Delete removes the checkpoint for host. It is a no-op if none exists.
func (s *Store) Delete(host string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := os.Remove(s.path(host))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (s *Store) path(host string) string {
	// Replace characters that are invalid in filenames.
	safe := filepath.Clean(host)
	for _, r := range []string{"/", ":", "\\"} {
		safe = replaceAll(safe, r, "_")
	}
	return filepath.Join(s.dir, safe+".json")
}

func replaceAll(s, old, new string) string {
	out := []byte(s)
	for i := 0; i < len(out); i++ {
		if string(out[i]) == old {
			out[i] = new[0]
		}
	}
	return string(out)
}
