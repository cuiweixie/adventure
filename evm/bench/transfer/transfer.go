package transfer

import (
	"encoding/json"
	"errors"
	ethcmm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/okex/adventure/common"
	"github.com/okex/adventure/evm/bench/utils"
	"github.com/okex/adventure/evm/config"
	evmtypes "github.com/okex/exchain-go-sdk/module/evm/types"
	sdk "github.com/okex/exchain/libs/cosmos-sdk/types"
	"github.com/okex/exchain/libs/tendermint/libs/rand"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

var (
	// used for flags
	fixed      bool
	configPath string
)

func transfer(cmd *cobra.Command, args []string) {
	amount := sdk.MustNewDecFromStr("0.00001").Int
	fixedAddr := ethcmm.BytesToAddress(crypto.Keccak256(rand.Bytes(64)))

	if configPath == "" {
		panic(errors.New("configPath must be not empty "))
	}

	if err := loadConfig(configPath); err != nil {
		panic(err)
	}

	var toAddrs []ethcmm.Address
	if !fixed {
		toAddrs = generateAddress()
	}

	utils.RunTxs(
		utils.DefaultBaseParamFromFlag(),
		func(addr ethcmm.Address) []utils.TxParam {
			to := fixedAddr
			if !fixed {
				to = toAddrs[rand.Intn(len(toAddrs))]
			}
			return []utils.TxParam{utils.NewTxParam(to, amount, 21000, evmtypes.DefaultGasPrice, nil)}
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

	return nil
}

func generateAddress() []ethcmm.Address {
	//privateKeys := constant.PrivateKeys
	//if privateKeyFile := viper.GetString(constant.FlagPrivateKeyFile); privateKeyFile != "" {
	//	privateKeys = common.ReadDataFromFile(privateKeyFile)
	//}

	privateKeys := config.TransferCfg.PrivateKeys

	leng := len(privateKeys)
	addrs := make([]ethcmm.Address, leng, leng)
	for i := 0; i < leng; i++ {
		pk, err := crypto.HexToECDSA(privateKeys[i])
		if err != nil {
			panic(err)
		}
		addrs[i] = common.GetEthAddressFromPK(pk)
	}
	return addrs
}
