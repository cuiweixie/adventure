package transfer

import (
	"encoding/json"
	"errors"
	"github.com/okex/adventure/common"
	"github.com/okex/adventure/evm/bench/utils"
	"github.com/okex/adventure/evm/config"
	"github.com/okex/exchain/libs/cosmos-sdk/types"
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

	if configPath == "" {
		panic(errors.New("configPath must be not empty "))
	}

	if err := loadConfig(configPath); err != nil {
		panic(err)
	}

	var toAddrs []types.Address
	if !fixed {
		toAddrs = generateAddress()
	}

	utils.RunTxs(
		utils.DefaultBaseParamFromFlag(),
		func() types.Address {
			return toAddrs[rand.Intn(len(toAddrs))]
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

func generateAddress() []types.Address {
	privateKeys := config.TransferCfg.PrivateKeys
	leng := len(privateKeys)
	addrs := make([]types.Address, 0, leng)
	for i := 0; i < leng; i++ {
		addrs = append(addrs, common.GetCosmosAddressFromPrivateKey(privateKeys[i]))
	}
	return addrs
}
