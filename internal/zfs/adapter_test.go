package zfs_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
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
	tests := []zfs.TestCase{
		{
			Name: "list all file systems",
			Call: func(t *testing.T, a zfs.Adapter) error {
				fileSystems, err := a.List(zfs.FileSystem)
				if err != nil {
					return err
				}
				expected := []string{"zsm_test", "zsm_test/fs_1", "zsm_test/fs_2", "zsm_test/fs_2/nested_fs_1"}
				assert.Equal(t, expected, fileSystems)
				return nil
			},
			ZFSArgs: []string{"list", "-H", "-t", "filesystem", "-o", "name"},
			Stdout: func(t *testing.T) []byte {
				file := filepath.Join("testdata", t.Name(), "zfs_list.out")
				bs, err := ioutil.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				return bs
			},
		},
		{
			Name: "list all snapshots",
			Call: func(t *testing.T, a zfs.Adapter) error {
				snapshots, err := a.List(zfs.Snapshot)
				if err != nil {
					return err
				}
				expected := []string{
					"zsm_test@2020-04-05T09:04:24.01925437Z",
					"zsm_test/fs_1@2020-04-05T09:04:24.01925437Z",
				}
				assert.Equal(t, expected, snapshots)
				return nil
			},
			ZFSArgs: []string{"list", "-H", "-t", "snapshot", "-o", "name"},
			Stdout: func(t *testing.T) []byte {
				file := filepath.Join("testdata", t.Name(), "zfs_list.out")
				bs, err := ioutil.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				return bs
			},
		},
		{
			Name: "list returns no output",
			Call: func(t *testing.T, a zfs.Adapter) error {
				res, err := a.List(zfs.FileSystem)
				if !errors.Is(err, zfs.ErrNoOutput) {
					return err
				}
				assert.Empty(t, res)
				return nil
			},
			ZFSArgs: []string{"list", "-H", "-t", "filesystem", "-o", "name"},
		},
		{
			Name: "list fails",
			Call: func(t *testing.T, a zfs.Adapter) error {
				_, err := a.List(zfs.Snapshot)
				return err
			},
			ZFSArgs: []string{"list", "-H", "-t", "snapshot", "-o", "name"},
			Stderr: func(t *testing.T) []byte {
				return []byte("zfs list wrote this to stderr")
			},
			ZFSExitCode: 10,
		},
	}
	zfs.RunTests(t, tests, true)
}

func TestAdapter_CreateSnapshot(t *testing.T) {
	tests := []zfs.TestCase{
		{
			Name: "create snapshot",
			Call: func(t *testing.T, a zfs.Adapter) error {
				return a.CreateSnapshot("zsm_test/fs_1@snapshot_name")
			},
			ZFSArgs: []string{"snapshot", "zsm_test/fs_1@snapshot_name"},
		},
		{
			Name: "snapshot fails with exit code",
			Call: func(t *testing.T, a zfs.Adapter) error {
				return a.CreateSnapshot("zsm_test/fs_1@snapshot_name")
			},
			ZFSArgs:     []string{"snapshot", "zsm_test/fs_1@snapshot_name"},
			ZFSExitCode: 10,
			Stderr: func(t *testing.T) []byte {
				return []byte("zfs snapshot wrote this to stderr")
			},
		},
	}
	zfs.RunTests(t, tests, true)
}

func TestAdapter_Destroy(t *testing.T) {
	tests := []zfs.TestCase{
		{
			Name: "zfs destroys object",
			Call: func(t *testing.T, a zfs.Adapter) error {
				return a.Destroy("zsm_test@2020-04-10T09:45:58.564585005Z")
			},
			ZFSArgs: []string{"destroy", "zsm_test@2020-04-10T09:45:58.564585005Z"},
		},
		{
			Name: "zfs fails with exit code",
			Call: func(t *testing.T, a zfs.Adapter) error {
				return a.Destroy("zsm_test@2020-04-10T09:45:58.564585005Z")
			},
			ZFSArgs:     []string{"destroy", "zsm_test@2020-04-10T09:45:58.564585005Z"},
			ZFSExitCode: 10,
			Stderr: func(t *testing.T) []byte {
				return []byte("zfs destroy wrote this to stderr")
			},
		},
	}
	zfs.RunTests(t, tests, true)
}

func TestAdapter_Receive(t *testing.T) {
	tests := []zfs.TestCase{
		{
			Name: "pass stdin to zfs receive",
			Call: func(t *testing.T, a zfs.Adapter) error {
				stdin := bytes.NewBuffer([]byte("the caller sent this to zfs"))
				return a.Receive("some-zfs-object", stdin)
			},
			ZFSArgs: []string{"receive", "some-zfs-object"},
			Stdin: func(t *testing.T) []byte {
				return []byte("the caller sent this to zfs")
			},
		},
		{
			Name: "zfs receive fails",
			Call: func(t *testing.T, a zfs.Adapter) error {
				return a.Receive("some-zfs-object", nil)
			},
			ZFSArgs:     []string{"receive", "some-zfs-object"},
			ZFSExitCode: 10,
			Stderr: func(t *testing.T) []byte {
				return []byte("zfs receive wrote this to stderr")
			},
		},
	}
	zfs.RunTests(t, tests, true)
}
