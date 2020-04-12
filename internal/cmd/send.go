package cmd

import (
	"github.com/fhofherr/zsm/internal/config"
	"github.com/spf13/cobra"
)

func newSendCommand(cmdCfg *zsmCommandConfig) *cobra.Command {
	sendCmd := &cobra.Command{
		Use:   "send <DESTINATION> <TARGET_FS> [SOURCE_FS]",
		Short: "Send snapshots to the <TARGET_FS> at <DESTINATION>.",
		Long: `Send snapshots to the <TARGET_FS> at <DESTINATION>.

<DESTINATION> must be of the form <USER>@<HOST>[:PORT]. A SSH server must listen
on HOST at PORT and USER must be allowed to log in using key-based authentication.
Additionally USER must be allowed to execute zsm receive on HOST.

If <DESTINATION> has no snapshot for a source file system, send transmits all
available snapshots. Otherwise send transmits only snapshots which are newer than
the last available snapshot on <DESTINATION>. send does not perform any kind of
clean up on the <DESTINATION>. The administrators of <DESTINATION> are
responsible for that.

If [SOURCE_FS] is specified only snapshots from [SOURCE_FS] will be transmitted.`,
		Args: cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			panic("implement me")
		},
	}
	sendCmd.Flags().StringSliceP("exclude", "e", nil,
		"File systems to exclude when sending snapshots.")
	cmdCfg.V.BindPFlag(config.SnapshotsSendExcludeFileSystems, sendCmd.Flags().Lookup("exclude"))

	return sendCmd
}
