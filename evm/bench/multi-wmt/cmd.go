package multiwmt

import (
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
	"math/big"
	"strconv"
)

var (
	wmtFile  = "./config/wmt.json"
	chainID  = new(big.Int).SetUint64(65)
	signer   = types.NewEIP155Signer(chainID)
	gasPrice = new(big.Int).SetUint64(1000000000)
	gasLimit = uint64(3000000)
)

func MultiWmtCmt() *cobra.Command {
	var wmtCmd = &cobra.Command{
		Use:   "multiwmt",
		Short: "wmt-run",
		Args:  cobra.NoArgs,
		Run:   wmtRun,
	}
	wmtCmd.Flags().StringVar(&wmtFile, "f", "", "the location of wmt config file")
	return wmtCmd
}

func MultiWmtInit() *cobra.Command {
	var wmtCmd = &cobra.Command{
		Use:   "multiwmt-init",
		Short: "wmt-init",
		Args:  cobra.NoArgs,
		Run:   wmtInit,
	}
	wmtCmd.Flags().StringVar(&wmtFile, "f", "", "the location of wmt config file")
	return wmtCmd
}

func MultiWmtTransferGas() *cobra.Command {
	var wmtCmd = &cobra.Command{
		Use:   "multiwmt-gas",
		Short: "wmt-gas",
		Args:  cobra.NoArgs,
		Run:   wmtTransferGas,
	}
	wmtCmd.Flags().StringVar(&wmtFile, "f", "", "the location of wmt config file")
	return wmtCmd
}

func MultiTokenBalance() *cobra.Command {
	var wmtCmd = &cobra.Command{
		Use:   "multiwmt-token",
		Short: "wmt-token",
		Args:  cobra.NoArgs,
		Run:   wmtToken,
	}
	wmtCmd.Flags().StringVar(&wmtFile, "f", "", "the location of wmt config file")
	return wmtCmd
}

func getM() *wmtManager {
	c := loadWMTConfig(wmtFile)

	initBuilder()
	initClient(c)
	cList := LoadContractList(c.ContractPath)
	clients := make([]*okcClient, 0)
	for i, v := range c.RPC {
		client, err := ethclient.Dial(v)
		panicerr(err)
		clients = append(clients, &okcClient{
			Client: client,
			rpc:    c.Node[i],
		})
	}
	superAcc := keyToAcc(c.SuperAcc)
	return newManager(cList, superAcc, c.WorkerPath, c.ParaNum, clients, c.SendOKTToWorker, c.Threshold)
}
func wmtRun(cmd *cobra.Command, args []string) {
	m := getM()
	m.Loop()
}

func wmtInit(cmd *cobra.Command, args []string) {
	m := getM()
	m.TransferToken0ToAccount()
}

func wmtTransferGas(cmd *cobra.Command, args []string) {
	m := getM()

	amount := int64(1)
	if len(args) > 0 {
		fmt.Println(args[1])
		num, _ := strconv.Atoi(args[1])
		amount = int64(num)
	}
	m.TransferGas(amount)
}

func wmtToken(cmd *cobra.Command, args []string) {
	m := getM()
	m.DisPlayToken()
}
