package notify

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/user/portwatch/internal/store"
)

type mockChannel struct {
	called  bool
	subject string
	body    string
	err     error
}

func (m *mockChannel) Send(subject, body string) error {
	m.called = true
	m.subject = subject
	m.body = body
	return m.err
}

func TestDispatch_NoDiff_NoOp(t *testing.T) {
	ch := &mockChannel{}
	n := New(ch)
	if err := n.Dispatch("host1", store.Diff{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ch.called {
		t.Fatal("channel should not be called for empty diff")
	}
}

func TestDispatch_WithDiff_CallsChannel(t *testing.T) {
	ch := &mockChannel{}
	n := New(ch)
	diff := store.Diff{Opened: []int{80, 443}}
	if err := n.Dispatch("host1", diff); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ch.called {
		t.Fatal("channel should have been called")
	}
	if !strings.Contains(ch.subject, "host1") {
		t.Errorf("subject missing host: %s", ch.subject)
	}
}

func TestDispatch_ChannelError_ReturnsErr(t *testing.T) {
	ch := &mockChannel{err: errors.New("send failed")}
	n := New(ch)
	diff := store.Diff{Closed: []int{22}}
	if err := n.Dispatch("host2", diff); err == nil {
		t.Fatal("expected error from failing channel")
	}
}

func TestDispatch_NoChannels_WritesToStdout(t *testing.T) {
	var buf bytes.Buffer
	n := New()
	n.out = &buf
	diff := store.Diff{Opened: []int{8080}}
	if err := n.Dispatch("host3", diff); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "host3") {
		t.Errorf("expected host3 in output, got: %s", buf.String())
	}
}
