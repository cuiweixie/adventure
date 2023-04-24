package cwoperate

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/okex/exchain-go-sdk/types"
	"github.com/okex/exchain/libs/cosmos-sdk/types/errors"

	"github.com/okex/adventure/common/client"
	cmwraptx "github.com/okex/adventure/common/types"
	"github.com/okex/adventure/wasm/bench/common/account"
	"github.com/okex/adventure/wasm/bench/core"
	"github.com/okex/adventure/wasm/bench/options"
)

type cwoperateBench struct {
	core.BaseBench
}

const (
	computeContractPath = "./config/devnet/wasm_contract/computeTest.wasm"
	writeContractPath   = "./config/devnet/wasm_contract/writeTest.wasm"
	readContractPath    = "./config/devnet/wasm_contract/readTest.wasm"
	routerContractPath  = "./config/devnet/wasm_contract/router.wasm"
	computeContractMsg  = `{"operate" : {"opts": ["1","1","1","1","1"],"times":"1"}}`
	writeContractMsg    = `{"operate" : {"opts": ["1","1","1","1","1","1"],"times":"1"}}`
	readContractMsg     = `{"operate" : {"opts": ["1","1","1","1"],"times":"1"}}`
	computeRouterMsg    = `{"operate" : {"contract_name": "computeTest","opts": ["1","1","1","1","1"],"times":"1"}}`
	writeRouterMsg      = `{"operate" : {"contract_name": "writeTest","opts": ["1","1","1","1","1","1"],"times":"1"}}`
	readRouterMsg       = `{"operate" : {"contract_name": "readTest","opts": ["1","1","1","1"],"times":"1"}}`
)

func NewCWOperateBench(option *options.CWOperateOption) (*cwoperateBench, error) {
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

	//var initMsg string
	var contractPath string
	var routerMsg string
	var execMsg string
	// use account[0] to deploy contract
	if option.ContractAddress == "" {
		//if option.WasmFilePath == "" {
		//	return nil, fmt.Errorf("should provide wasm file")
		//}
		switch option.WasmOperType {
		case "compute":
			contractPath = computeContractPath
			execMsg = computeContractMsg
		case "write":
			contractPath = writeContractPath
			execMsg = writeContractMsg
		case "read":
			contractPath = readContractPath
			execMsg = readContractMsg
		default:
			return nil, errors.Wrap(err, "resolve wasm contract name failed:wrong WasmOperType")
		}

		addr, err := deployCW20(accounts[0], clients[0], contractPath, "{}")

		if err != nil {
			return nil, errors.Wrap(err, "deploy wasm contract failed")
		}

		if option.WasmOperRouter {
			// deploy router contract
			raddr, err := deployCW20(accounts[0], clients[0], routerContractPath, "{}")
			if err != nil {
				return nil, errors.Wrap(err, "deploy router contract failed")
			}
			// register route
			switch option.WasmOperType {
			case "compute":
				routerMsg = fmt.Sprintf(`{"set_contract":{"contract_name":"computeTest","contract_addr":"%s"}}`, addr)
				execMsg = computeRouterMsg
			case "write":
				routerMsg = fmt.Sprintf(`{"set_contract":{"contract_name":"writeTest","contract_addr":"%s"}}`, addr)
				execMsg = writeRouterMsg
			case "read":
				routerMsg = fmt.Sprintf(`{"set_contract":{"contract_name":"readTest","contract_addr":"%s"}}`, addr)
				execMsg = readRouterMsg
			default:
				return nil, errors.Wrap(err, "resolve wasm contract name failed:wrong WasmOperType")
			}

			rClient := clients[0]
			rAccount := accounts[0]
			sender := (*rAccount).GetBech32Address()
			_, err = rClient.SendWasmTx(rAccount.GetPrivateKey(), rAccount.GetAccountNumber(), rAccount.GetNonce(), option.ChainId, "", raddr, routerMsg, *sender, "1okt")

			if err != nil {
				return nil, errors.Wrap(err, "add route to router contract failed")
			}
			log.Printf("register wasm %s contract route success", option.WasmOperType)
			option.ContractAddress = raddr
		} else {
			option.ContractAddress = addr
		}

		log.Printf("deploy wasm %s contract success: %s", option.WasmOperType, option.ContractAddress)

		// reinit account 0
		for accounts[0].Init(clients[0]) != nil {
			time.Sleep(500 * time.Millisecond)
		}
	}

	buildTxFn := func(sender int, accounts []*account.Account, client *client.CosmosClient) (stdTx []*cmwraptx.WrapCMTx) {
		account := accounts[sender]
		defer account.AddNonce()

		//execMsg := fmt.Sprintf(`{"transfer":{"recipient":"%s"}}`, randAddress(accounts))
		tx, err := core.BuildWasmTx(account.GetPrivateKey(), account.GetAccountNumber(), account.GetNonce(), option.ChainId, "", option.ContractAddress, execMsg, (*account).GetHexAddress().String(), "1okt")
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

	bench := cwoperateBench{
		core.BaseBench{
			Accounts:          accounts,
			Concurrency:       option.ConcurrentNum,
			BuildTxFn:         buildTxFn,
			TendermintClients: clients,
		},
	}

	return &bench, nil
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
