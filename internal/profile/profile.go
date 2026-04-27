package profile

import (
	"errors"
	"sync"
	"time"
)

// Profile holds a named scanning profile with per-host overrides.
type Profile struct {
	Name     string            `yaml:"name"`
	Hosts    []string          `yaml:"hosts"`
	Ports    string            `yaml:"ports"`
	Interval time.Duration     `yaml:"interval"`
	Tags     map[string]string `yaml:"tags"`
}

// Registry stores named profiles.
type Registry struct {
	mu       sync.RWMutex
	profiles map[string]*Profile
}

// New returns an empty Registry.
func New() *Registry {
	return &Registry{
		profiles: make(map[string]*Profile),
	}
}

// Set adds or replaces a profile by name.
func (r *Registry) Set(p *Profile) error {
	if p == nil {
		return errors.New("profile: nil profile")
	}
	if p.Name == "" {
		return errors.New("profile: name is required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.profiles[p.Name] = p
	return nil
}

// Get retrieves a profile by name. Returns false if not found.
func (r *Registry) Get(name string) (*Profile, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.profiles[name]
	return p, ok
}

// Delete removes a profile by name.
func (r *Registry) Delete(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.profiles, name)
}

// All returns a snapshot of all profiles.
func (r *Registry) All() []*Profile {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Profile, 0, len(r.profiles))
	for _, p := range r.profiles {
		out = append(out, p)
	}
	return out
}
