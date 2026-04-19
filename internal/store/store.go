package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Result holds the state of a single scanned port.
type Result struct {
	Port  int    `json:"port"`
	State string `json:"state"`
}

// Record is a timestamped snapshot of scan results for a host.
type Record struct {
	Timestamp time.Time `json:"timestamp"`
	Ports     []Result  `json:"ports"`
}

// Store persists scan records on disk as JSON files.
type Store struct {
	dir string
}

// New creates a Store rooted at dir, creating it if necessary.
func New(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("store: mkdir %s: %w", dir, err)
	}
	return &Store{dir: dir}, nil
}

func (s *Store) hostFile(host string) string {
	return filepath.Join(s.dir, host+".json")
}

// Append adds a new record for host.
func (s *Store) Append(host string, ports []Result) error {
	records, _ := s.load(host)
	records = append(records, Record{Timestamp: time.Now().UTC(), Ports: ports})
	return s.save(host, records)
}

// Latest returns the most recent record for host.
func (s *Store) Latest(host string) (*Record, error) {
	records, err := s.load(host)
	if err != nil || len(records) == 0 {
		return nil, errors.New("store: no records for " + host)
	}
	r := records[len(records)-1]
	return &r, nil
}

// History returns up to limit records for host, newest first.
func (s *Store) History(host string, limit int) ([]Record, error) {
	records, err := s.load(host)
	if err != nil {
		return nil, err
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].Timestamp.After(records[j].Timestamp)
	})
	if limit > 0 && len(records) > limit {
		records = records[:limit]
	}
	return records, nil
}

func (s *Store) load(host string) ([]Record, error) {
	data, err := os.ReadFile(s.hostFile(host))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var records []Record
	return records, json.Unmarshal(data, &records)
}

func (s *Store) save(host string, records []Record) error {
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.hostFile(host), data, 0o644)
}
