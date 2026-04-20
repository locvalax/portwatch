package throttle_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/user/portwatch/internal/throttle"
)

// startTCPServer spins up a local TCP listener and returns its address.
func startTCPServer(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("startTCPServer: %v", err)
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
	return ln.Addr().String()
}

func TestThrottledScanner_FirstScan_Succeeds(t *testing.T) {
	addr := startTCPServer(t)
	_, portStr, _ := net.SplitHostPort(addr)
	var port int
	fmt.Sscanf(portStr, "%d", &port)

	opts := throttle.DefaultOptions()
	th := throttle.New(opts)
	scanner := throttle.NewThrottledScanner(th, nil)

	ctx := context.Background()
	ports, err := scanner.Scan(ctx, "127.0.0.1", port, port)
	if err != nil {
		t.Fatalf("expected no error on first scan, got: %v", err)
	}
	if len(ports) == 0 {
		t.Fatal("expected at least one open port")
	}
}

func TestThrottledScanner_ExceedsBurst_Blocked(t *testing.T) {
	addr := startTCPServer(t)
	_, portStr, _ := net.SplitHostPort(addr)
	var port int
	fmt.Sscanf(portStr, "%d", &port)

	// Burst of 1 so second call is rate-limited.
	opts := throttle.Options{
		Burst:    1,
		Rate:     1,
		Interval: 10 * time.Second,
	}
	th := throttle.New(opts)
	scanner := throttle.NewThrottledScanner(th, nil)

	ctx := context.Background()
	// First call consumes the burst token.
	_, _ = scanner.Scan(ctx, "127.0.0.1", port, port)

	// Second call should be throttled.
	_, err := scanner.Scan(ctx, "127.0.0.1", port, port)
	if err == nil {
		t.Fatal("expected throttle error on second scan, got nil")
	}
}

func TestThrottledScanner_DifferentHosts_Independent(t *testing.T) {
	addr1 := startTCPServer(t)
	addr2 := startTCPServer(t)

	parsePort := func(addr string) int {
		_, portStr, _ := net.SplitHostPort(addr)
		var p int
		fmt.Sscanf(portStr, "%d", &p)
		return p
	}
	port1 := parsePort(addr1)
	port2 := parsePort(addr2)

	opts := throttle.Options{
		Burst:    1,
		Rate:     1,
		Interval: 10 * time.Second,
	}
	th := throttle.New(opts)
	scanner := throttle.NewThrottledScanner(th, nil)

	ctx := context.Background()
	// Exhaust token for host1.
	_, _ = scanner.Scan(ctx, "127.0.0.1", port1, port1)

	// host2 should still have its own token bucket.
	_, err := scanner.Scan(ctx, "127.0.0.2", port2, port2)
	if err != nil {
		t.Fatalf("expected success for different host, got: %v", err)
	}
}
