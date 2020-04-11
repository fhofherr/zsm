package cmd

import (
	"fmt"

	"github.com/fhofherr/zsm/internal/config"
	"github.com/fhofherr/zsm/internal/snapshot"
	"github.com/fhofherr/zsm/internal/zfs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// SnapshotManager represents a type that is capable of managing zfs snapshots.
type SnapshotManager interface {
	CreateSnapshots(...snapshot.CreateOption) error
	CleanSnapshots(snapshot.BucketConfig) error
}

// SnapshotManagerFactory creates a SnapshotManager from SnapshotManagerConfig.
type SnapshotManagerFactory func(*zsmCommandConfig) (SnapshotManager, error)

func defaultSnapshotManagerFactory(cfg *zsmCommandConfig) (SnapshotManager, error) {
	zfsPath := cfg.V.GetString(config.ZFSCmd)
	if zfsPath == "" {
		return nil, fmt.Errorf("--zfs-cmd empty")
	}
	zfsCmd, err := zfs.New(zfsPath)
	if err != nil {
		return nil, fmt.Errorf("default snapshot manager factory: %w", err)
	}
	return &snapshot.Manager{
		ZFS: zfsCmd,
	}, nil
}

type zsmCommandConfig struct {
	SMFactory SnapshotManagerFactory

	V *viper.Viper
}

// ZSMCommandOption represents a compile-time option for creating a zsm command.
type ZSMCommandOption func(*zsmCommandConfig)

// WithSnapshotManagerFactory tells NewZSMCommand to use the passed
// SnapshotManagerFactory instead of a default value.
func WithSnapshotManagerFactory(smf SnapshotManagerFactory) ZSMCommandOption {
	return func(o *zsmCommandConfig) {
		o.SMFactory = smf
	}
}

// NewZSMCommand creates a new zsm command.
//
// If no options are passed a command suitable for production use is created.
func NewZSMCommand(opts ...ZSMCommandOption) *cobra.Command {
	cmdCfg := &zsmCommandConfig{
		V: config.New(),
	}
	for _, opt := range opts {
		opt(cmdCfg)
	}
	if cmdCfg.SMFactory == nil {
		cmdCfg.SMFactory = defaultSnapshotManagerFactory
	}

	rootCmd := newRootCmd(cmdCfg)
	rootCmd.AddCommand(newCreateCommand(cmdCfg))
	rootCmd.AddCommand(newCleanCommand(cmdCfg))

	return rootCmd
}
