package remap

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type fileEntry struct {
	Port        int    `yaml:"port"`
	ServiceName string `yaml:"service"`
	Description string `yaml:"description"`
}

type fileSchema struct {
	Mappings []fileEntry `yaml:"mappings"`
}

// LoadFile parses a YAML file of port mappings and registers them in r.
// Existing entries are overwritten; entries not in the file are preserved.
//
// Example YAML:
//
//	mappings:
//	  - port: 8443
//	    service: https-alt
//	    description: HTTPS on alternate port
func LoadFile(path string, r *Remapper) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("remap: read %s: %w", path, err)
	}

	var schema fileSchema
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return fmt.Errorf("remap: parse %s: %w", path, err)
	}

	for _, fe := range schema.Mappings {
		e := Entry{
			Port:        fe.Port,
			ServiceName: fe.ServiceName,
			Description: fe.Description,
		}
		if err := r.Register(e); err != nil {
			return fmt.Errorf("remap: register port %d: %w", fe.Port, err)
		}
	}
	return nil
}
