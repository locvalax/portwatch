package topology

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type nodeConfig struct {
	Host  string            `yaml:"host"`
	Group string            `yaml:"group"`
	Tags  map[string]string `yaml:"tags"`
	Peers []string          `yaml:"peers"`
}

type fileConfig struct {
	Nodes []nodeConfig `yaml:"nodes"`
}

// LoadFile parses a YAML topology file and populates a Graph.
//
// Example YAML:
//
//	nodes:
//	  - host: web1
//	    group: web
//	    tags: {env: prod}
//	    peers: [db1]
func LoadFile(path string) (*Graph, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("topology: read file: %w", err)
	}
	var cfg fileConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("topology: parse yaml: %w", err)
	}
	g := New()
	for _, nc := range cfg.Nodes {
		if err := g.AddNode(Node{
			Host:  nc.Host,
			Group: nc.Group,
			Tags:  nc.Tags,
		}); err != nil {
			return nil, err
		}
	}
	for _, nc := range cfg.Nodes {
		for _, peer := range nc.Peers {
			if err := g.AddPeer(nc.Host, peer); err != nil {
				return nil, err
			}
		}
	}
	return g, nil
}
