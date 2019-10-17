package cmd

import "github.com/spf13/cobra"

func createWalCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:               "wal",
		Short:             "",
		SilenceUsage:      true,
		SilenceErrors:     true,
		PersistentPreRunE: rootCmdPreRun,
		RunE:              helpCmd,
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			printMemory()
		},
	}

	cmd.AddCommand(
		createWalDropSlotCmd(),
		createWalParserCmd(),
	)

	return cmd
}
