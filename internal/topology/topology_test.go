package topology

import (
	"testing"
)

func TestAddNode_AndGet(t *testing.T) {
	g := New()
	err := g.AddNode(Node{Host: "web1", Group: "web", Tags: map[string]string{"env": "prod"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n, err := g.Node("web1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Group != "web" {
		t.Errorf("want group=web, got %q", n.Group)
	}
}

func TestAddNode_EmptyHost_ReturnsError(t *testing.T) {
	g := New()
	if err := g.AddNode(Node{}); err == nil {
		t.Fatal("expected error for empty host")
	}
}

func TestNode_Unknown_ReturnsError(t *testing.T) {
	g := New()
	_, err := g.Node("ghost")
	if err == nil {
		t.Fatal("expected error for unknown host")
	}
}

func TestAddPeer_AndQuery(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{Host: "a"})
	_ = g.AddNode(Node{Host: "b"})
	if err := g.AddPeer("a", "b"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	peers := g.Peers("a")
	if len(peers) != 1 || peers[0] != "b" {
		t.Errorf("want [b], got %v", peers)
	}
}

func TestAddPeer_Deduplication(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{Host: "a"})
	_ = g.AddNode(Node{Host: "b"})
	_ = g.AddPeer("a", "b")
	_ = g.AddPeer("a", "b")
	if len(g.Peers("a")) != 1 {
		t.Errorf("expected exactly one peer, got %d", len(g.Peers("a")))
	}
}

func TestAddPeer_UnknownHost_ReturnsError(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{Host: "a"})
	if err := g.AddPeer("a", "missing"); err == nil {
		t.Fatal("expected error for unknown peer")
	}
}

func TestGroup_ReturnsMatchingNodes(t *testing.T) {
	g := New()
	_ = g.AddNode(Node{Host: "web1", Group: "web"})
	_ = g.AddNode(Node{Host: "web2", Group: "web"})
	_ = g.AddNode(Node{Host: "db1", Group: "db"})
	nodes := g.Group("web")
	if len(nodes) != 2 {
		t.Errorf("want 2 web nodes, got %d", len(nodes))
	}
}
