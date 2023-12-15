package script

import "github.com/spf13/cobra"

func ScriptionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "script",
		Short: "send scription transactions to X1",
		Run:   script,
	}
	cmd.Flags().StringVar(&configPath, "f", "", "the location of transfer config file")
	return cmd
}
