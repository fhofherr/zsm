package cmd

import (
	"github.com/fhofherr/zsm/internal/config"
	"github.com/fhofherr/zsm/internal/snapshot"
	"github.com/spf13/cobra"
)

func newCreateCommand(cmdCfg *zsmCommandConfig) *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create <FILE SYSTEM>",
		Short: "Create snapshots for all ZFS file systems",
		Long: `Creates snapshots for all ZFS file systems, except for those explicitly excluded.

If <FILE SYSTEM> is passed, then a snapshot of only the passed file system is
created unless <FILE SYSTEM> is marked as excluded.

The created snapshots start with the same name as the dataset and are suffixed with @TIMESTAMP
where TIMESTAMP is an RFC3339 timestamp. The time zone of the TIMESTAMP is always UTC regardles
of the system time.`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var createOpts []snapshot.CreateOption

			sm, err := cmdCfg.SnapshotManager()
			if err != nil {
				return err
			}
			if len(args) == 1 {
				createOpts = append(createOpts, snapshot.FromFileSystem(args[0]))
			}
			excludes := cmdCfg.V.GetStringSlice(config.FileSystemsExclude)
			for _, e := range excludes {
				createOpts = append(createOpts, snapshot.ExcludeFileSystem(e))
			}
			return sm.CreateSnapshots(createOpts...)
		},
	}

	createCmd.Flags().StringSliceP("exclude", "e", nil,
		"File systems to exclude when creating a snapshot.")
	cmdCfg.V.BindPFlag(config.FileSystemsExclude, createCmd.Flags().Lookup("exclude"))

	return createCmd
}
