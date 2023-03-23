package wasm

import (
	"github.com/okex/adventure/wasm/bench"
	"github.com/spf13/cobra"
)

func WasmCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "wasm",
		Short: "wasm cli of test strategy",
	}

	cmd.AddCommand(
		bench.BenchCmd(),
	)

	return cmd
}
