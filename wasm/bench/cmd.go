package bench

import (
	"github.com/okex/adventure/wasm/bench/case/cw20"
	"github.com/spf13/cobra"
)

func BenchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bench",
		Short: "subcommands are used for benchmarking performance test",
	}

	cw20case := cw20.NewCW20Case()

	cmd.AddCommand(
		cw20case.NewBenchCmd(),
	)
	return cmd
}
