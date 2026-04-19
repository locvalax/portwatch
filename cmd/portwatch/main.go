package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user"
	"github./store"
)

func main() {
	cfgPath := flag.String("config", "portwatch.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	s, err := store.New(cfg.StateFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "store: %v\n", err)
		os.Exit(1)
	}

	a, err := alert.New(os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "alert: %v\n", err)
		os.Exit(1)
	}

	opts := scanner.Options{
		Ports:   cfg.Ports,
		Timeout: time.Duration(cfg.TimeoutSecs) * time.Second,
	}

	interval := time.Duration(cfg.IntervalSecs) * time.Second
	r := schedule.New(cfg.Hosts, interval, s, a, opts)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Printf("portwatch started — monitoring %d host(s) every %s\n", len(cfg.Hosts), interval)
	r.Run(ctx)
	fmt.Println("portwatch stopped.")
}
