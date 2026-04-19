package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookChannel posts JSON payloads to an HTTP endpoint.
type WebhookChannel struct {
	URL    string
	client *http.Client
}

type webhookPayload struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// NewWebhook creates a WebhookChannel targeting url.
func NewWebhook(url string) *WebhookChannel {
	return &WebhookChannel{
		URL: url,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send marshals the subject/body as JSON and POSTs to the webhook URL.
func (w *WebhookChannel) Send(subject, body string) error {
	payload := webhookPayload{Subject: subject, Body: body}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook marshal: %w", err)
	}

	resp, err := w.client.Post(w.URL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("webhook post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook response: %s", resp.Status)
	}
	return nil
}
