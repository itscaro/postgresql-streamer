package cmd

import (
	"github.com/allocine/postgresql-streamer-go/services/postgresql"
	"github.com/spf13/cobra"
)

func createWalDropSlotCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:               "stop",
		Short:             "Kill the active slot PID and drop the current replication slot",
		SilenceUsage:      true,
		SilenceErrors:     true,
		PersistentPreRunE: rootCmdPreRun,
		RunE:              walDropSlotCmdE,
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			printMemory()
		},
	}

	return cmd
}

func walDropSlotCmdE(cmd *cobra.Command, args []string) error {
	if err := postgresql.InitPostgreSQLSettings(settings.DBDsn(), rootCmdOpts.logLevel >= LogLevelDebug); err != nil {
		return err
	}
	return postgresql.CleanSlot(settings.Slot())
}
