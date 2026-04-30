package digest

import (
	"fmt"
	"time"

	"github.com/user/portwatch/internal/config"
)

// FromConfig derives digest Options from the global application config.
// It respects the following config keys (with fallback to defaults):
//
//	digest.window  — duration string, e.g. "24h"
func FromConfig(cfg config.Config) (Options, error) {
	opts := DefaultOptions()

	if raw, ok := cfg.Extras["digest.window"]; ok {
		s, ok := raw.(string)
		if !ok {
			return opts, fmt.Errorf("digest: digest.window must be a string")
		}
		d, err := time.ParseDuration(s)
		if err != nil {
			return opts, fmt.Errorf("digest: invalid digest.window %q: %w", s, err)
		}
		if d <= 0 {
			return opts, fmt.Errorf("digest: digest.window must be positive")
		}
		opts.Window = d
	}

	return opts, nil
}
