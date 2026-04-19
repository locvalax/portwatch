package baseline

import (
	"encoding/json"
	"os"
	"time"
)

// Snapshot represents a saved baseline of open ports for a host.
type Snapshot struct {
	Host      string    `json:"host"`
	Ports     []int     `json:"ports"`
	CreatedAt time.Time `json:"created_at"`
}

// Manager handles reading and writing baseline snapshots.
type Manager struct {
	path string
}

// New returns a Manager backed by the given file path.
func New(path string) *Manager {
	return &Manager{path: path}
}

// Save writes the snapshot to disk, overwriting any existing baseline.
func (m *Manager) Save(s Snapshot) error {
	f, err := os.Create(m.path)
	if err != nil {
		return err
	}
	defer f.Close()
	s.CreatedAt = time.Now().UTC()
	return json.NewEncoder(f).Encode(s)
}

// Load reads the snapshot from disk.
func (m *Manager) Load() (Snapshot, error) {
	f, err := os.Open(m.path)
	if err != nil {
		return Snapshot{}, err
	}
	defer f.Close()
	var s Snapshot
	if err := json.NewDecoder(f).Decode(&s); err != nil {
		return Snapshot{}, err
	}
	return s, nil
}

// Exists reports whether a baseline file is present.
func (m *Manager) Exists() bool {
	_, err := os.Stat(m.path)
	return err == nil
}
