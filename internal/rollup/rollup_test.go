package rollup_test

import (
	"sync"
	"testing"
	"time"

	"github.com/user/portwatch/internal/rollup"
	"github.com/user/portwatch/internal/store"
)

func makeDiff(host string, opened []int) rollup.Diff {
	return store.Diff{Host: host, Opened: opened}
}

func TestAggregator_FlushesOnWindowExpiry(t *testing.T) {
	var mu sync.Mutex
	var got []rollup.Diff

	opts := rollup.DefaultOptions()
	opts.Window = 50 * time.Millisecond

	a := rollup.New(opts, func(batch []rollup.Diff) {
		mu.Lock()
		got = append(got, batch...)
		mu.Unlock()
	})
	defer a.Close()

	a.Add(makeDiff("host1", []int{80}))
	a.Add(makeDiff("host2", []int{443}))

	time.Sleep(120 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(got) != 2 {
		t.Fatalf("expected 2 diffs after window, got %d", len(got))
	}
}

func TestAggregator_FlushesOnMaxBatch(t *testing.T) {
	var mu sync.Mutex
	var got []rollup.Diff

	opts := rollup.Options{Window: 10 * time.Second, MaxBatch: 3}
	a := rollup.New(opts, func(batch []rollup.Diff) {
		mu.Lock()
		got = append(got, batch...)
		mu.Unlock()
	})
	defer a.Close()

	a.Add(makeDiff("h1", []int{22}))
	a.Add(makeDiff("h2", []int{80}))
	a.Add(makeDiff("h3", []int{443})) // triggers early flush

	time.Sleep(30 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(got) != 3 {
		t.Fatalf("expected 3 diffs after max-batch flush, got %d", len(got))
	}
}

func TestAggregator_Close_FlushesRemaining(t *testing.T) {
	var mu sync.Mutex
	var got []rollup.Diff

	opts := rollup.Options{Window: 10 * time.Second, MaxBatch: 100}
	a := rollup.New(opts, func(batch []rollup.Diff) {
		mu.Lock()
		got = append(got, batch...)
		mu.Unlock()
	})

	a.Add(makeDiff("host1", []int{8080}))
	a.Close()

	time.Sleep(20 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(got) != 1 {
		t.Fatalf("expected 1 diff after Close, got %d", len(got))
	}
}

func TestAggregator_Add_AfterClose_IsNoOp(t *testing.T) {
	var mu sync.Mutex
	var count int

	opts := rollup.Options{Window: 10 * time.Second, MaxBatch: 100}
	a := rollup.New(opts, func(batch []rollup.Diff) {
		mu.Lock()
		count += len(batch)
		mu.Unlock()
	})

	a.Close()
	a.Add(makeDiff("ghost", []int{9999}))

	time.Sleep(20 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if count != 0 {
		t.Fatalf("expected 0 diffs after Add post-Close, got %d", count)
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := rollup.DefaultOptions()
	if opts.Window <= 0 {
		t.Error("expected positive window")
	}
	if opts.MaxBatch <= 0 {
		t.Error("expected positive max batch")
	}
}
