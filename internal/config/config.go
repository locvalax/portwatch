package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds all portwatch runtime settings.
type Config struct {
	Hosts        []string `yaml:"hosts"`
	IntervalSecs int      `yaml:"interval_secs"`
	StorePath    string   `yaml:"store_path"`
	AlertOutput  string   `yaml:"alert_output"` // stdout | file
	AlertFile    string   `yaml:"alert_file"`
	ReportFormat string   `yaml:"report_format"` // text | json
	Filter       struct {
		IncludePorts []string `yaml:"include_ports"`
		ExcludePorts []string `yaml:"exclude_ports"`
		MinPort      int      `yaml:"min_port"`
		MaxPort      int      `yaml:"max_port"`
	} `yaml:"filter"`
}

// Defaults returns a Config with sensible default values.
func Defaults() Config {
	return Config{
		Hosts:        []string{"localhost"},
		IntervalSecs: 60,
		StorePath:    "portwatch.db",
		AlertOutput:  "stdout",
		ReportFormat: "text",
	}
}

// Load reads a YAML config file and merges it over the defaults.
func Load(path string) (Config, error) {
	cfg := Defaults()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
