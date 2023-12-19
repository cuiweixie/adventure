package polybridge

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcmm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"

	"github.com/okex/adventure/common"
	"github.com/okex/adventure/evm/bench/utils"
	"github.com/okex/adventure/evm/config"
)

var (
	// global variables
	eParam          utils.TxParam
	bridgeAmount    = new(big.Int).SetUint64(100000 * 1000000000) //0.0001 Ether = 10^14 Wei = 10^5 GWei
	acc0addr        string
	nilTokenAddress = "0x0000000000000000000000000000000000000000"
	//configurable gasPrice
	gasPrice = new(big.Int).SetUint64(10000000000)
)

func polybridge(cmd *cobra.Command, args []string) {

	if configPath == "" {
		panic(errors.New("configPath must be not empty "))
	}

	if err := loadConfig(configPath); err != nil {
		panic(err)
	}
	acc0pri, err := crypto.HexToECDSA(config.Bridgecfg.PrivateKeys[0])
	if err != nil {
		panic(errors.New("Acc0 private key error"))
	}
	acc0addr = common.GetEthAddressFromPK(acc0pri).String()

	//eParam = utils.NewTxParam(
	//	ethcmm.HexToAddress(config.Bridgecfg.BridgeAddress),
	//	bridgeAmount,
	//	uint64(3000000),
	//	new(big.Int).SetUint64(2500000000),
	//	generateTxData(acc0addr),
	//)

	utils.RunTxsForPoly(
		func(caller ethcmm.Address) []utils.TxParam {
			payload := generateTxData(caller.Hex())
			eParam = utils.NewTxParam(
				ethcmm.HexToAddress(config.Bridgecfg.BridgeAddress),
				bridgeAmount,
				uint64(3000000),
				gasPrice,
				payload,
			)
			return []utils.TxParam{eParam}
		},
	)

}

func generateTxData(toAddr string) []byte {
	erc20ABI, err := abi.JSON(strings.NewReader(BridgeABI))
	if err != nil {
		panic(err)
	}
	//txdata, err := erc20ABI.Pack("bridgeAsset", uint32(1), ethcmm.HexToAddress(config.Bridgecfg.BridgeAddress), bridgeAmount, ethcmm.HexToAddress(toAddr), true, []byte{})
	txdata, err := erc20ABI.Pack("bridgeAsset", uint32(1), ethcmm.HexToAddress(toAddr), bridgeAmount, ethcmm.HexToAddress(nilTokenAddress), true, []byte{})
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

	if err := json.Unmarshal(data, &config.Bridgecfg); err != nil {
		return err
	}

	privateKeys := common.ReadDataFromFile(config.Bridgecfg.AccountsFilePath)
	config.Bridgecfg.PrivateKeys = privateKeys
	gasPrice = utils.ParseGasPriceToBigInt(config.Bridgecfg.GasPrice, 9)

	return nil
}
