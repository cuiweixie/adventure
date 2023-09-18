package erc20

import "github.com/spf13/cobra"

const (
	FlagPriavteKey  = "private-key"
	FlagAddressFile = "address-file"
)

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

func ERC20InitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "erc20-init",
		Short: "deploy ERC20 and batch-transfer",
		Args:  cobra.ExactArgs(1),
		Run:   erc20init,
	}
	cmd.Flags().StringVarP(&privateKey, FlagPriavteKey, "s", "", "its private key should be imported as a rich account")
	cmd.Flags().StringVarP(&addressFile, FlagAddressFile, "a", "", "the path of ethereum-format address file")
	return cmd
}
