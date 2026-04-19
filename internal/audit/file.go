package audit

import (
	"fmt"
	"os"
)

// FileLogger returns an audit Logger that appends to the given file path.
func FileLogger(path string) (*Logger, *os.File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("audit: open file %q: %w", path, err)
	}
	return New(f), f, nil
}
