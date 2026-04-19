package notify

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/user/portwatch/internal/store"
)

// Channel represents a notification delivery channel.
type Channel interface {
	Send(subject, body string) error
}

// Notifier dispatches diff notifications to one or more channels.
type Notifier struct {
	channels []Channel
	out      io.Writer
}

// New creates a Notifier. If no channels are provided it falls back to stdout.
func New(channels ...Channel) *Notifier {
	return &Notifier{channels: channels, out: os.Stdout}
}

// Dispatch sends a notification for the given diff. It is a no-op when diff is empty.
func (n *Notifier) Dispatch(host string, diff store.Diff) error {
	if len(diff.Opened) == 0 && len(diff.Closed) == 0 {
		return nil
	}

	subject := fmt.Sprintf("[portwatch] changes detected on %s", host)
	body := buildBody(host, diff)

	if len(n.channels) == 0 {
		_, err := fmt.Fprintln(n.out, subject+"\n"+body)
		return err
	}

	var errs []string
	for _, ch := range n.channels {
		if err := ch.Send(subject, body); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("notify errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

func buildBody(host string, diff store.Diff) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Host: %s\n", host))
	if len(diff.Opened) > 0 {
		sb.WriteString(fmt.Sprintf("  Opened: %v\n", diff.Opened))
	}
	if len(diff.Closed) > 0 {
		sb.WriteString(fmt.Sprintf("  Closed: %v\n", diff.Closed))
	}
	return sb.String()
}
