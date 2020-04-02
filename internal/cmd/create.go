package cmd

import (
	"fmt"

	"github.com/fhofherr/zsm/internal/config"
	"github.com/fhofherr/zsm/internal/snapshot"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newCreateCommand(v *viper.Viper, sm *snapshot.Manager) *cobra.Command {
	var exclude []string

	createCmd := &cobra.Command{
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
			for _, e := range exclude {
				opts = append(opts, snapshot.ExcludeFileSystem(e))
			}
			if err := sm.CreateSnapshot(opts...); err != nil {
				return fmt.Errorf("%s: %w", cmd.Name(), err)
			}
			return nil
		},
	}

	createCmd.Flags().StringSliceVarP(&exclude, "exclude", "e", nil,
		"File systems to exclude when creating a snapshot.")
	v.BindPFlag(config.FileSystemsExclude, createCmd.Flags().Lookup("exclude"))
	return createCmd
}
