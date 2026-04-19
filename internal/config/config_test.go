package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/your-org/portwatch/internal/config"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "portwatch-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func TestDefaults(t *testing.T) {
	cfg := config.Defaults()
	if cfg.Interval != 60*time.Second {
		t.Errorf("expected 60s interval, got %v", cfg.Interval)
	}
	if cfg.Alert.Output != "stdout" {
		t.Errorf("expected stdout alert output, got %q", cfg.Alert.Output)
	}
	if len(cfg.Hosts) == 0 {
		t.Error("expected at least one default host")
	}
}

func TestLoad_OverridesDefaults(t *testing.T) {
	yaml := `
hosts:
  - 10.0.0.1
  - 10.0.0.2
ports:
  - 22
  - 3306
interval: 30s
store_path: /tmp/pw.db
alert:
  output: /var/log/portwatch.log
`
	path := writeTemp(t, yaml)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Hosts) != 2 || cfg.Hosts[0] != "10.0.0.1" {
		t.Errorf("hosts not loaded correctly: %v", cfg.Hosts)
	}
	if cfg.Interval != 30*time.Second {
		t.Errorf("expected 30s, got %v", cfg.Interval)
	}
	if cfg.StorePath != "/tmp/pw.db" {
		t.Errorf("unexpected store path: %s", cfg.StorePath)
	}
	if cfg.Alert.Output != "/var/log/portwatch.log" {
		t.Errorf("unexpected alert output: %s", cfg.Alert.Output)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load("/nonexistent/portwatch.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
