package redact

import "github.com/sgreben/portwatch/internal/config"

// FromConfig builds a Redactor from the application config.
func FromConfig(cfg config.Config) *Redactor {
	opts := DefaultOptions()

	switch cfg.Redact.Mode {
	case "mask":
		opts.Mode = ModeMask
	default:
		opts.Mode = ModeHash
	}

	if cfg.Redact.Salt != "" {
		opts.Salt = cfg.Redact.Salt
	}
	if cfg.Redact.HashLength > 0 {
		opts.Length = cfg.Redact.HashLength
	}

	return New(opts)
}
