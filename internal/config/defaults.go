package config

import "github.com/spf13/viper"

// ZSM setting keys and default values.
const (
	ZFSCmd        = "zfs.cmd"
	DefaultZFSCmd = "/sbin/zfs"

	SnapshotsCreateExcludeFileSystems = "snapshots.create.exclude_file_systems"
	SnapshotsSendExcludeFileSystems   = "snapshots.send.exclude_file_systems"

	SnapshotsKeepMinute        = "snapshots.keep.minute"
	DefaultSnapshotsKeepMinute = 60

	SnapshotsKeepHour        = "snapshots.keep.hour"
	DefaultSnapshotsKeepHour = 24

	SnapshotsKeepDay        = "snapshots.keep.day"
	DefaultSnapshotsKeepDay = 7

	SnapshotsKeepWeek        = "snapshots.keep.week"
	DefaultSnapshotsKeepWeek = 4

	SnapshotsKeepMonth        = "snapshots.keep.month"
	DefaultSnapshotsKeepMonth = 12

	SnapshotsKeepYear        = "snapshots.keep.year"
	DefaultSnapshotsKeepYear = 5
)

func setDefaults(v *viper.Viper) {
	v.SetDefault(ZFSCmd, DefaultZFSCmd)
}
