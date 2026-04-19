package notify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebhook_Send_Success(t *testing.T) {
	var received webhookPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	wh := NewWebhook(ts.URL)
	if err := wh.Send("subject", "body"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Subject != "subject" {
		t.Errorf("expected subject, got %q", received.Subject)
	}
}

func TestWebhook_Send_Non2xx_ReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	wh := NewWebhook(ts.URL)
	if err := wh.Send("s", "b"); err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}

func TestWebhook_Send_BadURL_ReturnsError(t *testing.T) {
	wh := NewWebhook("http://127.0.0.1:0")
	if err := wh.Send("s", "b"); err == nil {
		t.Fatal("expected error for unreachable URL")
	}
}
