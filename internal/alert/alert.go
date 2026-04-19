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

// Notify formats and writes alert events derived from a store.Diff.
func (n *Notifier) Notify(host string, d store.Diff) []Event {
	var events []Event

	for _, p := range d.Opened {
		e := Event{
			Timestamp: time.Now(),
			Host:      host,
			Level:     LevelAlert,
			Message:   fmt.Sprintf("port %d newly OPEN", p),
		}
		events = append(events, e)
		fmt.Fprintf(n.out, "[%s] %s %s: %s\n", e.Timestamp.Format(time.RFC3339), e.Level, e.Host, e.Message)
	}

	for _, p := range d.Closed {
		e := Event{
			Timestamp: time.Now(),
			Host:      host,
			Level:     LevelWarn,
			Message:   fmt.Sprintf("port %d newly CLOSED", p),
		}
		events = append(events, e)
		fmt.Fprintf(n.out, "[%s] %s %s: %s\n", e.Timestamp.Format(time.RFC3339), e.Level, e.Host, e.Message)
	}

	if len(events) == 0 {
		e := Event{
			Timestamp: time.Now(),
			Host:      host,
			Level:     LevelInfo,
			Message:   "no port changes detected",
		}
		events = append(events, e)
		fmt.Fprintf(n.out, "[%s] %s %s: %s\n", e.Timestamp.Format(time.RFC3339), e.Level, e.Host, e.Message)
	}

	return events
}
