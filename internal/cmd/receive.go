package cmd

import (
	"fmt"
	"os"

	"github.com/fhofherr/zsm/internal/snapshot"
	"github.com/spf13/cobra"
)

func newReceiveCommand(cmdCfg *zsmCommandConfig) *cobra.Command {
	receiveCommand := &cobra.Command{
		Use:   "receive <TARGET FILE SYSTEM> <SNAPSHOT>",
		Short: "Receive a snapshot from a remote host.",
		Long: `Receive a snapshot from a remote host.

The <SNAPSHOT> is stored in the passed <TARGET FILE SYSTEM>. Care must be taken
that the target file system is excluded when calling create. Otherwise additional
snapshots of <TARGET FILE SYSTEM> will be created.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			sm, err := cmdCfg.SnapshotManager()
			if err != nil {
				return err
			}
			targetFS := args[0]
			name, ok := snapshot.ParseName(args[1])
			if !ok {
				return fmt.Errorf("invalid snapshot name: %s", args[1])
			}
			return sm.ReceiveSnapshot(targetFS, name, os.Stdin)
		},
	}
	return receiveCommand
}
