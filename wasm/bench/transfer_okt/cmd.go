package transfer_okt

import "github.com/spf13/cobra"

func TransferCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer-okt",
		Short: "send native token to address",
		Run:   transfer,
	}
	cmd.Flags().StringVar(&configPath, "f", "", "the location of transfer config file")
	return cmd
}
