package retry

import (
	"time"

	"github.com/user/portwatch/internal/config"
)

// FromConfig builds a Policy from the [retry] section of the app config.
// Missing keys fall back to DefaultPolicy values.
func FromConfig(cfg config.Config) Policy {
	def := DefaultPolicy()

	attempts := cfg.Int("retry", "max_attempts", def.MaxAttempts)
	delayMS := cfg.Int("retry", "delay_ms", int(def.Delay.Milliseconds()))
	mult := cfg.Float("retry", "multiplier", def.Multiplier)

	if attempts < 1 {
		attempts = 1
	}
	if mult <= 0 {
		mult = 1.0
	}

	return Policy{
		MaxAttempts: attempts,
		Delay:       time.Duration(delayMS) * time.Millisecond,
		Multiplier:  mult,
	}
}
