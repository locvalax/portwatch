package profile

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// File represents the top-level YAML structure for profiles.
type File struct {
	Profiles []*Profile `yaml:"profiles"`
}

// LoadFile reads a YAML file and registers all profiles into r.
// Returns an error if the file cannot be read or parsed.
func LoadFile(path string, r *Registry) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("profile: read %s: %w", path, err)
	}

	var f File
	if err := yaml.Unmarshal(data, &f); err != nil {
		return fmt.Errorf("profile: parse %s: %w", path, err)
	}

	if len(f.Profiles) == 0 {
		return fmt.Errorf("profile: no profiles found in %s", path)
	}

	for _, p := range f.Profiles {
		if err := r.Set(p); err != nil {
			return fmt.Errorf("profile: register %q: %w", p.Name, err)
		}
	}
	return nil
}
