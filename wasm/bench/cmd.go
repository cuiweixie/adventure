package bench

import (
	"github.com/okex/adventure/wasm/bench/transfer_cw20"
	"github.com/okex/adventure/wasm/bench/transfer_okt"
	"github.com/spf13/cobra"
)

func BenchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bench",
		Short: "subcommands are used for benchmarking performance test",
	}

	cmd.AddCommand(
		transfer_cw20.TransferCmd(),
		transfer_okt.TransferCmd(),
	)
	return cmd
}
