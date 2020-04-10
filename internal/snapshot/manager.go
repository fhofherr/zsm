package snapshot

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fhofherr/zsm/internal/zfs"
)

// ZFSAdapter represents a type which is capable on performing calls to ZFS
// on the underlying system.
type ZFSAdapter interface {
	CreateSnapshot(string) error
	List(zfs.ListType) ([]string, error)
	Destroy(string) error
}

// CreateOption modifies the way CreateSnapshot creates a snapshot of one
// or more ZFS file systems.
type CreateOption func(*createOpts)

type createOpts struct {
	FileSystems         []string
	ExcludedFileSystems map[string]bool
}

// FromFileSystem makes CreateSnapshot create a snapshot of only the passed
// file system. If FileSystem is passed multiple times to CreateSnapshot it
// creates snapshots of all the passed file systems.
func FromFileSystem(fsName string) CreateOption {
	return func(o *createOpts) {
		o.FileSystems = append(o.FileSystems, fsName)
	}
}

// ExcludeFileSystem marks the passed file System as excluded from creating
// snapshots.
func ExcludeFileSystem(fsName string) CreateOption {
	return func(o *createOpts) {
		if o.ExcludedFileSystems == nil {
			o.ExcludedFileSystems = make(map[string]bool)
		}
		fsName = strings.TrimPrefix(fsName, "/")
		o.ExcludedFileSystems[fsName] = true
	}
}

// Manager manages ZFS snapshots.
type Manager struct {
	ZFS ZFSAdapter
}

// CreateSnapshot creates a snapshot of the ZFS file system.
//
// By default CreateSnapshot creates snapshots of all ZFS file systems
// available. This behavior can be modified by passing one or more
// CreateOptions.
func (m *Manager) CreateSnapshot(opts ...CreateOption) error {
	if m.ZFS == nil {
		return errors.New("initialization error: ZFSAdapter nil")
	}
	snapOpts := &createOpts{}
	for _, opt := range opts {
		opt(snapOpts)
	}

	allFileSystems, err := m.ZFS.List(zfs.FileSystem)
	if err != nil {
		return fmt.Errorf("create snapshot: %w", err)
	}

	selectedFileSystems := snapOpts.FileSystems
	// If no file systems are passed make snapshots of all available file
	// systems.
	if len(selectedFileSystems) == 0 {
		selectedFileSystems = allFileSystems
	}
	if err := selectedFileSystemsKnown(allFileSystems, selectedFileSystems); err != nil {
		return err
	}
	selectedFileSystems = removeExcludedFileSystems(selectedFileSystems, snapOpts.ExcludedFileSystems)

	ts := time.Now().UTC()
	for _, fs := range selectedFileSystems {
		name := Name{FileSystem: fs, Timestamp: ts}
		if err := m.ZFS.CreateSnapshot(name.String()); err != nil {
			return fmt.Errorf("create snapshot: %w", err)
		}
	}
	return nil
}

func selectedFileSystemsKnown(all, selected []string) error {
	fsSet := make(map[string]bool, len(all))
	for _, fs := range all {
		fsSet[fs] = true
	}
	for _, fs := range selected {
		if !fsSet[fs] {
			return fmt.Errorf("unknown filesystem: %q", fs)
		}
	}
	return nil
}

func removeExcludedFileSystems(selected []string, excluded map[string]bool) []string {
	remaining := make([]string, 0, len(selected))
	for _, fs := range selected {
		if excluded[fs] {
			continue
		}
		remaining = append(remaining, fs)
	}
	return remaining
}

// CleanSnapshots removes all snapshots outdated according to BucketConfig.
func (m *Manager) CleanSnapshots(cfg BucketConfig) error {
	snapshots, err := m.ZFS.List(zfs.Snapshot)
	if err != nil {
		return fmt.Errorf("clean snapshots: %w", err)
	}

	names := make(map[string][]Name, len(snapshots))
	for _, s := range snapshots {
		name, ok := ParseName(s)
		if !ok {
			// snapshot was not created by us
			continue
		}
		names[name.FileSystem] = append(names[name.FileSystem], name)
	}

	rejects := make([]Name, 0, len(snapshots))
	for _, ns := range names {
		_, rjs := clean(cfg, ns)
		rejects = append(rejects, rjs...)
	}

	for _, rj := range rejects {
		if err := m.ZFS.Destroy(rj.String()); err != nil {
			return fmt.Errorf("clean snapshots: %w", err)
		}
	}

	return nil
}
