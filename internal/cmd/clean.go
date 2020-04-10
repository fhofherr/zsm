package cmd

import (
	"github.com/fhofherr/zsm/internal/config"
	"github.com/fhofherr/zsm/internal/snapshot"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newCleanCommand(v *viper.Viper, sm *snapshot.Manager) *cobra.Command {
	var cfg snapshot.BucketConfig

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
			return sm.CleanSnapshots(cfg)
		},
	}

	cleanCmd.Flags().IntVarP(&cfg[snapshot.Minute], "minute", "m", config.DefaultSnapshotsKeepMinute,
		"Keep the last m minutely snapshots.")
	v.BindPFlag(config.SnapshotsKeepMinute, cleanCmd.Flags().Lookup("minute"))

	cleanCmd.Flags().IntVarP(&cfg[snapshot.Hour], "hour", "H", config.DefaultSnapshotsKeepHour,
		"Keep the last H hourly snapshots.")
	v.BindPFlag(config.SnapshotsKeepHour, cleanCmd.Flags().Lookup("hour"))

	cleanCmd.Flags().IntVarP(&cfg[snapshot.Day], "day", "d", config.DefaultSnapshotsKeepDay,
		"Keep the last d daily snapshots.")
	v.BindPFlag(config.SnapshotsKeepDay, cleanCmd.Flags().Lookup("day"))

	cleanCmd.Flags().IntVarP(&cfg[snapshot.Week], "week", "w", config.DefaultSnapshotsKeepWeek,
		"Keep the last w weekly snapshots.")
	v.BindPFlag(config.SnapshotsKeepWeek, cleanCmd.Flags().Lookup("week"))

	cleanCmd.Flags().IntVarP(&cfg[snapshot.Month], "month", "M", config.DefaultSnapshotsKeepMonth,
		"Keep the last M monthly snapshots.")
	v.BindPFlag(config.SnapshotsKeepMonth, cleanCmd.Flags().Lookup("month"))

	cleanCmd.Flags().IntVarP(&cfg[snapshot.Year], "year", "y", config.DefaultSnapshotsKeepYear,
		"Keep the last y yearly snapshots.")
	v.BindPFlag(config.SnapshotsKeepYear, cleanCmd.Flags().Lookup("year"))

	return cleanCmd
}
