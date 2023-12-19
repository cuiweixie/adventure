package script

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/big"
	"os"

	ethcmm "github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
	"github.com/status-im/keycard-go/hexutils"

	"github.com/okex/adventure/common"
	"github.com/okex/adventure/evm/bench/utils"
	"github.com/okex/adventure/evm/config"
)

const SCRIPTION_TEXT = `data:,{"p":"xrc-20","op":"mint","tick":"xone","amt":"10000"}`

var (
	configPath string
	//configurable gasPrice
	gasPrice = new(big.Int).SetUint64(10000000000)
)

func script(cmd *cobra.Command, args []string) {
	amount := big.NewInt(0)

	Hexdata := hexutils.BytesToHex([]byte(SCRIPTION_TEXT))
	data := ethcmm.Hex2Bytes(Hexdata)

	if configPath == "" {
		panic(errors.New("configPath must be not empty "))
	}

	if err := loadConfig(configPath); err != nil {
		panic(err)
	}

	utils.RunTxs(
		utils.DefaultBaseParamFromFlag(),
		func(addr ethcmm.Address) []utils.TxParam {
			return []utils.TxParam{utils.NewTxParam(addr, amount, 300000, gasPrice, data)}
		},
	)
}

func loadConfig(configPath string) error {
	file, err := os.Open(configPath)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	defer file.Close()

	if err := json.Unmarshal(data, &config.TransferCfg); err != nil {
		return err
	}

	privateKeys := common.ReadDataFromFile(config.TransferCfg.AccountsFilePath)
	config.TransferCfg.PrivateKeys = privateKeys
	gasPrice = utils.ParseGasPriceToBigInt(config.TransferCfg.GasPrice, 9)
	return nil
}
