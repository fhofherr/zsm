package zfs

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// ListType defines the type of items the caller of List is interested in.
type ListType string

const (
	// FileSystem causes List to list ZFS file systems only.
	FileSystem ListType = "filesystem"

	// Snapshot causes List to list ZFS snapshots only.
	Snapshot ListType = "snapshot"
)

// Adapter wraps the CmdFunc for a zfs executable.
type Adapter CmdFunc

// New creates a new Adapter for the zfs executable located at zfsCmdPath.
func New(zfsCmdPath string) (Adapter, error) {
	s, err := os.Stat(zfsCmdPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("zfs cmd: not found: %s", zfsCmdPath)
		}
		return nil, fmt.Errorf("zfs cmd %s: %w", zfsCmdPath, err)
	}
	if s.Mode()&0111 == 0 {
		return nil, fmt.Errorf("zfs cmd: not executable: %s", zfsCmdPath)
	}
	return Adapter(NewCmdFunc(zfsCmdPath)), nil
}

// List returns a slice containing the names of all available zfs objects of
// the passed type.
//
// List does not make any guarantees about the ordering of the names in the
// returned slice. In fact, callers must expect to be shuffled.
//
// The following caveats for the various supported types apply:
//
// FileSystem
//     List returns the names of all file systems in path notation. In the
//     case of nested file systems, children are prefixed with the name of their
//     parent file system. The topmost file system is usually the zfs pool
//     itself.
//     It is up to the caller to re-construct this hierarchical structure if
//     required.
//
// List returns an error if calling the zfs CmdFunc fails or the output could
// not be parsed.
func (z Adapter) List(typ ListType) ([]string, error) {
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	cmd := z("list", "-H", "-t", string(typ), "-o", "name")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError

		if errors.As(err, &exitErr) {
			return nil, &Error{
				SubCommand: "list",
				ExitCode:   exitErr.ExitCode(),
				Stderr:     stderr.String(),
			}
		}
		return nil, fmt.Errorf("zfs list: %w", err)
	}

	lines := strings.Split(stdout.String(), "\n")
	names := make([]string, 0, len(lines))
	for _, line := range lines {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}
		names = append(names, name)
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("zfs list: %w", ErrNoOutput)
	}

	return names, nil
}

// CreateSnapshot creates a snapshot with the name.
//
// As described in the zfs(8) man page name must be of the format
// filesystem@snapname or volume@snapname. filesystem must be an existing zfs
// filesystem, volume an existing zfs volume. snapname will be the name of the
// snapshot.
func (z Adapter) CreateSnapshot(name string) error {
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	cmd := z("snapshot", name)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError

		if errors.As(err, &exitErr) {
			return &Error{
				SubCommand: "snapshot",
				ExitCode:   exitErr.ExitCode(),
				Stderr:     stderr.String(),
			}
		}
		return fmt.Errorf("zfs snapshot: %w", err)
	}
	return nil
}

// Destroy removes the zfs object with name.
//
// Destroy merely calls zfs destroy. Provided all conditions for destroying an
// object are met, the object will be destroyed.
func (z Adapter) Destroy(name string) error {
	var stderr bytes.Buffer

	cmd := z("destroy", name)
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError

		if errors.As(err, &exitErr) {
			return &Error{
				SubCommand: "destroy",
				ExitCode:   exitErr.ExitCode(),
				Stderr:     stderr.String(),
			}
		}
		return fmt.Errorf("zfs destroy: %w", err)
	}
	return nil
}

// Receive receives a named zfs object from r.
func (z Adapter) Receive(name string, r io.Reader) error {
	var stderr bytes.Buffer

	cmd := z("receive", name)
	cmd.Stderr = &stderr
	cmd.Stdin = r

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError

		if errors.As(err, &exitErr) {
			return &Error{
				SubCommand: "receive",
				ExitCode:   exitErr.ExitCode(),
				Stderr:     stderr.String(),
			}
		}
		return fmt.Errorf("zfs receive: %w", err)
	}
	return nil
}
