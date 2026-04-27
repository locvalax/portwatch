package backoff

import "time"

// Config mirrors the YAML/env representation of back-off settings.
type Config struct {
	InitialIntervalMs int     `yaml:"initial_interval_ms"`
	MaxIntervalMs     int     `yaml:"max_interval_ms"`
	Multiplier        float64 `yaml:"multiplier"`
	MaxAttempts       int     `yaml:"max_attempts"`
}

// FromConfig converts a Config into Options, falling back to defaults for
// any zero-valued fields.
func FromConfig(c Config) Options {
	defaults := DefaultOptions()

	opts := Options{
		Multiplier:  defaults.Multiplier,
		MaxAttempts: defaults.MaxAttempts,
	}

	if c.InitialIntervalMs > 0 {
		opts.InitialInterval = time.Duration(c.InitialIntervalMs) * time.Millisecond
	} else {
		opts.InitialInterval = defaults.InitialInterval
	}

	if c.MaxIntervalMs > 0 {
		opts.MaxInterval = time.Duration(c.MaxIntervalMs) * time.Millisecond
	} else {
		opts.MaxInterval = defaults.MaxInterval
	}

	if c.Multiplier > 1 {
		opts.Multiplier = c.Multiplier
	}

	if c.MaxAttempts > 0 {
		opts.MaxAttempts = c.MaxAttempts
	}

	return opts
}
