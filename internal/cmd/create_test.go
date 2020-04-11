package cmd_test

import (
	"testing"

	"github.com/fhofherr/zsm/internal/cmd"
	"github.com/fhofherr/zsm/internal/snapshot"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	tests := []cmd.TestCase{
		{
			Name: "default args",
			MakeArgs: func(t *testing.T) []string {
				return []string{"create"}
			},
			MakeMSM: func(t *testing.T) *cmd.MockSnapshotManager {
				sm := &cmd.MockSnapshotManager{}
				sm.On("CreateSnapshots").Return(nil)
				return sm
			},
		},
		{
			Name: "specify file system",
			MakeArgs: func(t *testing.T) []string {
				return []string{"create", "zsm_test/fs_1"}
			},
			MakeMSM: func(t *testing.T) *cmd.MockSnapshotManager {
				sm := &cmd.MockSnapshotManager{}
				sm.On("CreateSnapshots", mock.MatchedBy(snapshot.EqualCreateOptions(t,
					snapshot.FromFileSystem("zsm_test/fs_1")),
				)).Return(nil)
				return sm
			},
		},
		{
			Name: "specify excluded file systems",
			MakeArgs: func(t *testing.T) []string {
				return []string{"create", "-e", "zsm_test/fs_1", "--exclude", "zsm_test/fs_2"}
			},
			MakeMSM: func(t *testing.T) *cmd.MockSnapshotManager {
				sm := &cmd.MockSnapshotManager{}
				sm.On("CreateSnapshots", mock.MatchedBy(snapshot.EqualCreateOptions(t,
					snapshot.ExcludeFileSystem("zsm_test/fs_1"),
					snapshot.ExcludeFileSystem("zsm_test/fs_2"),
				))).Return(nil)
				return sm
			},
		},
		{
			Name: "config file",
			MakeArgs: func(t *testing.T) []string {
				cfgFile := cmd.ConfigFile(t, "config.yaml")
				return []string{"--config-file", cfgFile, "create"}
			},
			MakeMSM: func(t *testing.T) *cmd.MockSnapshotManager {
				sm := &cmd.MockSnapshotManager{}
				sm.On("CreateSnapshots", mock.MatchedBy(snapshot.EqualCreateOptions(t,
					snapshot.ExcludeFileSystem("zsm_test/fs_3"),
					snapshot.ExcludeFileSystem("zsm_test/fs_4"),
				))).Return(nil)
				return sm
			},
		},
	}

	cmd.RunTests(t, tests)
}
