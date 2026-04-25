package quota

import (
	"time"

	"github.com/user/portwatch/internal/config"
)

// FromConfig builds quota Options from the application config.
// Falls back to DefaultOptions for any zero values.
func FromConfig(cfg config.Config) Options {
	opts := DefaultOptions()
	if cfg.Quota.MaxScans > 0 {
		opts.MaxScans = cfg.Quota.MaxScans
	}
	if cfg.Quota.WindowSeconds > 0 {
		opts.Window = time.Duration(cfg.Quota.WindowSeconds) * time.Second
	}
	return opts
}
