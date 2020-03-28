package zfs_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

	zfs.Fake()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var swallowedArgs []string

			env := []string{
				fmt.Sprintf("%s=1", zfs.KeyIsFakeZFSCmd),
				fmt.Sprintf("%s=%s", zfs.KeyFakeZFSOutFile, filepath.Join("testdata", t.Name(), "zfs_list.out")),
				fmt.Sprintf("%s=%s", zfs.KeyFakeZFSErrFile, filepath.Join("testdata", t.Name(), "zfs_list.err")),
				fmt.Sprintf("%s=%d", zfs.KeyFakeZFSExitCode, tt.zfsExitCode),
			}
			fakeZFS := zfs.NewCmdFunc(os.Args[0], "-test.run=TestAdapter_List")
			fakeZFS = zfs.WithEnv(fakeZFS, env)
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

	zfs.Fake()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var swallowedArgs []string

			env := []string{
				fmt.Sprintf("%s=1", zfs.KeyIsFakeZFSCmd),
				fmt.Sprintf("%s=%s", zfs.KeyFakeZFSErrFile, filepath.Join("testdata", t.Name(), "zfs_snapshot.err")),
				fmt.Sprintf("%s=%d", zfs.KeyFakeZFSExitCode, tt.zfsExitCode),
			}
			fakeZFS := zfs.NewCmdFunc(os.Args[0], "-test.run=TestAdapter_CreateSnapshot")
			fakeZFS = zfs.WithEnv(fakeZFS, env)
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
