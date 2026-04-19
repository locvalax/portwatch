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

// HasPort reports whether the entry contains the given port and protocol.
func (e *Entry) HasPort(port int, proto string) bool {
	for _, p := range e.Ports {
		if p.Port == port && p.Proto == proto {
			return true
		}
	}
	return false
}
