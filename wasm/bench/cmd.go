package bench

import (
	"github.com/spf13/cobra"

	"github.com/okex/adventure/wasm/bench/case/cw20"
	"github.com/okex/adventure/wasm/bench/case/cwokt"
	"github.com/okex/adventure/wasm/bench/case/cwoperate"
)

func BenchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bench",
		Short: "subcommands are used for benchmarking performance test",
	}

	cw20Cmd := cw20.NewCW20Cmd()
	oktCmd := cwokt.NewCW20Cmd()
	operateCmd := cwoperate.NewCWOperateCmd()

	cmd.AddCommand(
		cw20Cmd.NewBenchCmd(),
		cw20Cmd.NewConfigCmd(),
		oktCmd.NewBenchCmd(),
		oktCmd.NewConfigCmd(),
		operateCmd.NewBenchCmd(),
		operateCmd.NewConfigCmd(),
	)
	return cmd
}
