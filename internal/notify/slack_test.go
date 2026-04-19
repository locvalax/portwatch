package notify

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSlack_Send_Success(t *testing.T) {
	var got map[string]string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := NewSlack(ts.URL)
	if err := s.Send("Alert", "port 22 opened"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got["text"], "Alert") {
		t.Errorf("expected subject in text, got %q", got["text"])
	}
	if !strings.Contains(got["text"], "port 22 opened") {
		t.Errorf("expected body in text, got %q", got["text"])
	}
}

func TestSlack_Send_Non2xx_ReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	s := NewSlack(ts.URL)
	if err := s.Send("Alert", "body"); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestSlack_Send_BadURL_ReturnsError(t *testing.T) {
	s := NewSlack("http://127.0.0.1:0/no-server")
	if err := s.Send("Alert", "body"); err == nil {
		t.Fatal("expected error for bad URL")
	}
}
