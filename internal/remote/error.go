package remote

import "fmt"

// Error represents an error that occurred while calling zsm on a remote host.
type Error struct {
	SubCommand string
	ExitCode   int
	Stderr     string
}

func (e *Error) Error() string {
	return fmt.Sprintf("remote: zsm %s: exit code: %d", e.SubCommand, e.ExitCode)
}
