package topology

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "topo-*.yaml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	_ = f.Close()
	return f.Name()
}

func TestLoadFile_ParsesNodes(t *testing.T) {
	path := writeTemp(t, `nodes:
  - host: web1
    group: web
    tags: {env: prod}
  - host: db1
    group: db
`)
	g, err := LoadFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n, err := g.Node("web1")
	if err != nil {
		t.Fatalf("node web1 not found: %v", err)
	}
	if n.Tags["env"] != "prod" {
		t.Errorf("want env=prod, got %q", n.Tags["env"])
	}
	if nodes := g.Group("db"); len(nodes) != 1 {
		t.Errorf("want 1 db node, got %d", len(nodes))
	}
}

func TestLoadFile_ParsesPeers(t *testing.T) {
	path := writeTemp(t, `nodes:
  - host: web1
    peers: [db1]
  - host: db1
`)
	g, err := LoadFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	peers := g.Peers("web1")
	if len(peers) != 1 || peers[0] != "db1" {
		t.Errorf("want [db1], got %v", peers)
	}
}

func TestLoadFile_MissingFile(t *testing.T) {
	_, err := LoadFile(filepath.Join(t.TempDir(), "no-such-file.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadFile_InvalidYAML(t *testing.T) {
	path := writeTemp(t, `:::: bad yaml`)
	_, err := LoadFile(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}
