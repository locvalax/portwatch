package scanner_test

import (
	"net"
	"testing"
	"time"

	"github.com/yourorg/portwatch/internal/scanner"
)

// startTCPServer opens a random local TCP port and returns its port number and a stop func.
func startTCPServer(t *testing.T) (int, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()
	port := ln.Addr().(*net.TCPAddr).Port
	return port, func() { ln.Close() }
}

func TestScan_DetectsOpenPort(t *testing.T) {
	port, stop := startTCPServer(t)
	defer stop()

	opts := scanner.Options{Timeout: 200 * time.Millisecond, Concurrency: 10}
	res, err := scanner.Scan("127.0.0.1", []int{port, port + 1}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.OpenPorts) != 1 || res.OpenPorts[0] != port {
		t.Errorf("expected [%d], got %v", port, res.OpenPorts)
	}
}

func TestScan_NoPorts_ReturnsError(t *testing.T) {
	opts := scanner.DefaultOptions()
	_, err := scanner.Scan("127.0.0.1", []int{}, opts)
	if err == nil {
		t.Error("expected error for empty port list")
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := scanner.DefaultOptions()
	if opts.Timeout <= 0 {
		t.Error("expected positive timeout")
	}
	if opts.Concurrency <= 0 {
		t.Error("expected positive concurrency")
	}
}
