package bench

import (
	"github.com/okex/adventure/evm/bench/celt"
	multiwmt "github.com/okex/adventure/evm/bench/multi-wmt"
	"github.com/okex/adventure/evm/bench/operate"
	"github.com/okex/adventure/evm/bench/query"
	"github.com/okex/adventure/evm/bench/scenario"
	"github.com/okex/adventure/evm/bench/transfer"
	"github.com/okex/adventure/evm/bench/wmt"
	"github.com/okex/adventure/evm/bench/xen"
	"github.com/okex/adventure/evm/constant"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func BenchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bench",
		Short: "subcommands are used for benchmarking performance test",
	}

	cmd.AddCommand(
		transfer.TransferCmd(),
		operate.OperateCmd(),
		wmt.WMTCmd(),
		query.QueryCmd(),
		multiwmt.MultiWmtCmt(),
		multiwmt.MultiWmtInit(),
		multiwmt.MultiTokenBalance(),
		scenario.ScenarioCmd(),
		transfer.TxRpcCmd(),
		transfer.TxRlpEncode(),
		celt.CeltInit(),
		celt.CeltRun(),
		xen.XenRun(),
	)

	cmd.PersistentFlags().IntP(constant.FlagConcurrency, "c", 1, "The number of fixed goroutines that need to be set")
	viper.BindPFlag(constant.FlagConcurrency, cmd.PersistentFlags().Lookup(constant.FlagConcurrency))
	cmd.PersistentFlags().IntP(constant.FlagSleep, "t", 1000, "The sleep time (ms) when every goroutine needs to wait per round")
	viper.BindPFlag(constant.FlagSleep, cmd.PersistentFlags().Lookup(constant.FlagSleep))
	cmd.PersistentFlags().StringP(constant.FlagPrivateKeyFile, "p", "", "The filepath of private keys")
	viper.BindPFlag(constant.FlagPrivateKeyFile, cmd.PersistentFlags().Lookup(constant.FlagPrivateKeyFile))
	return cmd
}
