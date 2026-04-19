package store

import (
	"encoding/json"
	"os"
	"time"
)

// Snapshot represents the state of open ports on a host at a point in time.
type Snapshot struct {
	Host      string    `json:"host"`
	Ports     []int     `json:"ports"`
	ScannedAt time.Time `json:"scanned_at"`
}

// Store persists and retrieves port snapshots from a JSON file.
type Store struct {
	path string
}

// New creates a new Store backed by the given file path.
func New(path string) *Store {
	return &Store{path: path}
}

// Load reads all snapshots from the store file.
// Returns an empty slice if the file does not exist.
func (s *Store) Load() ([]Snapshot, error) {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return []Snapshot{}, nil
	}
	if err != nil {
		return nil, err
	}
	var snapshots []Snapshot
	if err := json.Unmarshal(data, &snapshots); err != nil {
		return nil, err
	}
	return snapshots, nil
}

// Save writes the given snapshots to the store file, overwriting any existing data.
func (s *Store) Save(snapshots []Snapshot) error {
	data, err := json.MarshalIndent(snapshots, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

// Latest returns the most recent snapshot for the given host, or nil if none exists.
func (s *Store) Latest(host string) (*Snapshot, error) {
	snapshots, err := s.Load()
	if err != nil {
		return nil, err
	}
	var latest *Snapshot
	for i := range snapshots {
		if snapshots[i].Host == host {
			if latest == nil || snapshots[i].ScannedAt.After(latest.ScannedAt) {
				latest = &snapshots[i]
			}
		}
	}
	return latest, nil
}

// Append adds a new snapshot and persists the updated list.
func (s *Store) Append(snap Snapshot) error {
	snapshots, err := s.Load()
	if err != nil {
		return err
	}
	snapshots = append(snapshots, snap)
	return s.Save(snapshots)
}
