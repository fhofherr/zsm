package config

import "github.com/spf13/viper"

const (
	// ZFSCmd setting
	ZFSCmd = "zfs.cmd"
	// DefaultZFSCmd is the default value for ZFSCmd
	DefaultZFSCmd = "/sbin/zfs"

	// FileSystemsExclude setting
	FileSystemsExclude = "filesystems.exclude"
)

func setDefaults(v *viper.Viper) {
	v.SetDefault(ZFSCmd, DefaultZFSCmd)
}
