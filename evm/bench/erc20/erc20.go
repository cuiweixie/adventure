package erc20

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcmm "github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/okex/adventure/common"
	"github.com/okex/adventure/evm/bench/utils"
	"github.com/okex/adventure/evm/config"
)

var (
	// used for flags
	contract   string
	configPath string
	//configurable gasPrice
	gasPrice = new(big.Int).SetUint64(1)
	// global variables
	eParam utils.TxParam
)

func erc20(cmd *cobra.Command, args []string) {

	if configPath == "" {
		panic(errors.New("configPath must be not empty "))
	}

	if err := loadConfig(configPath); err != nil {
		panic(err)
	}

	eParam = utils.NewTxParam(
		ethcmm.HexToAddress(contract),
		nil,
		uint64(100000),
		gasPrice,
		generateTxData(),
	)

	utils.RunTxs(
		utils.DefaultBaseParamFromFlag(),
		func(_ ethcmm.Address) []utils.TxParam {
			return []utils.TxParam{eParam}
		},
	)
}

func generateTxData() []byte {
	erc20ABI, err := abi.JSON(strings.NewReader(ERC20ABI))
	if err != nil {
		panic(err)
	}
	txdata, err := erc20ABI.Pack("transfer", ethcmm.HexToAddress("0x2ECF31eCe36ccaC2d3222A303b1409233ECBB225"), new(big.Int).SetInt64(1))
	if err != nil {
		panic(err)
	}
	return txdata
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
