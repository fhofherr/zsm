package cmd

import (
	"fmt"

	"github.com/fhofherr/zsm/internal/config"
	"github.com/fhofherr/zsm/internal/snapshot"
	"github.com/fhofherr/zsm/internal/zfs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newRootCmd(v *viper.Viper, sm *snapshot.Manager) *cobra.Command {
	var (
		configFile string
		zfsCmd     string
	)

	rootCmd := &cobra.Command{
		Use:          "zsm",
		Short:        "zsm - ZFS snapshot manager",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			if configFile != "" {
				err = config.ReadFile(v, configFile)
			} else {
				err = config.Read(v)
			}
			if err != nil {
				return
			}

			if zfsCmd == "" {
				err = fmt.Errorf("--zfs-cmd empty")
				return
			}
			sm.ZFS, err = zfs.New(zfsCmd)
			return
		},
	}

	rootCmd.PersistentFlags().
		StringVar(&configFile, "config-file", "", "File containing zsm settings")
	rootCmd.PersistentFlags().
		StringVar(&zfsCmd, "zfs-cmd", config.DefaultZFSCmd, "Full path to zfs executable")
	v.BindPFlag(config.ZFSCmd, rootCmd.Flags().Lookup("zfs-cmd"))

	return rootCmd
}

// Execute executes the zsm command.
func Execute() error {
	v := config.New()
	sm := &snapshot.Manager{}

	rootCmd := newRootCmd(v, sm)
	rootCmd.AddCommand(newCreateCommand(v, sm))
	rootCmd.AddCommand(newCleanCommand(v, sm))

	return rootCmd.Execute()
}
