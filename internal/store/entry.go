package store

import "time"

// PortResult holds a single open port result.
type PortResult struct {
	Port  int    `json:"port"`
	Proto string `json:"proto"`
}

// Entry represents a scan result snapshot for a single host.
type Entry struct {
	Host      string       `json:"host"`
	ScannedAt time.Time    `json:"scanned_at"`
	Ports     []PortResult `json:"ports"`
}
