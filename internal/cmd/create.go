package cmd

import (
	"fmt"

	"github.com/fhofherr/zsm/internal/snapshot"
	"github.com/spf13/cobra"
)

func newCreateCommand(createSnapshots func(opts ...snapshot.CreateOption) error) *cobra.Command {
	return &cobra.Command{
		Use:   "create [FILE SYSTEM]",
		Short: "Create snapshots for all ZFS file systems",
		Long: `Creates snapshots for all ZFS file systems, except for those explicitly excluded.

If [FILE SYSTEM] is passed, then a snapshot of only the passed file system is
created.`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var opts []snapshot.CreateOption

			if len(args) == 1 {
				opts = append(opts, snapshot.FromFileSystem(args[0]))
			}
			if err := createSnapshots(opts...); err != nil {
				return fmt.Errorf("%s: %w", cmd.Name(), err)
			}
			return nil
		},
	}
}
