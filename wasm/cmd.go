package evm

import (
	"github.com/okex/adventure/evm/bench"
	"github.com/okex/adventure/evm/constant"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func WasmCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "wasm",
		Short: "wasm cli of test strategy",
	}

	cmd.AddCommand(
		bench.BenchCmd(),
	)

	cmd.PersistentFlags().StringSliceP(constant.FlagIPs, "i", []string{}, "IP list or domain list is accepted, cosmos port or eth port is accepted")
	viper.BindPFlag(constant.FlagIPs, cmd.PersistentFlags().Lookup(constant.FlagIPs))
	return cmd
}
