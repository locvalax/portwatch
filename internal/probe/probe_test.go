package probe

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"
)

func startTCPServer(t *testing.T) (host string, port int) {
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
	addr := ln.Addr().(*net.TCPAddr)
	p, _ := strconv.Atoi(strconv.Itoa(addr.Port))
	return "127.0.0.1", p
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if opts.Timeout <= 0 {
		t.Errorf("expected positive timeout, got %v", opts.Timeout)
	}
	if opts.Concurrency <= 0 {
		t.Errorf("expected positive concurrency, got %d", opts.Concurrency)
	}
}

func TestNew_InvalidConcurrency_ReturnsError(t *testing.T) {
	opts := DefaultOptions()
	opts.Concurrency = 0
	_, err := New(opts)
	if err == nil {
		t.Error("expected error for zero concurrency")
	}
}

func TestProbe_OpenPort_ReturnsOpen(t *testing.T) {
	host, port := startTCPServer(t)
	p, err := New(DefaultOptions())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := p.Probe(context.Background(), host, port)
	if !res.Open {
		t.Errorf("expected open, got closed: %v", res.Err)
	}
	if res.Latency <= 0 {
		t.Errorf("expected positive latency")
	}
}

func TestProbe_ClosedPort_ReturnsClosed(t *testing.T) {
	opts := DefaultOptions()
	opts.Timeout = 200 * time.Millisecond
	p, _ := New(opts)
	res := p.Probe(context.Background(), "127.0.0.1", 1)
	if res.Open {
		t.Error("expected closed port")
	}
	if res.Err == nil {
		t.Error("expected non-nil error for closed port")
	}
}

func TestProbeAll_MixedTargets(t *testing.T) {
	host, port := startTCPServer(t)
	p, _ := New(DefaultOptions())

	targets := []Target{
		{Host: host, Port: port},
		{Host: "127.0.0.1", Port: 1},
	}

	results := p.ProbeAll(context.Background(), targets)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if !results[0].Open {
		t.Errorf("target[0] should be open")
	}
	if results[1].Open {
		t.Errorf("target[1] should be closed")
	}
}

func TestProbeAll_Empty_ReturnsNil(t *testing.T) {
	p, _ := New(DefaultOptions())
	res := p.ProbeAll(context.Background(), nil)
	if res != nil {
		t.Errorf("expected nil for empty targets")
	}
}
