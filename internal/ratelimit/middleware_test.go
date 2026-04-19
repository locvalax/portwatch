package ratelimit_test

import (
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/user/portwatch/internal/ratelimit"
	"github.com/user/portwatch/internal/scanner"
)

func startServer(t *testing.T) (string, int) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
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

func TestGuardedScanner_FirstScan_Succeeds(t *testing.T) {
	host, port := startServer(t)
	opts := scanner.DefaultOptions()
	opts.Ports = []int{port}
	gs := ratelimit.NewGuardedScanner(500*time.Millisecond, opts)
	ports, err := gs.Scan(host)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) == 0 {
		t.Fatal("expected open port")
	}
}

func TestGuardedScanner_SecondScan_RateLimited(t *testing.T) {
	host, port := startServer(t)
	_ = strconv.Itoa(port)
	opts := scanner.DefaultOptions()
	opts.Ports = []int{port}
	gs := ratelimit.NewGuardedScanner(500*time.Millisecond, opts)
	gs.Scan(host)
	_, err := gs.Scan(host)
	if err == nil {
		t.Fatal("expected rate limit error on second scan")
	}
}
