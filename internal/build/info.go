package build

import (
	"fmt"
	"io"
	"time"
)

// Build coordinates.
var (
	Version = "dev"
	Commit  = "dev"
	Date    = time.Now().UTC().Format(time.RFC3339)
)

// WriteInfo writes version info about zsm to w.
func WriteInfo(w io.Writer) error {
	_, err := fmt.Fprintf(w, "zsm %s %s (%s)\n", Version, Commit, Date)
	if err != nil {
		return fmt.Errorf("write info: %w", err)
	}
	return nil
}
