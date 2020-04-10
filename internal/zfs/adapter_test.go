package zfs_test

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/fhofherr/zsm/internal/zfs"
	"github.com/stretchr/testify/assert"
)

func TestAdapter_New(t *testing.T) {
	tests := []struct {
		name    string
		cmdPath string
		errMsg  string
	}{
		{
			name:    "cmdPath does not exist",
			cmdPath: "path/to/missing/cmd",
			errMsg:  "zfs cmd: not found: path/to/missing/cmd",
		},
		{
			name:    "cmdPath not executable",
			cmdPath: filepath.Join("testdata", t.Name(), "not_executable"),
			errMsg: fmt.Sprintf("zfs cmd: not executable: %s",
				filepath.Join("testdata", t.Name(), "not_executable")),
		},
		{
			name:    "cmdPath executable",
			cmdPath: filepath.Join("testdata", t.Name(), "executable"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			adapter, err := zfs.New(tt.cmdPath)
			if err != nil {
				if tt.errMsg == "" {
					t.Error("No error expected")
				}
				assert.EqualError(t, err, tt.errMsg)
				return
			}
			assert.NoError(t, adapter().Run(), "failed to execute adapter")
		})
	}
}

func TestAdapter_List(t *testing.T) {
	tests := []struct {
		name        string
		typ         zfs.ListType
		zfsExitCode int
		expected    []string
		expectedErr error
	}{
		{
			name:     "list all file systems",
			typ:      zfs.FileSystem,
			expected: []string{"zsm_test", "zsm_test/fs_1", "zsm_test/fs_2", "zsm_test/fs_2/nested_fs_1"},
		},
		{
			name: "list all snapshots",
			typ:  zfs.Snapshot,
			expected: []string{
				"zsm_test@2020-04-05T09:04:24.01925437Z",
				"zsm_test/fs_1@2020-04-05T09:04:24.01925437Z",
			},
		},
		{
			name:        "list fails with exit code",
			typ:         zfs.FileSystem,
			zfsExitCode: 10,
			expectedErr: &zfs.Error{
				SubCommand: "list",
				ExitCode:   10,
				Stderr:     "zfs list wrote this to stderr\n",
			},
		},
		{
			name:        "list returns no output",
			typ:         zfs.FileSystem,
			expectedErr: zfs.ErrNoOutput,
		},
	}

	fakeZFS := zfs.Fake(t)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var swallowedArgs []string

			// Shadow the top-level fakeZFS variable!
			fakeZFS := zfs.WithEnv(fakeZFS, map[string]string{
				zfs.KeyIsFakeZFSCmd:    "1",
				zfs.KeyFakeZFSOutFile:  filepath.Join("testdata", t.Name(), "zfs_list.out"),
				zfs.KeyFakeZFSErrFile:  filepath.Join("testdata", t.Name(), "zfs_list.err"),
				zfs.KeyFakeZFSExitCode: strconv.Itoa(tt.zfsExitCode),
			})
			fakeZFS = zfs.SwallowFurtherArgs(fakeZFS, &swallowedArgs)
			adapter := zfs.Adapter(fakeZFS)

			actual, err := adapter.List(tt.typ)
			if tt.expectedErr == nil && !assert.NoError(t, err) {
				return
			}
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("unexpected error: want: %v; got: %v", tt.expectedErr, err)
			}
			assert.Equal(t, tt.expected, actual)
			assert.Equal(t, []string{"list", "-H", "-t", string(tt.typ), "-o", "name"}, swallowedArgs)
		})
	}
}

func TestAdapter_CreateSnapshot(t *testing.T) {
	tests := []struct {
		name            string
		snapshotName    string
		zfsExitCode     int
		expectedZFSArgs []string
		expectedErr     error
	}{
		{
			name:         "create snapshot",
			snapshotName: "zsm_test/fs_1@snapshot_name",
			expectedZFSArgs: []string{
				"snapshot", "zsm_test/fs_1@snapshot_name",
			},
		},
		{
			name:         "snapshot fails with exit code",
			snapshotName: "zsm_test/fs_1@snapshot_name",
			expectedZFSArgs: []string{
				"snapshot", "zsm_test/fs_1@snapshot_name",
			},
			zfsExitCode: 10,
			expectedErr: &zfs.Error{
				SubCommand: "snapshot",
				ExitCode:   10,
				Stderr:     "zfs snapshot wrote this to stderr\n",
			},
		},
	}

	fakeZFS := zfs.Fake(t)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var swallowedArgs []string

			// Shadow the top-level fakeZFS variable!
			fakeZFS := zfs.WithEnv(fakeZFS, map[string]string{
				zfs.KeyIsFakeZFSCmd:    "1",
				zfs.KeyFakeZFSErrFile:  filepath.Join("testdata", t.Name(), "zfs_snapshot.err"),
				zfs.KeyFakeZFSExitCode: strconv.Itoa(tt.zfsExitCode),
			})
			fakeZFS = zfs.SwallowFurtherArgs(fakeZFS, &swallowedArgs)
			adapter := zfs.Adapter(fakeZFS)

			err := adapter.CreateSnapshot(tt.snapshotName)
			if tt.expectedErr == nil && !assert.NoError(t, err) {
				return
			}
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("unexpected error: want: %v; got: %v", tt.expectedErr, err)
			}
			assert.Equal(t, []string{"snapshot", tt.snapshotName}, swallowedArgs)
		})
	}
}

func TestAdapter_Destroy(t *testing.T) {
	tests := []struct {
		name        string
		objName     string
		zfsExitCode int
		expectedErr error
	}{
		{
			name:    "zfs destroys object",
			objName: "zsm_test@2020-04-10T09:45:58.564585005Z",
		},
		{
			name:        "zfs fails with exit code",
			zfsExitCode: 10,
			expectedErr: &zfs.Error{
				SubCommand: "destroy",
				ExitCode:   10,
				Stderr:     "zfs destroy wrote this to stderr\n",
			},
		},
	}

	fakeZFS := zfs.Fake(t)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var swallowedArgs []string

			// Shadow the top-level fakeZFS variable!
			fakeZFS := zfs.WithEnv(fakeZFS, map[string]string{
				zfs.KeyIsFakeZFSCmd:    "1",
				zfs.KeyFakeZFSErrFile:  filepath.Join("testdata", t.Name(), "zfs_destroy.err"),
				zfs.KeyFakeZFSExitCode: strconv.Itoa(tt.zfsExitCode),
			})
			fakeZFS = zfs.SwallowFurtherArgs(fakeZFS, &swallowedArgs)
			adapter := zfs.Adapter(fakeZFS)

			err := adapter.Destroy(tt.objName)
			if tt.expectedErr == nil && !assert.NoError(t, err) {
				return
			}
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("unexpected error: want: %v; got: %v", tt.expectedErr, err)
			}
			assert.Equal(t, []string{"destroy", tt.objName}, swallowedArgs)
		})
	}
}
