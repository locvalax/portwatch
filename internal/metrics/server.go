package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Server exposes collected metrics over HTTP in a lightweight JSON endpoint.
// It is intentionally minimal — no Prometheus dependency, just plain JSON.
type Server struct {
	metrics *Metrics
	server  *http.Server
}

// MetricsSnapshot is the JSON-serialisable view of all collected metrics.
type MetricsSnapshot struct {
	Hosts map[string]HostSnapshot `json:"hosts"`
}

// HostSnapshot holds per-host metric counters.
type HostSnapshot struct {
	Scans    int64     `json:"scans"`
	Alerts   int64     `json:"alerts"`
	Errors   int64     `json:"errors"`
	LastScan time.Time `json:"last_scan,omitempty"`
}

// NewServer creates an HTTP server that serves metrics from m on addr.
// Call ListenAndServe to start accepting connections.
func NewServer(m *Metrics, addr string) *Server {
	s := &Server{metrics: m}

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", s.handleMetrics)
	mux.HandleFunc("/healthz", s.handleHealth)

	s.server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	return s
}

// ListenAndServe starts the HTTP server. It blocks until the server stops.
// Use Shutdown to stop it gracefully.
func (s *Server) ListenAndServe() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully stops the HTTP server, waiting up to 5 seconds for
// in-flight requests to complete.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// handleMetrics writes a JSON snapshot of all host metrics.
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	snapshot := s.buildSnapshot()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(snapshot); err != nil {
		http.Error(w, fmt.Sprintf("encode error: %v", err), http.StatusInternalServerError)
	}
}

// handleHealth returns 200 OK as a simple liveness probe.
func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// buildSnapshot reads current state from the Metrics store and returns a
// serialisable snapshot. The lock inside Metrics guards concurrent access.
func (s *Server) buildSnapshot() MetricsSnapshot {
	s.metrics.mu.RLock()
	defer s.metrics.mu.RUnlock()

	snap := MetricsSnapshot{
		Hosts: make(map[string]HostSnapshot, len(s.metrics.hosts)),
	}
	for host, h := range s.metrics.hosts {
		snap.Hosts[host] = HostSnapshot{
			Scans:    h.scans,
			Alerts:   h.alerts,
			Errors:   h.errors,
			LastScan: h.lastScan,
		}
	}
	return snap
}
