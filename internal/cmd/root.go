package cmd

import (
	"fmt"

	"github.com/fhofherr/zsm/internal/snapshot"
	"github.com/fhofherr/zsm/internal/zfs"
	"github.com/spf13/cobra"
)

func newRootCmd(sm *snapshot.Manager) *cobra.Command {
	var zfsCmd string

	rootCmd := &cobra.Command{
		Use:          "zsm",
		Short:        "zsm - ZFS snapshot manager",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			if zfsCmd == "" {
				err = fmt.Errorf("--zfs-cmd empty")
				return
			}
			sm.ZFS, err = zfs.New(zfsCmd)
			return
		},
	}

	rootCmd.PersistentFlags().
		StringVar(&zfsCmd, "zfs-cmd", "/sbin/zfs", "Full path to zfs executable")

	return rootCmd
}

// Execute executes the zsm command.
func Execute() error {
	sm := &snapshot.Manager{}

	rootCmd := newRootCmd(sm)
	rootCmd.AddCommand(newCreateCommand(sm.CreateSnapshot))

	return rootCmd.Execute()
}
