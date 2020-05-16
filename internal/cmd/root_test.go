package cmd_test

import (
	"testing"

	"github.com/fhofherr/zsm/internal/cmd"
	"github.com/fhofherr/zsm/internal/snapshot"
	"github.com/stretchr/testify/assert"
)

func TestRootCommand(t *testing.T) {
	tests := []cmd.TestCase{
		{
			Name: "set zfs command",
			MakeArgs: func(t *testing.T) []string {
				return []string{"--zfs-cmd", "path/to/zfs", "create"}
			},
			MakeMSM: func(t *testing.T) *snapshot.MockManager {
				sm := &snapshot.MockManager{}
				sm.On("CreateSnapshots").Return(nil)
				return sm
			},
			AssertMSM: func(t *testing.T, msm *snapshot.MockManager) {
				assert.Equal(t, "path/to/zfs", msm.ZFS)
			},
		},
		{
			Name: "set zfs command config file",
			MakeArgs: func(t *testing.T) []string {
				cfgFile := cmd.ConfigFile(t, "config.yaml")
				return []string{"--config-file", cfgFile, "create"}
			},
			MakeMSM: func(t *testing.T) *snapshot.MockManager {
				sm := &snapshot.MockManager{}
				sm.On("CreateSnapshots").Return(nil)
				return sm
			},
			AssertMSM: func(t *testing.T, msm *snapshot.MockManager) {
				assert.Equal(t, "another/path/to/zfs", msm.ZFS)
			},
		},
	}
	cmd.RunTests(t, tests)
}
