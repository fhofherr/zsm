package cmd

import (
	"github.com/fhofherr/zsm/internal/build"
	"github.com/spf13/cobra"
)

func newVersionCommand(cmdCfg *zsmCommandConfig) *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print zsm version info",
		RunE: func(cmd *cobra.Command, args []string) error {
			return build.WriteInfo(cmdCfg.Stdout())
		},
	}

	return versionCmd
}
