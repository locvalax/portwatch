package alert

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/user/portwatch/internal/store"
)

// Level represents the severity of an alert.
type Level string

const (
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelAlert Level = "ALERT"
)

// Event describes a single port change alert.
type Event struct {
	Timestamp time.Time
	Host      string
	Level     Level
	Message   string
}

// Notifier writes alert events to an output.
type Notifier struct {
	out io.Writer
}

// New returns a Notifier that writes to w.
// Pass nil to default to os.Stdout.
func New(w io.Writer) *Notifier {
	if w == nil {
		w = os.Stdout
	}
	return &Notifier{out: w}
}

// newEvent creates an Event with the current timestamp and writes it to the Notifier's output.
func (n *Notifier) newEvent(host string, level Level, message string) Event {
	e := Event{
		Timestamp: time.Now(),
		Host:      host,
		Level:     level,
		Message:   message,
	}
	fmt.Fprintf(n.out, "[%s] %s %s: %s\n", e.Timestamp.Format(time.RFC3339), e.Level, e.Host, e.Message)
	return e
}

// Notify formats and writes alert events derived from a store.Diff.
func (n *Notifier) Notify(host string, d store.Diff) []Event {
	var events []Event

	for _, p := range d.Opened {
		events = append(events, n.newEvent(host, LevelAlert, fmt.Sprintf("port %d newly OPEN", p)))
	}

	for _, p := range d.Closed {
		events = append(events, n.newEvent(host, LevelWarn, fmt.Sprintf("port %d newly CLOSED", p)))
	}

	if len(events) == 0 {
		events = append(events, n.newEvent(host, LevelInfo, "no port changes detected"))
	}

	return events
}
