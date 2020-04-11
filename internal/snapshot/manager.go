package snapshot

import (
	"errors"
	"fmt"
	"io"
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
	Receive(string, io.Reader) error
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

// CreateSnapshots creates snapshots of the ZFS file system.
//
// By default CreateSnapshots creates snapshots of all ZFS file systems
// available. This behavior can be modified by passing one or more
// CreateOptions.
func (m *Manager) CreateSnapshots(opts ...CreateOption) error {
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
	nSnapshots := 0
	names := make(map[string][]Name)
	err := m.listSnapshots(func(name Name) {
		names[name.FileSystem] = append(names[name.FileSystem], name)
		nSnapshots++
	})
	if err != nil {
		return fmt.Errorf("clean snapshots: %w", err)
	}

	rejects := make([]Name, 0, nSnapshots)
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

// ListSnapshots returns a list of snapshot names managed by zsm.
func (m *Manager) ListSnapshots() ([]Name, error) {
	var names []Name

	err := m.listSnapshots(func(name Name) {
		names = append(names, name)
	})
	return names, err
}

func (m *Manager) listSnapshots(collect func(Name)) error {
	snapshots, err := m.ZFS.List(zfs.Snapshot)
	if err != nil {
		return fmt.Errorf("list snapshots: %w", err)
	}

	for _, s := range snapshots {
		name, ok := ParseName(s)
		if !ok {
			// snapshot was not created by us
			continue
		}
		collect(name)
	}
	return nil
}

// ReceiveSnapshot receives a snapshot with the passed name.
//
// It writes the data read from r to the snapshot.  ReceiveSnapshot returns an
// error if name.FileSystem does not exist, or if a snapshot with the same name
// already exists.
func (m *Manager) ReceiveSnapshot(targetFS string, name Name, r io.Reader) error {
	allFileSystems, err := m.ZFS.List(zfs.FileSystem)
	if err != nil {
		return fmt.Errorf("receive snapshot: %w", err)
	}
	fsExists := false
	for _, fs := range allFileSystems {
		if fs == targetFS {
			fsExists = true
			break
		}
	}
	if !fsExists {
		return fmt.Errorf("receive snapshot: missing file system: %s", targetFS)
	}

	snExists := false
	err = m.listSnapshots(func(n Name) {
		if name == n {
			snExists = true
		}
	})
	if err != nil {
		return fmt.Errorf("receive snapshot: %w", err)
	}
	if snExists {
		return fmt.Errorf("receive snapshot: exists: %s", name)
	}

	if err := m.ZFS.Receive(name.String(), r); err != nil {
		return fmt.Errorf("receive snapshot: %w", err)
	}
	return nil
}
