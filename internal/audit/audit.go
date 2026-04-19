package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// Event represents a single audit log entry.
type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Host      string    `json:"host"`
	Action    string    `json:"action"`
	Details   string    `json:"details"`
}

// Logger writes audit events to a destination.
type Logger struct {
	w io.Writer
}

// New returns a Logger writing to w. If w is nil, os.Stdout is used.
func New(w io.Writer) *Logger {
	if w == nil {
		w = os.Stdout
	}
	return &Logger{w: w}
}

// Log writes an audit event.
func (l *Logger) Log(host, action, details string) error {
	e := Event{
		Timestamp: time.Now().UTC(),
		Host:      host,
		Action:    action,
		Details:   details,
	}
	b, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("audit: marshal: %w", err)
	}
	_, err = fmt.Fprintf(l.w, "%s\n", b)
	return err
}

// LogPortChange is a convenience wrapper for port-change events.
func (l *Logger) LogPortChange(host string, opened, closed []uint16) error {
	if len(opened) > 0 {
		if err := l.Log(host, "ports_opened", fmt.Sprintf("%v", opened)); err != nil {
			return err
		}
	}
	if len(closed) > 0 {
		if err := l.Log(host, "ports_closed", fmt.Sprintf("%v", closed)); err != nil {
			return err
		}
	}
	return nil
}
