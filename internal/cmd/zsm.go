package cmd

import (
	"fmt"
	"io"
	"os"

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
	ListSnapshots() ([]snapshot.Name, error)
	ReceiveSnapshot(string, snapshot.Name, io.Reader) error
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
	smFactory SnapshotManagerFactory
	stdout    io.Writer
	stderr    io.Writer

	V *viper.Viper
}

func (c *zsmCommandConfig) SnapshotManager() (SnapshotManager, error) {
	smFactory := c.smFactory
	if smFactory == nil {
		smFactory = defaultSnapshotManagerFactory
	}
	sm, err := smFactory(c)
	if err != nil {
		return nil, fmt.Errorf("create snapshot manager: %w", err)
	}
	return sm, nil
}

func (c *zsmCommandConfig) Stdout() io.Writer {
	if c.stdout == nil {
		return os.Stdout
	}
	return c.stdout
}

func (c *zsmCommandConfig) Stderr() io.Writer {
	if c.stderr == nil {
		return os.Stderr
	}
	return c.stderr
}

// ZSMCommandOption represents a compile-time option for creating a zsm command.
type ZSMCommandOption func(*zsmCommandConfig)

// WithSnapshotManagerFactory tells NewZSMCommand to use the passed
// SnapshotManagerFactory instead of a default value.
func WithSnapshotManagerFactory(smf SnapshotManagerFactory) ZSMCommandOption {
	return func(o *zsmCommandConfig) {
		o.smFactory = smf
	}
}

// WithStdout sets the standard output used by zsm.
func WithStdout(stdout io.Writer) ZSMCommandOption {
	return func(o *zsmCommandConfig) {
		o.stdout = stdout
	}
}

// WithStderr sets the standard error output used by zsm.
func WithStderr(stderr io.Writer) ZSMCommandOption {
	return func(o *zsmCommandConfig) {
		o.stderr = stderr
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

	rootCmd := newRootCmd(cmdCfg)
	rootCmd.AddCommand(newCreateCommand(cmdCfg))
	rootCmd.AddCommand(newCleanCommand(cmdCfg))
	rootCmd.AddCommand(newListCommand(cmdCfg))
	rootCmd.AddCommand(newReceiveCommand(cmdCfg))
	rootCmd.AddCommand(newSendCommand(cmdCfg))
	rootCmd.AddCommand(newVersionCommand(cmdCfg))

	return rootCmd
}
