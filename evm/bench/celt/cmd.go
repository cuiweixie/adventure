package celt

import (
	"context"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/okex/adventure/evm/bench/celt/abi_bin"
	"github.com/spf13/cobra"
	"math/big"
)

var (
	configFile = "./config/celt.json"
	chainID    = new(big.Int).SetUint64(65)
	signer     = types.NewEIP155Signer(chainID)
	gasPrice   = new(big.Int).SetUint64(1000000000)
	gasLimit   = uint64(3000000)
)

func initClient(c *CeltConfig) {
	client, err := ethclient.Dial(c.RPC[0])
	if err != nil {
		panic(err)
	}

	chainID, err = client.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	signer = types.NewEIP155Signer(chainID)
}

func CeltRun() *cobra.Command {
	var wmtCmd = &cobra.Command{
		Use:   "celt",
		Short: "celt run",
		Args:  cobra.NoArgs,
		Run:   wmtRun,
	}
	wmtCmd.Flags().StringVar(&configFile, "f", "", "the location of wmt config file")
	return wmtCmd
}

func CeltInit() *cobra.Command {
	var wmtCmd = &cobra.Command{
		Use:   "celt-init",
		Short: "celt init",
		Args:  cobra.NoArgs,
		Run:   wmtInit,
	}
	wmtCmd.Flags().StringVar(&configFile, "f", "", "the location of wmt config file")
	return wmtCmd
}

func getM() *CeltManager {
	c := loadCeltConfig(configFile)

	abi_bin.InitBuilder()
	initClient(c)
	cList := LoadContractList(c.ContractPath)
	clients := make([]*ethclient.Client, 0)
	for _, v := range c.RPC {
		c, err := ethclient.Dial(v)
		panicerr(err)
		clients = append(clients, c)
	}
	superAcc := keyToAcc(c.SuperAcc)
	return newManager(cList, superAcc, c.WorkerPath, c.ParaNum, clients, c.SendOKTToWorker)
}
func wmtRun(cmd *cobra.Command, args []string) {
	m := getM()
	m.Loop()
}

func wmtInit(cmd *cobra.Command, args []string) {
	m := getM()
	m.Init()
}
