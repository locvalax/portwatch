package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/user/portwatch/internal/store"
)

// Snapshot represents a point-in-time capture of port state for a host.
type Snapshot struct {
	Host      string    `json:"host"`
	Ports     []int     `json:"ports"`
	CapturedAt time.Time `json:"captured_at"`
}

// Manager handles saving and loading snapshots to disk.
type Manager struct {
	dir string
}

// New returns a Manager that stores snapshots under dir.
func New(dir string) (*Manager, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("snapshot: create dir: %w", err)
	}
	return &Manager{dir: dir}, nil
}

// Save writes the latest entry for host from st as a snapshot file.
func (m *Manager) Save(host string, st *store.Store) error {
	entry, err := st.Latest(host)
	if err != nil {
		return fmt.Errorf("snapshot: latest for %s: %w", host, err)
	}

	snap := Snapshot{
		Host:       host,
		Ports:      entry.Ports,
		CapturedAt: time.Now().UTC(),
	}

	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return fmt.Errorf("snapshot: marshal: %w", err)
	}

	path := m.path(host)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("snapshot: write %s: %w", path, err)
	}
	return nil
}

// Load reads the snapshot for host from disk.
func (m *Manager) Load(host string) (*Snapshot, error) {
	path := m.path(host)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("snapshot: no snapshot for %s", host)
		}
		return nil, fmt.Errorf("snapshot: read %s: %w", path, err)
	}

	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("snapshot: unmarshal: %w", err)
	}
	return &snap, nil
}

// Exists reports whether a snapshot file exists for host.
func (m *Manager) Exists(host string) bool {
	_, err := os.Stat(m.path(host))
	return err == nil
}

func (m *Manager) path(host string) string {
	safe := filepath.Clean(host)
	return filepath.Join(m.dir, safe+".json")
}
