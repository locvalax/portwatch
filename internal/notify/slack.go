package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Slack sends notifications to a Slack incoming webhook URL.
type Slack struct {
	webhookURL string
	client     *http.Client
}

// NewSlack creates a new Slack notifier.
func NewSlack(webhookURL string) *Slack {
	return &Slack{
		webhookURL: webhookURL,
		client:     &http.Client{},
	}
}

// Send posts a message to Slack.
func (s *Slack) Send(subject, body string) error {
	payload := map[string]string{
		"text": fmt.Sprintf("*%s*\n%s", subject, body),
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("slack: marshal payload: %w", err)
	}
	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("slack: post: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack: unexpected status %d", resp.StatusCode)
	}
	return nil
}
