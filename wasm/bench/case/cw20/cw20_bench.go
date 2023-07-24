package cw20

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/okex/exchain-go-sdk/types"
	"github.com/okex/exchain/libs/cosmos-sdk/types/errors"
	"github.com/okex/exchain/libs/tendermint/libs/rand"

	"github.com/okex/adventure/common/client"
	cmwraptx "github.com/okex/adventure/common/types"
	"github.com/okex/adventure/wasm/bench/common/account"
	"github.com/okex/adventure/wasm/bench/core"
	"github.com/okex/adventure/wasm/bench/options"
)

type cw20Bench struct {
	core.BaseBench
}

func NewCW20Bench(option *options.CWTransferOption) (*cw20Bench, error) {
	clients, err := client.GenerateCosmosClients(option.TendermintUrls) // generate CosmosClient
	if err != nil {
		return nil, err
	}

	chainId, err := clients[0].QueryChainID()
	if err != nil {
		return nil, errors.Wrap(err, "query chainID failed")
	}

	option.ChainId = chainId

	accounts, err := account.ParseAccountsFromFile(option.PrivateKeysFile)
	if err != nil {
		return nil, err
	}

	// init accounts to set accNumber, nonce
	if err := initAccounts(accounts, clients); err != nil {
		return nil, err
	}
	log.Println("complete init account")

	// use account[0] to deploy contract
	if option.ContractAddress == "" {
		if option.WasmFilePath == "" {
			return nil, fmt.Errorf("should provide wasm file")
		}

		initMsg := fmt.Sprintf(`{
          "name": "USDT",
          "symbol": "USDT",
          "decimals": 9,
          "initial_balances": [
            {
              "address": "%s",
              "amount": "100000000000000000000"
            }
          ],
          "mint": {
            "minter": "%s",
            "cap": "100000000000000000000000000"
          }
        }`, accounts[0].GetHexAddress().String(), accounts[0].GetHexAddress().String())
		addr, err := deployCW20(accounts[0], clients[0], option.WasmFilePath, initMsg)
		if err != nil {
			return nil, errors.Wrap(err, "deploy cw20 contract failed")
		}
		option.ContractAddress = addr
		log.Println("deploy cw20 success: ", addr)

		// reinit account 0
		for accounts[0].Init(clients[0]) != nil {
			time.Sleep(500 * time.Millisecond)
		}
	}

	buildTxFn := func(sender int, accounts []*account.Account, client *client.CosmosClient) (stdTx []*cmwraptx.WrapCMTx) {
		account := accounts[sender]
		defer account.AddNonce()

		execMsg := fmt.Sprintf(`{"transfer": {"amount":"1", "recipient":"%s"}}`, randAddress(accounts))
		tx, err := core.BuildWasmTx(account.GetPrivateKey(), account.GetAccountNumber(), account.GetNonce(), option.ChainId, "", option.ContractAddress, execMsg, (*account).GetHexAddress().String(), "")
		if err != nil {
			panic(err)
		}

		cli := client.Auth().(types.BaseClient)
		bytes, err := cli.GetCodec().MarshalBinaryLengthPrefixed(tx)

		wrapedTx := &cmwraptx.WrapCMTx{
			Tx:    bytes,
			Nonce: account.GetNonce(),
		}

		return []*cmwraptx.WrapCMTx{wrapedTx}
	}

	bench := cw20Bench{
		core.BaseBench{
			Accounts:          accounts,
			Concurrency:       option.ConcurrentNum,
			BuildTxFn:         buildTxFn,
			TendermintClients: clients,
			MempoolThreshold:  option.Threshold,
		},
	}

	return &bench, nil
}

func randAddress(accounts []*account.Account) string {
	acc := accounts[rand.Intn(len(accounts)-1)]
	return acc.GetHexAddress().String()
	//return acc.GetBech32Address().Bech32StringOptimized("ex")
}

func deployCW20(account *account.Account, client *client.CosmosClient, wasmFilePath, initMsg string) (string, error) {
	return client.DeployContract(account.GetPrivateKey(), account.GetAccountNumber(), account.GetNonce(), wasmFilePath, initMsg, *account.GetBech32Address())
}

func initAccounts(accounts []*account.Account, clients []*client.CosmosClient) error {

	wg := &sync.WaitGroup{}

	goroutineNum := len(accounts) / 50
	if goroutineNum == 0 {
		goroutineNum = 1
	}

	if goroutineNum > 1000 {
		goroutineNum = 1000
	}

	gap := len(accounts) / goroutineNum
	remain := len(accounts) % goroutineNum

	wg.Add(goroutineNum)

	for i := 0; i < goroutineNum; i++ {
		go func(gIndex int) {
			defer wg.Done()

			start := gIndex * gap
			end := (gIndex + 1) * gap
			if gIndex == goroutineNum-1 && remain != 0 {
				end += remain
			}

			client := clients[goroutineNum%len(clients)]

			for i := start; i < end; i++ {
				for (accounts)[i].Init(client) != nil {
					time.Sleep(500 * time.Millisecond)
				}
			}
		}(i)
	}

	wg.Wait()
	return nil
}
