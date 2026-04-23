package timeout

import (
	"fmt"
	"time"

	"github.com/user/portwatch/internal/config"
)

// FromConfig builds Options from the application config.
// If no timeout is configured the default deadline is used.
func FromConfig(cfg config.Config) (Options, error) {
	opts := DefaultOptions()

	if cfg.Timeout == "" {
		return opts, nil
	}

	d, err := time.ParseDuration(cfg.Timeout)
	if err != nil {
		return Options{}, fmt.Errorf("timeout: invalid duration %q: %w", cfg.Timeout, err)
	}

	if d <= 0 {
		return Options{}, fmt.Errorf("timeout: duration must be positive, got %s", cfg.Timeout)
	}

	opts.Deadline = d
	return opts, nil
}
