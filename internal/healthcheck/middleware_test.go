package healthcheck_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/user/portwatch/internal/healthcheck"
	"github.com/user/portwatch/internal/scanner"
)

func TestGuardedScanner_ReachableHost_DelegatestoScan(t *testing.T) {
	host, port := startTCPServer(t)

	opts := healthcheck.Options{Timeout: time.Second, ProbePort: port}
	checker := healthcheck.New(opts)

	called := false
	fakeScan := func(_ context.Context, h string, _ scanner.Options) ([]int, error) {
		called = true
		return []int{port}, nil
	}

	gs := healthcheck.NewGuardedScanner(checker, fakeScan)
	ports, err := gs.Scan(context.Background(), host, scanner.DefaultOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected underlying scan to be called")
	}
	if len(ports) == 0 {
		t.Error("expected ports returned")
	}
}

func TestGuardedScanner_UnreachableHost_ReturnsError(t *testing.T) {
	opts := healthcheck.Options{Timeout: 200 * time.Millisecond, ProbePort: 1}
	checker := healthcheck.New(opts)

	called := false
	fakeScan := func(_ context.Context, _ string, _ scanner.Options) ([]int, error) {
		called = true
		return nil, nil
	}

	gs := healthcheck.NewGuardedScanner(checker, fakeScan)
	_, err := gs.Scan(context.Background(), "192.0.2.1", scanner.DefaultOptions())
	if err == nil {
		t.Fatal("expected error for unreachable host")
	}
	if called {
		t.Error("expected underlying scan NOT to be called")
	}
	if !errors.Is(err, err) { // structural check
		t.Errorf("unexpected error type: %v", err)
	}
}
