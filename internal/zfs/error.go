package zfs

import (
	"errors"
	"fmt"
)

// ErrNoOutput signals that zfs did not write any output to stdout.
var ErrNoOutput = errors.New("no output")

// Error represents errors that occurred while executing zfs.
type Error struct {
	SubCommand string
	ExitCode   int
	Stderr     string
}

func (e *Error) Error() string {
	return fmt.Sprintf("zfs %s: exit code: %d", e.SubCommand, e.ExitCode)
}

// Is returns true if target is an *Error and is equal to e.
func (e *Error) Is(target error) bool {
	var zfsErr *Error

	if !errors.As(target, &zfsErr) {
		return false
	}
	return *e == *zfsErr
}
