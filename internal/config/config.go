package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level portwatch configuration.
type Config struct {
	Hosts    []string      `yaml:"hosts"`
	Ports    []int         `yaml:"ports"`
	Interval time.Duration `yaml:"interval"`
	StorePath string       `yaml:"store_path"`
	Alert    AlertConfig   `yaml:"alert"`
}

// AlertConfig configures alerting output.
type AlertConfig struct {
	Output string `yaml:"output"` // "stdout" or a file path
}

// Defaults returns a Config populated with sensible defaults.
func Defaults() Config {
	return Config{
		Hosts:     []string{"localhost"},
		Ports:     []int{22, 80, 443, 8080},
		Interval:  60 * time.Second,
		StorePath: "portwatch.db",
		Alert: AlertConfig{
			Output: "stdout",
		},
	}
}

// Load reads a YAML config file from path and merges it over defaults.
func Load(path string) (Config, error) {
	cfg := Defaults()

	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	if err := dec.Decode(&cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
