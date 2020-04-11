package cmd_test

import (
	"testing"

	"github.com/fhofherr/zsm/internal/cmd"
	"github.com/fhofherr/zsm/internal/config"
	"github.com/fhofherr/zsm/internal/snapshot"
)

func TestClean(t *testing.T) {
	tests := []cmd.TestCase{
		{
			Name: "default buckets",
			MakeArgs: func(_ *testing.T) []string {
				return []string{"clean"}
			},
			MakeMSM: func(t *testing.T) *cmd.MockSnapshotManager {
				cfg := snapshot.BucketConfig{
					snapshot.Minute: config.DefaultSnapshotsKeepMinute,
					snapshot.Hour:   config.DefaultSnapshotsKeepHour,
					snapshot.Day:    config.DefaultSnapshotsKeepDay,
					snapshot.Week:   config.DefaultSnapshotsKeepWeek,
					snapshot.Month:  config.DefaultSnapshotsKeepMonth,
					snapshot.Year:   config.DefaultSnapshotsKeepYear,
				}

				sm := &cmd.MockSnapshotManager{}
				sm.On("CleanSnapshots", cfg).Return(nil)

				return sm
			},
		},
		{
			Name: "command line user-defined buckets",
			MakeArgs: func(t *testing.T) []string {
				return []string{"clean", "-m", "6", "-H", "5", "-d", "4", "-w", "3", "-M", "2", "-y", "1"}
			},
			MakeMSM: func(t *testing.T) *cmd.MockSnapshotManager {
				cfg := snapshot.BucketConfig{
					snapshot.Minute: 6,
					snapshot.Hour:   5,
					snapshot.Day:    4,
					snapshot.Week:   3,
					snapshot.Month:  2,
					snapshot.Year:   1,
				}

				sm := &cmd.MockSnapshotManager{}
				sm.On("CleanSnapshots", cfg).Return(nil)

				return sm
			},
		},
		{
			Name: "config file",
			MakeArgs: func(t *testing.T) []string {
				cfgFile := cmd.ConfigFile(t, "config.yaml")
				return []string{"--config-file", cfgFile, "clean"}
			},
			MakeMSM: func(t *testing.T) *cmd.MockSnapshotManager {
				cfg := snapshot.BucketConfig{
					snapshot.Minute: 1,
					snapshot.Hour:   2,
					snapshot.Day:    3,
					snapshot.Week:   4,
					snapshot.Month:  5,
					snapshot.Year:   6,
				}

				sm := &cmd.MockSnapshotManager{}
				sm.On("CleanSnapshots", cfg).Return(nil)

				return sm
			},
		},
	}

	cmd.RunTests(t, tests)
}
