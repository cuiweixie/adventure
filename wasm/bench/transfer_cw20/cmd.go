package transfer_cw20

import "github.com/spf13/cobra"

func TransferCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer-cw20",
		Short: "send cw20 token to address",
		Run:   transfer,
	}
	cmd.Flags().StringVar(&configPath, "f", "", "the location of transfer config file")
	return cmd
}
