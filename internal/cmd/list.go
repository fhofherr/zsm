package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newListCommand(cmdCfg *zsmCommandConfig) *cobra.Command {
	var outType string

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all snapshots managed by zsm.",
		Long: `List all snapshots managed by zsm.

The --output option allows to switch the output format of list. The currently
supported values are text and jsonl. The jsonl format prints one json document
per line (see http://jsonlines.org/) and is meant for easy programmatic
consumption.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			sm, err := cmdCfg.SnapshotManager()
			if err != nil {
				return err
			}
			names, err := sm.ListSnapshots()
			if err != nil {
				return err
			}

			stdout := cmdCfg.Stdout()
			for _, name := range names {
				switch outType {
				case "text":
					fmt.Fprintln(stdout, name)
				case "jsonl":
					name.ToJSONW(stdout) // nolint: errcheck
				default:
					return fmt.Errorf("unsupported output format: %s", outType)
				}
			}
			return nil
		},
	}

	listCmd.Flags().StringVarP(&outType, "output", "o", "text",
		"Change the output format of list. Supported values: text, jsonl.")

	return listCmd
}
