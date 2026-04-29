package probe

import (
	"fmt"
	"time"

	"github.com/user/portwatch/internal/config"
)

// FromConfig derives probe Options from the application config.
func FromConfig(cfg config.Config) (Options, error) {
	opts := DefaultOptions()

	if cfg.ProbeTimeout != "" {
		d, err := time.ParseDuration(cfg.ProbeTimeout)
		if err != nil {
			return Options{}, fmt.Errorf("probe: invalid timeout %q: %w", cfg.ProbeTimeout, err)
		}
		opts.Timeout = d
	}

	if cfg.ProbeConcurrency > 0 {
		opts.Concurrency = cfg.ProbeConcurrency
	}

	return opts, nil
}
