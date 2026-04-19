package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(p, []byte(content), 0644))
	return p
}

func TestDefaults(t *testing.T) {
	cfg := Defaults()
	assert.Equal(t, []string{"localhost"}, cfg.Hosts)
	assert.Equal(t, 60, cfg.IntervalSecs)
	assert.Equal(t, "portwatch.db", cfg.StorePath)
	assert.Equal(t, "stdout", cfg.AlertOutput)
	assert.Equal(t, "text", cfg.ReportFormat)
}

func TestLoad_OverridesDefaults(t *testing.T) {
	p := writeTemp(t, `
hosts:
  - 192.168.1.1
  - 10.0.0.2
interval_secs: 30
report_format: json
`)
	cfg, err := Load(p)
	require.NoError(t, err)
	assert.Equal(t, []string{"192.168.1.1", "10.0.0.2"}, cfg.Hosts)
	assert.Equal(t, 30, cfg.IntervalSecs)
	assert.Equal(t, "json", cfg.ReportFormat)
	// defaults preserved
	assert.Equal(t, "portwatch.db", cfg.StorePath)
}

func TestLoad_MissingFile(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.yaml")
	require.NoError(t, err)
	assert.Equal(t, Defaults(), cfg)
}

func TestLoad_FilterSection(t *testing.T) {
	p := writeTemp(t, `
filter:
  include_ports: ["80", "443"]
  exclude_ports: ["22"]
  min_port: 1
  max_port: 9000
`)
	cfg, err := Load(p)
	require.NoError(t, err)
	assert.Equal(t, []string{"80", "443"}, cfg.Filter.IncludePorts)
	assert.Equal(t, []string{"22"}, cfg.Filter.ExcludePorts)
	assert.Equal(t, 9000, cfg.Filter.MaxPort)
}
