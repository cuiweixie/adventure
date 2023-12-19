package xen

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"
)

func XenRun() *cobra.Command {
	var xenCmd = &cobra.Command{
		Use:   "xen",
		Short: "xen run",
		Args:  cobra.NoArgs,
		Run:   run,
	}
	xenCmd.Flags().StringVar(&xenConfigFile, "f", "", "the location of celt config file")
	return xenCmd
}

var (
	signer        = types.NewEIP155Signer(new(big.Int).SetInt64(65))
	gasPrice      = new(big.Int).SetInt64(int64(10000000001)) //  // 10gwe
	xenRandom int = 1
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func run(cmd *cobra.Command, args []string) {
	m := getM()

	chainID, err := m.clientList[0].ChainID(context.Background())
	panicerr(err)
	signer = types.NewEIP155Signer(chainID)

	m.CreateAddress()
	m.Loop()
}
