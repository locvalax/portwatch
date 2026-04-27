package shadow

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Config holds shadow-mode configuration, typically loaded from the main
// portwatch config file under the `shadow:` key.
type Config struct {
	// Enabled activates shadow-mode scanning.
	Enabled bool `yaml:"enabled"`
	// Hosts lists the secondary hosts to shadow-scan. If empty, the same
	// primary hosts are used.
	Hosts []string `yaml:"hosts"`
	// Timeout is the per-scan deadline for the shadow scanner.
	Timeout time.Duration `yaml:"timeout"`
	// LogFile writes divergences to a file instead of stderr. Optional.
	LogFile string `yaml:"log_file"`
}

// DefaultConfig returns safe defaults.
func DefaultConfig() Config {
	return Config{
		Enabled: false,
		Timeout: 5 * time.Second,
	}
}

// Writer returns an io.Writer for divergence logging based on the config.
// The caller is responsible for closing the file if one is opened.
func (c Config) Writer() (io.Writer, func(), error) {
	if c.LogFile == "" {
		return os.Stderr, func() {}, nil
	}
	f, err := os.OpenFile(c.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, func() {}, fmt.Errorf("shadow: open log file: %w", err)
	}
	return f, func() { f.Close() }, nil
}
