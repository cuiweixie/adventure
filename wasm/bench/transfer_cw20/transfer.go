package transfer_cw20

import (
	"encoding/json"
	"errors"
	"github.com/okex/adventure/common"
	"github.com/okex/adventure/wasm/bench/options"
	"github.com/okex/exchain/libs/cosmos-sdk/types"
	"github.com/okex/exchain/libs/tendermint/libs/rand"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

var (
	transferOption options.TransferOption
	configPath     string
	privateKeys    []string
	accounts       []*EthAccount
)

func transfer(cmd *cobra.Command, args []string) {
	if configPath == "" {
		panic(errors.New("configPath must be not empty "))
	}
	if err := loadConfig(configPath); err != nil {
		panic(err)
	}
	accounts = generateAccounts(privateKeys)

	RunTxs(func() types.AccAddress {
		acc := accounts[rand.Intn(len(accounts)-1)]
		ethAddr := acc.GetHexAddress()
		var bench32Addr types.AccAddress
		bench32Addr = ethAddr[:]
		return bench32Addr
	})
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

	if err := json.Unmarshal(data, &transferOption); err != nil {
		return err
	}

	privateKeys = common.ReadDataFromFile(transferOption.PrivateKeysFile)
	return nil
}
