package decay

import (
	"time"

	"github.com/user/portwatch/internal/config"
)

// FromConfig builds Options from the application config.
// Falls back to DefaultOptions for any zero values.
func FromConfig(cfg config.Config) Options {
	opts := DefaultOptions()
	if cfg.Decay.HalfLifeSeconds > 0 {
		opts.HalfLife = time.Duration(cfg.Decay.HalfLifeSeconds) * time.Second
	}
	if cfg.Decay.Floor > 0 {
		opts.Floor = cfg.Decay.Floor
	}
	return opts
}
