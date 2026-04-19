package ratelimit

import (
	"fmt"
	"time"

	"github.com/user/portwatch/internal/config"
)

const defaultInterval = 60 * time.Second

// FromConfig reads rate limit settings from the app config.
// It expects a [ratelimit] section with an "interval" key (e.g. "30s").
func FromConfig(cfg *config.Config) (*Limiter, error) {
	raw, ok := cfg.Section("ratelimit")["interval"]
	if !ok {
		return New(defaultInterval), nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return nil, fmt.Errorf("ratelimit: invalid interval %q: %w", raw, err)
	}
	if d <= 0 {
		return nil, fmt.Errorf("ratelimit: interval must be positive, got %v", d)
	}
	return New(d), nil
}
