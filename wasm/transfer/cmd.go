package transfer

import "github.com/spf13/cobra"

func TransferCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "send native token to address",
		Run:   transfer,
	}
	cmd.Flags().StringVar(&configPath, "f", "", "the location of transfer config file")
	cmd.Flags().BoolVarP(&fixed, "fixed", "f", false, "if true, transfer to one address; otherwise, transfer to a random address")
	return cmd
}

func TxRpcCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "txrpc",
		Short: "",
		Run:   txrpc,
	}

	cmd.Flags().BoolVarP(&fixed, "fixed", "f", false, "if true, transfer to one address; otherwise, transfer to a random address")
	return cmd
}

func TxRlpEncode() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "txrlpencode",
		Short: "input accounts and reture the rlp encode signtx list",
		Run:   txrlpencode,
	}
	return cmd
}
