package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SlackChannel sends messages to a Slack incoming webhook URL.
type SlackChannel struct {
	WebhookURL string
	client     *http.Client
}

type slackPayload struct {
	Text string `json:"text"`
}

// NewSlack creates a SlackChannel.
func NewSlack(webhookURL string) *SlackChannel {
	return &SlackChannel{
		WebhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// Send posts a formatted message to Slack.
func (s *SlackChannel) Send(subject, body string) error {
	text := fmt.Sprintf("*%s*\n```%s```", subject, body)
	payload := slackPayload{Text: text}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("slack marshal: %w", err)
	}

	resp, err := s.client.Post(s.WebhookURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("slack post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("slack response: %s", resp.Status)
	}
	return nil
}
