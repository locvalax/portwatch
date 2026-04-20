package tag

import "sync"

// Registry maps host names to their tag sets.
type Registry struct {
	mu   sync.RWMutex
	hosts map[string]Set
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{hosts: make(map[string]Set)}
}

// Set stores (or replaces) the tag set for a host.
func (r *Registry) Set(host string, s Set) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hosts[host] = s
}

// Get returns the tag set for a host. Returns nil, false if unknown.
func (r *Registry) Get(host string) (Set, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.hosts[host]
	return s, ok
}

// Delete removes the tag set for a host.
func (r *Registry) Delete(host string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.hosts, host)
}

// Filter returns all hosts whose tag sets match the given filter.
func (r *Registry) Filter(filter Set) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var matched []string
	for host, s := range r.hosts {
		if s.Match(filter) {
			matched = append(matched, host)
		}
	}
	return matched
}

// Hosts returns all registered host names.
func (r *Registry) Hosts() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	hosts := make([]string, 0, len(r.hosts))
	for h := range r.hosts {
		hosts = append(hosts, h)
	}
	return hosts
}
