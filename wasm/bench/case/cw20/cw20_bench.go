package cw20

import (
	"fmt"
	"github.com/okex/adventure/common/client"
	"github.com/okex/adventure/wasm/bench/common/account"
	"github.com/okex/adventure/wasm/bench/core"
	"github.com/okex/adventure/wasm/bench/options"
	"github.com/okex/exchain/libs/cosmos-sdk/types/errors"
	"github.com/okex/exchain/libs/cosmos-sdk/x/auth"
	"github.com/okex/exchain/libs/tendermint/libs/rand"
	"sync"
	"time"
)

type cw20Bench struct {
	core.BaseBench
}

func NewCW20Bench(option *options.CW20TransferOption) (*cw20Bench, error) {
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

	buildTxFn := func(sender int, accounts []*account.Account) (stdTx []*auth.StdTx) {
		account := accounts[sender]
		execMsg := fmt.Sprintf(`{"transfer": {"amount":"1", "recipient":"%s"}}`, randAddress(accounts))
		tx, err := core.BuildWasmTx(account.GetPrivateKey(), account.GetAccountNumber(), account.GetNonce(), option.ChainId, "", option.CW20Address, execMsg, *account.GetBech32Address(), "")
		if err != nil {
			panic(err)
		}

		return []*auth.StdTx{tx}
	}

	bench := cw20Bench{
		core.BaseBench{
			Accounts:          accounts,
			Concurrency:       option.ConcurrentNum,
			BuildTxFn:         buildTxFn,
			TendermintClients: clients,
		},
	}

	return &bench, nil
}

func randAddress(accounts []*account.Account) string {
	acc := accounts[rand.Intn(len(accounts)-1)]
	return acc.GetBech32Address().Bech32StringOptimized("ex")
}

func initAccounts(accounts []*account.Account, clients []*client.CosmosClient) error {

	wg := &sync.WaitGroup{}

	goroutineNum := len(accounts) / 50
	if goroutineNum > 1000 {
		goroutineNum = 1000
	}

	gap := len(accounts) / goroutineNum
	remain := len(accounts) % goroutineNum

	for i := 0; i < goroutineNum; i++ {
		go func(gIndex int) {
			wg.Add(1)
			defer wg.Done()

			start := gIndex * gap
			end := (gIndex + 1) * gap
			if gIndex == goroutineNum-1 && remain != 0 {
				end += remain
			}

			client := clients[goroutineNum%len(clients)]

			for i := start; i < end; i++ {
				account := accounts[i]
				for account.Init(client) != nil {
					fmt.Println(account.GetHexAddress(), "init failed")
					time.Sleep(500 * time.Millisecond)
				}
				if i%200 == 0 {
					fmt.Println("init ", i)
				}
			}
		}(i)
	}

	wg.Wait()
	return nil
}
