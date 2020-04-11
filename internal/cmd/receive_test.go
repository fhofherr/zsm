package cmd_test

import (
	"os"
	"testing"

	"github.com/fhofherr/zsm/internal/cmd"
	"github.com/fhofherr/zsm/internal/snapshot"
)

func TestReceive(t *testing.T) {
	tests := []cmd.TestCase{
		{
			Name: "receive snapshot",
			MakeArgs: func(t *testing.T) []string {
				return []string{"receive", "target_fs", "zsm_test@2020-04-10T09:45:58.564585005Z"}
			},
			MakeMSM: func(t *testing.T) *cmd.MockSnapshotManager {
				name := snapshot.MustParseName(t, "zsm_test@2020-04-10T09:45:58.564585005Z")

				sm := &cmd.MockSnapshotManager{}
				sm.On("ReceiveSnapshot", "target_fs", name, os.Stdin).Return(nil)
				return sm
			},
		},
	}

	cmd.RunTests(t, tests)
}
