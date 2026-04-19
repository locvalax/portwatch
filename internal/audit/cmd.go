package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Tail reads and pretty-prints audit events from a log file to w.
func Tail(path string, w io.Writer) error {
	if w == nil {
		w = os.Stdout
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("audit tail: %w", err)
	}
	if len(data) == 0 {
		fmt.Fprintln(w, "(no audit events)")
		return nil
	}
	lines := splitLines(data)
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		var e Event
		if err := json.Unmarshal(line, &e); err != nil {
			continue
		}
		fmt.Fprintf(w, "[%s] host=%-20s action=%-16s %s\n",
			e.Timestamp.Format("2006-01-02 15:04:05"),
			e.Host, e.Action, e.Details)
	}
	return nil
}

func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}
