package watch

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/config"
	"github.com/user/portwatch/internal/scanner"
	"github.com/user/portwatch/internal/store"
)

// RunCmd is the entry point for the "watch" sub-command.
func RunCmd(ctx context.Context, cfgPath string, once bool) error {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if len(cfg.Hosts) == 0 {
		return fmt.Errorf("no hosts configured")
	}

	sc := scanner.New(scanner.Options{
		Ports:   cfg.Ports,
		Timeout: time.Duration(cfg.TimeoutSecs) * time.Second,
		Workers: cfg.Workers,
	})

	st, err := store.New(cfg.StorePath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}

	al := alert.New(alert.Options{Writer: os.Stdout})

	interval := time.Duration(cfg.IntervalSecs) * time.Second
	if interval <= 0 {
		interval = DefaultOptions().Interval
	}

	opts := Options{
		Hosts:    cfg.Hosts,
		Interval: interval,
		Once:     once,
	}

	w := New(opts, sc, st, al)
	return w.Run(ctx)
}
