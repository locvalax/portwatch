package limiter_test

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/user/portwatch/internal/limiter"
	"github.com/user/portwatch/internal/scanner"
)

func startTCPServer(t *testing.T) string {
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
	return ln.Addr().String()
}

func TestDefaultOptions(t *testing.T) {
	opts := limiter.DefaultOptions()
	if opts.MaxConcurrent != 10 {
		t.Fatalf("expected MaxConcurrent=10, got %d", opts.MaxConcurrent)
	}
}

func TestNew_InvalidMax_ReturnsError(t *testing.T) {
	_, err := limiter.New(limiter.Options{MaxConcurrent: 0})
	if err == nil {
		t.Fatal("expected error for MaxConcurrent=0")
	}
}

func TestLimitedScanner_CapsInFlight(t *testing.T) {
	const max = 3
	l, err := limiter.New(limiter.Options{MaxConcurrent: max})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	var peak int32
	var inflight int32
	var wg sync.WaitGroup

	slow := func(ctx context.Context, host string, opts scanner.Options) ([]int, error) {
		cur := atomic.AddInt32(&inflight, 1)
		for {
			p := atomic.LoadInt32(&peak)
			if cur <= p || atomic.CompareAndSwapInt32(&peak, p, cur) {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
		atomic.AddInt32(&inflight, -1)
		return []int{80}, nil
	}

	scanned := limiter.NewLimitedScanner(l, slow)

	for i := 0; i < 9; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = scanned(context.Background(), "127.0.0.1", scanner.DefaultOptions())
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&peak); got > int32(max) {
		t.Fatalf("peak in-flight %d exceeded max %d", got, max)
	}
}

func TestLimitedScanner_ContextCancel_ReturnsError(t *testing.T) {
	l, _ := limiter.New(limiter.Options{MaxConcurrent: 1})

	// Fill the single slot.
	ctxFill := context.Background()
	if err := l.Acquire(ctxFill); err != nil {
		t.Fatalf("acquire: %v", err)
	}
	defer l.Release()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	noop := func(ctx context.Context, host string, opts scanner.Options) ([]int, error) {
		return nil, nil
	}
	scanned := limiter.NewLimitedScanner(l, noop)
	_, err := scanned(ctx, "127.0.0.1", scanner.DefaultOptions())
	if err == nil {
		t.Fatal("expected context deadline error")
	}
}
