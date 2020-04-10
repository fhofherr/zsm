package cmd

import (
	"github.com/fhofherr/zsm/internal/config"
	"github.com/spf13/cobra"
)

func newRootCmd(cmdCfg *zsmCommandConfig) *cobra.Command {
	var configFile string

	rootCmd := &cobra.Command{
		Use:          "zsm",
		Short:        "zsm - ZFS snapshot manager",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error

			if configFile != "" {
				err = config.ReadFile(cmdCfg.V, configFile)
			} else {
				err = config.Read(cmdCfg.V)
			}
			return err
		},
	}

	rootCmd.PersistentFlags().
		StringVar(&configFile, "config-file", "", "File containing zsm settings")
	rootCmd.PersistentFlags().
		String("zfs-cmd", config.DefaultZFSCmd, "Full path to zfs executable")
	cmdCfg.V.BindPFlag(config.ZFSCmd, rootCmd.PersistentFlags().Lookup("zfs-cmd"))

	return rootCmd
}
