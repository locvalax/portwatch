package topology

import (
	"fmt"
	"sync"
)

// Node represents a host and its associated metadata.
type Node struct {
	Host  string
	Tags  map[string]string
	Group string
}

// Graph holds the host topology: groupings and peer relationships.
type Graph struct {
	mu    sync.RWMutex
	nodes map[string]*Node
	peers map[string][]string // host -> peer hosts
}

// New returns an empty topology Graph.
func New() *Graph {
	return &Graph{
		nodes: make(map[string]*Node),
		peers: make(map[string][]string),
	}
}

// AddNode registers a host node in the graph.
func (g *Graph) AddNode(n Node) error {
	if n.Host == "" {
		return fmt.Errorf("topology: host must not be empty")
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	copy := n
	g.nodes[n.Host] = &copy
	return nil
}

// AddPeer declares host b as a peer of host a (unidirectional).
func (g *Graph) AddPeer(a, b string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, ok := g.nodes[a]; !ok {
		return fmt.Errorf("topology: unknown host %q", a)
	}
	if _, ok := g.nodes[b]; !ok {
		return fmt.Errorf("topology: unknown host %q", b)
	}
	for _, p := range g.peers[a] {
		if p == b {
			return nil
		}
	}
	g.peers[a] = append(g.peers[a], b)
	return nil
}

// Peers returns all peers registered for host.
func (g *Graph) Peers(host string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make([]string, len(g.peers[host]))
	copy(out, g.peers[host])
	return out
}

// Group returns all hosts belonging to the named group.
func (g *Graph) Group(name string) []*Node {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var out []*Node
	for _, n := range g.nodes {
		if n.Group == name {
			copy := *n
			out = append(out, &copy)
		}
	}
	return out
}

// Node returns the node for host, or an error if unknown.
func (g *Graph) Node(host string) (*Node, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	n, ok := g.nodes[host]
	if !ok {
		return nil, fmt.Errorf("topology: unknown host %q", host)
	}
	copy := *n
	return &copy, nil
}
