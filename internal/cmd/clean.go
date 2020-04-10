package cmd

import (
	"fmt"

	"github.com/fhofherr/zsm/internal/config"
	"github.com/fhofherr/zsm/internal/snapshot"
	"github.com/spf13/cobra"
)

var cleanIntervalFlags = []struct {
	Interval snapshot.Interval
	Key      string
	Default  int
	Short    string
	Long     string
	Help     string
}{
	{
		Interval: snapshot.Minute,
		Key:      config.SnapshotsKeepMinute,
		Default:  config.DefaultSnapshotsKeepMinute,
		Short:    "m",
		Long:     "minute",
		Help:     "Keep the last m minutely snapshots.",
	},
	{
		Interval: snapshot.Hour,
		Key:      config.SnapshotsKeepHour,
		Default:  config.DefaultSnapshotsKeepHour,
		Short:    "H",
		Long:     "hour",
		Help:     "Keep the last H hourly snapshots.",
	},
	{
		Interval: snapshot.Day,
		Key:      config.SnapshotsKeepDay,
		Default:  config.DefaultSnapshotsKeepDay,
		Short:    "d",
		Long:     "day",
		Help:     "Keep the last d daily snapshots.",
	},
	{
		Interval: snapshot.Week,
		Key:      config.SnapshotsKeepWeek,
		Default:  config.DefaultSnapshotsKeepWeek,
		Short:    "w",
		Long:     "week",
		Help:     "Keep the last w weekly snapshots.",
	},
	{
		Interval: snapshot.Month,
		Key:      config.SnapshotsKeepMonth,
		Default:  config.DefaultSnapshotsKeepMonth,
		Short:    "M",
		Long:     "month",
		Help:     "Keep the last M monthly snapshots.",
	},
	{
		Interval: snapshot.Year,
		Key:      config.SnapshotsKeepYear,
		Default:  config.DefaultSnapshotsKeepYear,
		Short:    "y",
		Long:     "year",
		Help:     "Keep the last y yearly snapshots.",
	},
}

func newCleanCommand(cmdCfg *zsmCommandConfig) *cobra.Command {
	cleanCmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean obsolete zfs snapshots created by zsm",
		Long: `Cleans obsolete snapshots created by zsm.

The clean command retains the last m minutely, H hourly, d daily, w weekly,
M monthly, and y yearly snapshots. clean tries to retain as much snapshots as
possible. Thus if the number of available snapshots is less than m zsm will
retain all snapshots that are at least a minute apart.

A single snapshot may be retained for multiple reasons. For example the last
snapshot in the hour zsm clean is called, is retained because it is one of the
last m minutely as well as one of the last H hourly snapshots.

Snapshots that have not been created by zsm, i.e. snapshots that do not fit zsm's
naming conventions, are not removed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			sm, err := cmdCfg.SMFactory(cmdCfg)
			if err != nil {
				return fmt.Errorf("create snapshot manager: %w", err)
			}

			var cfg snapshot.BucketConfig
			for _, iv := range cleanIntervalFlags {
				cfg[iv.Interval] = cmdCfg.V.GetInt(iv.Key)
			}
			return sm.CleanSnapshots(cfg)
		},
	}

	for _, iv := range cleanIntervalFlags {
		cleanCmd.Flags().IntP(iv.Long, iv.Short, iv.Default, iv.Help)
		cmdCfg.V.BindPFlag(iv.Key, cleanCmd.Flags().Lookup(iv.Long))
	}

	return cleanCmd
}
