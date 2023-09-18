package erc20

import "github.com/spf13/cobra"

func ERC20Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "erc20",
		Short: "send erc20 tx",
		Run:   erc20,
	}
	cmd.Flags().StringVar(&configPath, "f", "", "the location of transfer config file")
	cmd.Flags().StringVar(&contract, "contract", "", "ERC20 contract address")
	return cmd
}
