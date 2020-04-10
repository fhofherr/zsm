package snapshot_test

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"

	"github.com/fhofherr/zsm/internal/snapshot"
	"github.com/fhofherr/zsm/internal/zfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestManager_Initialization(t *testing.T) {
	tests := []struct {
		name        string
		mgr         *snapshot.Manager
		callMgr     func(mgr *snapshot.Manager) error
		expectedErr error
	}{
		{
			name: "CreateSnapshot fails on missing ZFSAdapter",
			mgr:  &snapshot.Manager{},
			callMgr: func(mgr *snapshot.Manager) error {
				return mgr.CreateSnapshots()
			},
			expectedErr: errors.New("initialization error: ZFSAdapter nil"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := tt.callMgr(tt.mgr)
			assert.EqualError(t, err, tt.expectedErr.Error())
		})
	}
}

func TestManager_CreateSnapshots(t *testing.T) {
	allFileSystems := []string{
		"zsm_test",
		"zsm_test/fs_1",
		"zsm_test/fs_2",
		"zsm_test/fs_2/nested_fs_1",
	}

	t.Run("create snapshots of all file systems", func(t *testing.T) {
		adapter := &snapshot.MockZFSAdapter{}
		adapter.Test(t)
		adapter.On("List", zfs.FileSystem).Return(allFileSystems, nil)

		for _, fs := range allFileSystems {
			fs := fs
			adapter.On("CreateSnapshot", mock.MatchedBy(func(n interface{}) bool {
				if name, ok := n.(string); ok {
					return snapshot.AssertNameFormat(t, fs, name)
				}
				t.Errorf("%v is not string", n)
				return false
			})).Return(nil)
		}
		mgr := &snapshot.Manager{ZFS: adapter}
		err := mgr.CreateSnapshots()

		assert.NoError(t, err)
		adapter.AssertExpectations(t)
	})

	t.Run("create snapshots of selected file systems", func(t *testing.T) {
		selectedFileSystems := allFileSystems[1:2]

		adapter := &snapshot.MockZFSAdapter{}
		adapter.Test(t)
		adapter.On("List", zfs.FileSystem).Return(allFileSystems, nil)

		opts := make([]snapshot.CreateOption, 0, len(selectedFileSystems))
		for _, fs := range selectedFileSystems {
			fs := fs
			adapter.On("CreateSnapshot", mock.MatchedBy(func(n interface{}) bool {
				if name, ok := n.(string); ok {
					return snapshot.AssertNameFormat(t, fs, name)
				}
				t.Errorf("%v is not string", n)
				return false
			})).Return(nil)
			opts = append(opts, snapshot.FromFileSystem(fs))
		}

		mgr := &snapshot.Manager{ZFS: adapter}
		err := mgr.CreateSnapshots(opts...)

		assert.NoError(t, err)
		adapter.AssertExpectations(t)
	})

	t.Run("don't create snapshot of unknown file system", func(t *testing.T) {
		unknownFileSystem := "some/unknown/file_system"

		adapter := &snapshot.MockZFSAdapter{}
		adapter.Test(t)
		adapter.On("List", zfs.FileSystem).Return(allFileSystems, nil)

		mgr := &snapshot.Manager{ZFS: adapter}
		err := mgr.CreateSnapshots(snapshot.FromFileSystem(unknownFileSystem))

		assert.EqualError(t, err, fmt.Sprintf("unknown filesystem: %q", unknownFileSystem))
	})

	t.Run("ignore excluded file systems", func(t *testing.T) {
		excludedFileSystems := []string{
			"zsm_test/fs_1",
			"zsm_test/fs_2/nested_fs_1",
		}

		adapter := &snapshot.MockZFSAdapter{}
		adapter.Test(t)
		adapter.On("List", zfs.FileSystem).Return(allFileSystems, nil)

		var opts []snapshot.CreateOption
		for _, fs := range allFileSystems {
			if isFileSystemExcluded(excludedFileSystems, fs) {
				// Sometimes the file systems might be specified with a
				// leading slash. This is wrong, but we want to be liberal
				// in what we accept.
				fs = "/" + fs
				opts = append(opts, snapshot.ExcludeFileSystem(fs))
				continue
			}
			fs := fs
			adapter.On("CreateSnapshot", mock.MatchedBy(func(n interface{}) bool {
				if name, ok := n.(string); ok {
					return snapshot.AssertNameFormat(t, fs, name)
				}
				t.Errorf("%v is not string", n)
				return false
			})).Return(nil)
		}

		mgr := &snapshot.Manager{ZFS: adapter}
		err := mgr.CreateSnapshots(opts...)
		assert.NoError(t, err)
		adapter.AssertExpectations(t)
	})
}

func TestManager_CleanSnapshots(t *testing.T) {
	allSnapshots := []string{
		"zfs_test@2020-04-10T09:45:58.564585005Z",
		"zfs_test@2020-04-10T09:44:58.564585005Z",
		"zfs_test@2020-04-10T09:43:58.564585005Z", // outdated according to cfg
		"zfs_test@2020-04-10T08:44:58.564585005Z",
		"zfs_test@2020-04-10T07:44:58.564585005Z", // outdated according to cfg

		"zfs_test/fs_1@2020-04-10T09:45:58.564585005Z",
		"zfs_test/fs_1@2020-04-10T09:44:58.564585005Z",
		"zfs_test/fs_1@2020-04-10T09:43:58.564585005Z", // outdated according to cfg
		"zfs_test/fs_1@2020-04-10T08:44:58.564585005Z",
		"zfs_test/fs_1@2020-04-10T07:44:58.564585005Z", // outdated according to cfg
	}
	// Shuffle allSnapshots to ensure we don't assume anything about the order
	// in CleanSnapshots.
	rand.Shuffle(len(allSnapshots), func(i, j int) {
		allSnapshots[i], allSnapshots[j] = allSnapshots[j], allSnapshots[i]
	})
	cfg := snapshot.BucketConfig{snapshot.Minute: 2, snapshot.Hour: 2}

	adapter := &snapshot.MockZFSAdapter{}
	adapter.Test(t)
	adapter.On("List", zfs.Snapshot).Return(allSnapshots, nil)
	adapter.On("Destroy", "zfs_test@2020-04-10T09:43:58.564585005Z").Return(nil)
	adapter.On("Destroy", "zfs_test@2020-04-10T07:44:58.564585005Z").Return(nil)
	adapter.On("Destroy", "zfs_test/fs_1@2020-04-10T09:43:58.564585005Z").Return(nil)
	adapter.On("Destroy", "zfs_test/fs_1@2020-04-10T07:44:58.564585005Z").Return(nil)

	mgr := &snapshot.Manager{ZFS: adapter}
	err := mgr.CleanSnapshots(cfg)
	assert.NoError(t, err)
	adapter.AssertExpectations(t)
}

func isFileSystemExcluded(excludedFileSystems []string, fs string) bool {
	for _, efs := range excludedFileSystems {
		if fs == efs {
			return true
		}
	}
	return false
}
