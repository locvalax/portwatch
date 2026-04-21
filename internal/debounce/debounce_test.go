package debounce_test

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/debounce"
	"github.com/user/portwatch/internal/store"
)

func makeDiff(host string, opened []int) store.Diff {
	return store.Diff{Host: host, Opened: opened}
}

func TestDefaultOptions(t *testing.T) {
	opts := debounce.DefaultOptions()
	if opts.Wait <= 0 {
		t.Fatalf("expected positive Wait, got %v", opts.Wait)
	}
}

func TestDebounce_DeliversAfterWait(t *testing.T) {
	d := debounce.New(debounce.Options{Wait: 50 * time.Millisecond})
	d.Add(makeDiff("host1", []int{80}))

	select {
	case diff := <-d.C():
		if diff.Host != "host1" {
			t.Fatalf("unexpected host %q", diff.Host)
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatal("timed out waiting for debounced event")
	}
}

func TestDebounce_CoalescesRapidUpdates(t *testing.T) {
	d := debounce.New(debounce.Options{Wait: 80 * time.Millisecond})

	// Fire three updates in quick succession; only the last should arrive.
	d.Add(makeDiff("host1", []int{80}))
	d.Add(makeDiff("host1", []int{443}))
	d.Add(makeDiff("host1", []int{8080}))

	var got store.Diff
	select {
	case got = <-d.C():
	case <-time.After(400 * time.Millisecond):
		t.Fatal("timed out waiting for debounced event")
	}

	if len(got.Opened) != 1 || got.Opened[0] != 8080 {
		t.Fatalf("expected last diff (port 8080), got %v", got.Opened)
	}

	// Ensure no second event arrives.
	select {
	case extra := <-d.C():
		t.Fatalf("unexpected second event: %v", extra)
	case <-time.After(150 * time.Millisecond):
	}
}

func TestDebounce_DifferentHosts_Independent(t *testing.T) {
	d := debounce.New(debounce.Options{Wait: 40 * time.Millisecond})
	d.Add(makeDiff("hostA", []int{22}))
	d.Add(makeDiff("hostB", []int{3306}))

	seen := map[string]bool{}
	deadline := time.After(300 * time.Millisecond)
	for len(seen) < 2 {
		select {
		case diff := <-d.C():
			seen[diff.Host] = true
		case <-deadline:
			t.Fatalf("only received events for: %v", seen)
		}
	}
}

func TestDebounce_Flush_DeliversImmediately(t *testing.T) {
	d := debounce.New(debounce.Options{Wait: 10 * time.Second})
	d.Add(makeDiff("host1", []int{9200}))
	d.Flush()

	select {
	case diff := <-d.C():
		if diff.Host != "host1" {
			t.Fatalf("unexpected host %q", diff.Host)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Flush did not deliver event immediately")
	}
}
