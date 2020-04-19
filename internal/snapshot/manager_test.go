package snapshot_test

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

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

func isFileSystemExcluded(excludedFileSystems []string, fs string) bool {
	for _, efs := range excludedFileSystems {
		if fs == efs {
			return true
		}
	}
	return false
}

func TestManager_ListSnapshots(t *testing.T) {
	tests := []struct {
		name          string
		allSnapshots  []string
		expectedNames []snapshot.Name
	}{
		{
			name: "list all snapshots",
			allSnapshots: []string{
				"zsm_test@2020-04-10T09:45:58.564585005Z",
				"zsm_test@2020-04-10T09:44:58.564585005Z",
			},
			expectedNames: []snapshot.Name{
				snapshot.MustParseName(t, "zsm_test@2020-04-10T09:45:58.564585005Z"),
				snapshot.MustParseName(t, "zsm_test@2020-04-10T09:44:58.564585005Z"),
			},
		},
		{
			name: "filter snapshots created by someone else",
			allSnapshots: []string{
				"zsm_test@2020-04-10T09:45:58.564585005Z",
				"zsm_test@monday",
				"zsm_test@important",
			},
			expectedNames: []snapshot.Name{
				snapshot.MustParseName(t, "zsm_test@2020-04-10T09:45:58.564585005Z"),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			adapter := &snapshot.MockZFSAdapter{}
			adapter.On("List", zfs.Snapshot).Return(tt.allSnapshots, nil)

			sm := &snapshot.Manager{ZFS: adapter}
			names, err := sm.ListSnapshots()
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, tt.expectedNames, names)
		})
	}
}

func TestManager_CleanSnapshots(t *testing.T) {
	allSnapshots := []string{
		"zsm_test@2020-04-10T09:45:58.564585005Z",
		"zsm_test@2020-04-10T09:44:58.564585005Z",
		"zsm_test@2020-04-10T09:43:58.564585005Z", // outdated according to cfg
		"zsm_test@2020-04-10T08:44:58.564585005Z",
		"zsm_test@2020-04-10T07:44:58.564585005Z", // outdated according to cfg

		"zsm_test/fs_1@2020-04-10T09:45:58.564585005Z",
		"zsm_test/fs_1@2020-04-10T09:44:58.564585005Z",
		"zsm_test/fs_1@2020-04-10T09:43:58.564585005Z", // outdated according to cfg
		"zsm_test/fs_1@2020-04-10T08:44:58.564585005Z",
		"zsm_test/fs_1@2020-04-10T07:44:58.564585005Z", // outdated according to cfg
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
	adapter.On("Destroy", "zsm_test@2020-04-10T09:43:58.564585005Z").Return(nil)
	adapter.On("Destroy", "zsm_test@2020-04-10T07:44:58.564585005Z").Return(nil)
	adapter.On("Destroy", "zsm_test/fs_1@2020-04-10T09:43:58.564585005Z").Return(nil)
	adapter.On("Destroy", "zsm_test/fs_1@2020-04-10T07:44:58.564585005Z").Return(nil)

	mgr := &snapshot.Manager{ZFS: adapter}
	err := mgr.CleanSnapshots(cfg)
	assert.NoError(t, err)
	adapter.AssertExpectations(t)
}

func TestManager_ReceiveSnapshot(t *testing.T) {
	var (
		in           bytes.Buffer
		allSnapshots []string

		fileSystems = []string{"target_fs"}
		name        = snapshot.Name{
			FileSystem: "zsm_test",
			Timestamp:  time.Now().UTC(),
		}
	)

	adapter := &snapshot.MockZFSAdapter{}
	adapter.Test(t)
	adapter.On("List", zfs.FileSystem).Return(fileSystems, nil)
	adapter.On("List", zfs.Snapshot).Return(allSnapshots, nil)
	adapter.On("Receive", name.String(), &in).Return(nil)

	sm := &snapshot.Manager{ZFS: adapter}
	err := sm.ReceiveSnapshot(fileSystems[0], name, &in)
	assert.NoError(t, err)
	adapter.AssertExpectations(t)
}

func TestManager_ReceiveSnapshot_Errors(t *testing.T) {
	tests := []struct {
		name         string
		targetFS     string
		snapshot     snapshot.Name
		fileSystems  []string
		allSnapshots []string
		expectedErr  error
	}{
		{
			name:        "error on missing file system",
			targetFS:    "missing_target_fs",
			snapshot:    snapshot.MustParseName(t, "zsm_test@2020-04-10T09:45:58.564585005Z"),
			expectedErr: fmt.Errorf("receive snapshot: missing file system: missing_target_fs"),
		},
		{
			name:         "error on existing snapshot",
			targetFS:     "target_fs",
			snapshot:     snapshot.MustParseName(t, "zsm_test@2020-04-10T09:45:58.564585005Z"),
			fileSystems:  []string{"target_fs"},
			allSnapshots: []string{"zsm_test@2020-04-10T09:45:58.564585005Z"},
			expectedErr:  fmt.Errorf("receive snapshot: exists: zsm_test@2020-04-10T09:45:58.564585005Z"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			adapter := &snapshot.MockZFSAdapter{}
			adapter.Test(t)
			adapter.On("List", zfs.FileSystem).Return(tt.fileSystems, nil)
			adapter.On("List", zfs.Snapshot).Return(tt.allSnapshots, nil)

			mgr := &snapshot.Manager{ZFS: adapter}
			err := mgr.ReceiveSnapshot(tt.targetFS, tt.snapshot, nil)
			if tt.expectedErr == nil && !assert.NoError(t, err) {
				return
			}
			assert.EqualError(t, err, tt.expectedErr.Error())
		})
	}
}

func TestManager_SendSnapshot(t *testing.T) {
	type testCase struct {
		name string
		sn   snapshot.Name
		ref  snapshot.Name
		mock func(*testing.T, *testCase, *snapshot.MockZFSAdapter)
		call func(*testing.T, *testCase, *snapshot.Manager) error
		out  bytes.Buffer
	}
	tests := []testCase{
		{
			name: "snapshot does not exist",
			sn:   snapshot.Name{FileSystem: "missing", Timestamp: time.Now().UTC()},
			mock: func(t *testing.T, tt *testCase, a *snapshot.MockZFSAdapter) {
				a.On("List", zfs.Snapshot).Return([]string(nil), nil)
			},
			call: func(t *testing.T, tt *testCase, sm *snapshot.Manager) error {
				err := sm.SendSnapshot(tt.sn, &tt.out)
				if !assert.EqualError(t, err, fmt.Sprintf("send snapshot: does not exist: %s", tt.sn)) {
					return err
				}
				return nil
			},
		},
		{
			name: "no reference",
			sn:   snapshot.Name{FileSystem: "zsm_test", Timestamp: time.Now().UTC()},
			mock: func(t *testing.T, tt *testCase, a *snapshot.MockZFSAdapter) {
				a.On("List", zfs.Snapshot).Return([]string{tt.sn.String()}, nil)
				a.On("Send", tt.sn.String(), "", &tt.out).Return(nil)
			},
			call: func(t *testing.T, tt *testCase, sm *snapshot.Manager) error {
				return sm.SendSnapshot(tt.sn, &tt.out)
			},
		},
		{
			name: "with reference",
			sn:   snapshot.Name{FileSystem: "zsm_test", Timestamp: time.Now().UTC()},
			ref:  snapshot.Name{FileSystem: "zsm_test", Timestamp: time.Now().UTC().Add(-time.Hour)},
			mock: func(t *testing.T, tt *testCase, a *snapshot.MockZFSAdapter) {
				a.On("List", zfs.Snapshot).Return([]string{tt.sn.String(), tt.ref.String()}, nil)
				a.On("Send", tt.sn.String(), tt.ref.String(), &tt.out).Return(nil)
			},
			call: func(t *testing.T, tt *testCase, sm *snapshot.Manager) error {
				return sm.SendSnapshot(tt.sn, &tt.out, snapshot.Reference(tt.ref))
			},
		},
		{
			name: "reference does not exist",
			sn:   snapshot.Name{FileSystem: "zsm_test", Timestamp: time.Now().UTC()},
			ref:  snapshot.Name{FileSystem: "zsm_test", Timestamp: time.Now().UTC().Add(-time.Hour)},
			mock: func(t *testing.T, tt *testCase, a *snapshot.MockZFSAdapter) {
				a.On("List", zfs.Snapshot).Return([]string{tt.sn.String()}, nil)
			},
			call: func(t *testing.T, tt *testCase, sm *snapshot.Manager) error {
				err := sm.SendSnapshot(tt.sn, &tt.out, snapshot.Reference(tt.ref))
				if !assert.EqualError(t, err, fmt.Sprintf("send snapshot: reference does not exist: %s", tt.sn)) {
					return err
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			adapter := &snapshot.MockZFSAdapter{}
			adapter.Test(t)
			tt.mock(t, &tt, adapter)

			sm := &snapshot.Manager{ZFS: adapter}
			err := tt.call(t, &tt, sm)
			assert.NoError(t, err)
			adapter.AssertExpectations(t)
		})
	}
}
