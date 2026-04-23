package sampler

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/user/portwatch/internal/scanner"
)

func TestDefaultOptions_RateIsOne(t *testing.T) {
	opts := DefaultOptions()
	if opts.Rate != 1.0 {
		t.Fatalf("expected rate 1.0, got %v", opts.Rate)
	}
}

func TestNew_InvalidRate_ReturnsError(t *testing.T) {
	for _, r := range []float64{-0.1, 1.1, 2.0} {
		_, err := New(Options{Rate: r})
		if err == nil {
			t.Errorf("expected error for rate %v", r)
		}
	}
}

func TestAllow_RateOne_AlwaysTrue(t *testing.T) {
	s, _ := New(Options{Rate: 1.0, Seed: 42})
	for i := 0; i < 20; i++ {
		if !s.Allow("host") {
			t.Fatal("expected Allow to return true for rate 1.0")
		}
	}
}

func TestAllow_RateZero_AlwaysFalse(t *testing.T) {
	s, _ := New(Options{Rate: 0.0, Seed: 42})
	for i := 0; i < 20; i++ {
		if s.Allow("host") {
			t.Fatal("expected Allow to return false for rate 0.0")
		}
	}
}

func TestAllow_PartialRate_Probabilistic(t *testing.T) {
	s, _ := New(Options{Rate: 0.5, Seed: 1})
	allowed := 0
	const n = 1000
	for i := 0; i < n; i++ {
		if s.Allow("h") {
			allowed++
		}
	}
	// Expect roughly 50% ± 10%
	if allowed < 400 || allowed > 600 {
		t.Errorf("expected ~500 allowed, got %d", allowed)
	}
}

func startTCPServer(t *testing.T) (string, int) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()
	addr := ln.Addr().(*net.TCPAddr)
	return "127.0.0.1", addr.Port
}

func TestSampledScanner_Skipped_ReturnsErrSkipped(t *testing.T) {
	s, _ := New(Options{Rate: 0.0})
	sc, _ := scanner.New(scanner.DefaultOptions())
	ss := NewSampledScanner(sc, s)
	_, err := ss.Scan(context.Background(), "127.0.0.1")
	if !errors.Is(err, ErrSkipped) {
		t.Fatalf("expected ErrSkipped, got %v", err)
	}
}

func TestSampledScanner_Allowed_DelegatesToScan(t *testing.T) {
	_, port := startTCPServer(t)
	s, _ := New(Options{Rate: 1.0})
	opts := scanner.DefaultOptions()
	opts.Ports = []int{port}
	sc, _ := scanner.New(opts)
	ss := NewSampledScanner(sc, s)
	ports, err := ss.Scan(context.Background(), "127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) == 0 {
		t.Fatal("expected at least one open port")
	}
}
