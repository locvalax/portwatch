// Package remap translates raw port numbers to human-readable service names.
package remap

import (
	"fmt"
	"sync"
)

// Entry maps a port number to a service name and optional description.
type Entry struct {
	Port        int
	ServiceName string
	Description string
}

// Remapper holds a registry of port-to-service mappings.
type Remapper struct {
	mu      sync.RWMutex
	entries map[int]Entry
}

// New returns a Remapper pre-loaded with well-known service names.
func New() *Remapper {
	r := &Remapper{entries: make(map[int]Entry)}
	for _, e := range defaults {
		r.entries[e.Port] = e
	}
	return r
}

// Register adds or replaces a mapping for the given port.
func (r *Remapper) Register(e Entry) error {
	if e.Port < 1 || e.Port > 65535 {
		return fmt.Errorf("remap: port %d out of range", e.Port)
	}
	if e.ServiceName == "" {
		return fmt.Errorf("remap: service name must not be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[e.Port] = e
	return nil
}

// Lookup returns the service name for a port, or a numeric fallback.
func (r *Remapper) Lookup(port int) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if e, ok := r.entries[port]; ok {
		return e.ServiceName
	}
	return fmt.Sprintf("%d", port)
}

// LookupEntry returns the full Entry for a port and whether it was found.
func (r *Remapper) LookupEntry(port int) (Entry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.entries[port]
	return e, ok
}

// Delete removes a mapping. It is a no-op if the port is not registered.
func (r *Remapper) Delete(port int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, port)
}

// defaults contains common well-known port mappings.
var defaults = []Entry{
	{Port: 21, ServiceName: "ftp", Description: "File Transfer Protocol"},
	{Port: 22, ServiceName: "ssh", Description: "Secure Shell"},
	{Port: 23, ServiceName: "telnet", Description: "Telnet"},
	{Port: 25, ServiceName: "smtp", Description: "Simple Mail Transfer Protocol"},
	{Port: 53, ServiceName: "dns", Description: "Domain Name System"},
	{Port: 80, ServiceName: "http", Description: "Hypertext Transfer Protocol"},
	{Port: 110, ServiceName: "pop3", Description: "Post Office Protocol v3"},
	{Port: 143, ServiceName: "imap", Description: "Internet Message Access Protocol"},
	{Port: 443, ServiceName: "https", Description: "HTTP over TLS"},
	{Port: 3306, ServiceName: "mysql", Description: "MySQL Database"},
	{Port: 5432, ServiceName: "postgres", Description: "PostgreSQL Database"},
	{Port: 6379, ServiceName: "redis", Description: "Redis"},
	{Port: 8080, ServiceName: "http-alt", Description: "HTTP Alternate"},
	{Port: 27017, ServiceName: "mongodb", Description: "MongoDB"},
}
