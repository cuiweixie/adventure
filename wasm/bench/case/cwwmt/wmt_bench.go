package cwwmt

import (
	"encoding/base64"
	"fmt"
	"log"
	"path/filepath"
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

const (
	cw20FileName    = "cw20_base.wasm"
	pairFileName    = "terraswap_pair.wasm"
	lpFileName      = "lp_staking.wasm"
	factoryFileName = "terraswap_factory.wasm"
	routerFileName  = "terraswap_router.wasm"
)

type cwwmtBench struct {
	core.BaseBench
}

type StakeFlag struct {
	mp []bool
}

func NewCWWMTBench(option *options.CWWMTOption) (*cwwmtBench, error) {
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
	if !option.ContractAvailable || option.ContractSetNum != len(option.WmtContracts) {
		err = wmtContractDeploy(option, accounts[0], clients[0])
		if err != nil {
			return nil, errors.Wrap(err, "Wmt Contract Deploy Error")
		}
	}

	buildTxFn := func(sender int, accounts []*account.Account, client *client.CosmosClient) (stdTx []*cmwraptx.WrapCMTx) {
		contractInd := sender % option.ContractSetNum
		account := accounts[sender]

		var tempWmtTx *cmwraptx.WrapCMTx
		wmtTxs := []*cmwraptx.WrapCMTx{}
		//Add Liquidity
		tempWmtTx = addLiquidityTxBuilder(client, account, option, contractInd)
		wmtTxs = append(wmtTxs, tempWmtTx)
		//SwapAtoB
		tempWmtTx = swapAToBTxBuilder(client, account, "AtoB", option, contractInd)
		wmtTxs = append(wmtTxs, tempWmtTx)
		//SwapBtoA
		tempWmtTx = swapAToBTxBuilder(client, account, "BtoA", option, contractInd)
		wmtTxs = append(wmtTxs, tempWmtTx)
		//Stake
		tempWmtTx = stakeTxBuilder(client, account, option, contractInd)
		wmtTxs = append(wmtTxs, tempWmtTx)
		//Withdraw
		tempWmtTx = withdrawTxBuilder(client, account, option, contractInd)
		wmtTxs = append(wmtTxs, tempWmtTx)

		return wmtTxs
	}

	bench := cwwmtBench{
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

func swapBase(ask string, offer string) string {
	msg := fmt.Sprintf(`{
    "execute_swap_operations": {
      "minimum_receive": "0",
      "operations": [
        {
          "terra_swap": {
            "ask_asset_info": {
              "token": {
                "contract_addr": "%s"
              }
            },
            "offer_asset_info": {
              "token": {
                "contract_addr": "%s"
              }
            }
          }
        }
      ]
    }
  }`, ask, offer)
	return base64.StdEncoding.EncodeToString([]byte(msg))
}

func boundBase() string {
	msg := fmt.Sprintf(`{ "bond": {} }`)
	return base64.StdEncoding.EncodeToString([]byte(msg))
}

func wmtContractDeploy(option *options.CWWMTOption, account *account.Account, client *client.CosmosClient) error {

	privKey := account.GetPrivateKey()
	accNumber := account.GetAccountNumber()
	seqNumber := account.GetNonce()
	sender := account.GetBech32Address()
	ethaddress := account.GetHexAddress().String()

	//amountStr := "1okb"

	var wasmPath string

	// Store code
	wasmPath = filepath.Join(option.ContractFolderPath, cw20FileName)
	cw20Id, err := client.StoreCode(privKey, accNumber, seqNumber, wasmPath, *sender)
	if err != nil {
		return err
	}
	seqNumber += 1
	account.AddNonce()

	wasmPath = filepath.Join(option.ContractFolderPath, pairFileName)
	pairId, err := client.StoreCode(privKey, accNumber, seqNumber, wasmPath, *sender)
	if err != nil {
		return err
	}
	seqNumber += 1
	account.AddNonce()

	wasmPath = filepath.Join(option.ContractFolderPath, lpFileName)
	lpId, err := client.StoreCode(privKey, accNumber, seqNumber, wasmPath, *sender)
	if err != nil {
		return err
	}
	seqNumber += 1
	account.AddNonce()

	wasmPath = filepath.Join(option.ContractFolderPath, factoryFileName)
	factoryId, err := client.StoreCode(privKey, accNumber, seqNumber, wasmPath, *sender)
	if err != nil {
		return err
	}
	seqNumber += 1
	account.AddNonce()

	wasmPath = filepath.Join(option.ContractFolderPath, routerFileName)
	routerId, err := client.StoreCode(privKey, accNumber, seqNumber, wasmPath, *sender)
	if err != nil {
		return err
	}
	seqNumber += 1
	account.AddNonce()

	//Deploy multi set of WMT contract
	option.WmtContracts = make([]options.WmtContract, option.ContractSetNum)
	var initMsg, execMsg string
	for i := 0; i < option.ContractSetNum; i++ {
		// Instantiate Atoken
		initMsg = fmt.Sprintf(`{
		"name": "testA",
		"symbol": "testA", 
		"decimals": 6, 
		"initial_balances": [
		   { 
			 "address": "%s",
			 "amount": "100000000000000000000" 
		   }
		 ], 
		 "mint": {
			"minter": "%s",
			"cap": "10000000000000000000000000000000"
		 }
	}`, ethaddress, ethaddress)
		aTokenAddr, err := client.InstantiateContract(privKey, accNumber, seqNumber, cw20Id, initMsg, *sender, "", "deploy", sender.String())
		if err != nil {
			return err
		}
		seqNumber += 1
		account.AddNonce()
		option.WmtContracts[i].A = aTokenAddr

		// Instantiate Btoken
		initMsg = fmt.Sprintf(`{
		"name": "testB",
		"symbol": "testB", 
		"decimals": 6, 
		"initial_balances": [
		   { 
			 "address": "%s",
			 "amount": "100000000000000000000" 
		   }
		 ], 
		 "mint": {
			"minter": "%s",
			"cap": "10000000000000000000000000000000"
		 }
	}`, ethaddress, ethaddress)
		bTokenAddr, err := client.InstantiateContract(privKey, accNumber, seqNumber, cw20Id, initMsg, *sender, "", "deploy", sender.String())
		if err != nil {
			return err
		}
		seqNumber += 1
		account.AddNonce()
		option.WmtContracts[i].B = bTokenAddr

		// Instantiate Factory
		initMsg = fmt.Sprintf(`{"pair_code_id": %d, "token_code_id": %d }`, pairId, cw20Id)
		factoryAddr, err := client.InstantiateContract(privKey, accNumber, seqNumber, factoryId, initMsg, *sender, "", "deploy", sender.String())
		if err != nil {
			return err
		}
		seqNumber += 1
		account.AddNonce()
		option.WmtContracts[i].Factory = factoryAddr

		// Instantiate router
		initMsg = fmt.Sprintf(`{"terraswap_factory": "%s"}`, factoryAddr)
		routerAddr, err := client.InstantiateContract(privKey, accNumber, seqNumber, routerId, initMsg, *sender, "", "deploy", sender.String())
		if err != nil {
			return err
		}
		seqNumber += 1
		account.AddNonce()
		option.WmtContracts[i].Router = routerAddr

		// Create Pair & Query Pair
		execMsg = fmt.Sprintf(`{
		"create_pair": {
			"asset_infos":[
				{"token": {
					"contract_addr": "%s"
					}
				},
				{"token": { 
					"contract_addr": "%s" 
					}
				}
			]
		}
	}`, aTokenAddr, bTokenAddr)
		pairAddr, err := client.CreatePair(privKey, accNumber, seqNumber, option.ChainId, "", factoryAddr, execMsg, *sender, "")
		if err != nil {
			return err
		}
		seqNumber += 1
		account.AddNonce()
		option.WmtContracts[i].Pair = pairAddr

		// Instantiate rewardToken
		initMsg = fmt.Sprintf(`{
		"name": "reward",
		"symbol": "reward", 
		"decimals": 6, 
		"initial_balances": [
		   { 
			 "address": "%s",
			 "amount": "100000000000000000000" 
		   }
		 ], 
		 "mint": {
			"minter": "%s",
			"cap": "10000000000000000000000000000000"
		 }
	}`, ethaddress, ethaddress)
		rewardTokenAddr, err := client.InstantiateContract(privKey, accNumber, seqNumber, cw20Id, initMsg, *sender, "", "deploy", sender.String())
		if err != nil {
			return err
		}
		seqNumber += 1
		account.AddNonce()
		option.WmtContracts[i].RewardToken = rewardTokenAddr

		// Instantiate stakeToken
		initMsg = fmt.Sprintf(`{
		"name": "staking",
		"symbol": "staking", 
		"decimals": 6, 
		"initial_balances": [
		   { 
			 "address": "%s",
			 "amount": "100000000000000000000" 
		   }
		 ], 
		 "mint": {
			"minter": "%s",
			"cap": "10000000000000000000000000000000"
		 }
	}`, ethaddress, ethaddress)
		stakingTokenAddr, err := client.InstantiateContract(privKey, accNumber, seqNumber, cw20Id, initMsg, *sender, "", "deploy", sender.String())
		if err != nil {
			return err
		}
		seqNumber += 1
		account.AddNonce()
		option.WmtContracts[i].ABLPToken = stakingTokenAddr

		// Instantiate lp staking
		initMsg = fmt.Sprintf(`{
		"owner":"%s", 
		"reward_token":"%s", 
		"staking_token":"%s", 
		"distribution_schedule":[[1682584812, 1714120812, "100000000000"]]
	}`, ethaddress, rewardTokenAddr, stakingTokenAddr)
		lpAddr, err := client.InstantiateContract(privKey, accNumber, seqNumber, lpId, initMsg, *sender, "", "deploy", sender.String())
		if err != nil {
			return err
		}
		seqNumber += 1
		account.AddNonce()
		option.WmtContracts[i].LPStaking = lpAddr

		// Initialize lp staking account
		execMsg = fmt.Sprintf(`{"transfer":{"recipient": "%s", "amount": "100000000000" }}`, lpAddr)
		_, err = client.SendWasmTx(privKey, accNumber, seqNumber, option.ChainId, "", rewardTokenAddr, execMsg, *sender, "")
		if err != nil {
			return err
		}
		seqNumber += 1
		account.AddNonce()

		// Output deployment result
		log.Printf("Wasm wmt contract set %d is successfully deployed", i+1)
		log.Printf("Atoken:%s,", aTokenAddr)
		log.Printf("Btoken:%s", bTokenAddr)
		log.Printf("Factory:%s", factoryAddr)
		log.Printf("Pair:%s", pairAddr)
		log.Printf("Router:%s", routerAddr)
		log.Printf("RewardToken:%s", rewardTokenAddr)
		log.Printf("ABLPToken:%s", stakingTokenAddr)
		log.Printf("LPStaking:%s", lpAddr)
	}

	return nil
}

func addLiquidityTxBuilder(client *client.CosmosClient, acc *account.Account, option *options.CWWMTOption, index int) *cmwraptx.WrapCMTx {
	execMsg := fmt.Sprintf(`{
    "provide_liquidity":{
        "assets": [
          {
            "amount": "1000",
            "info": {
              "token": {
                "contract_addr": "%s"
              }
            }
          },
          {
            "amount": "1000",
            "info": {
              "token": {
                "contract_addr": "%s"
              }
            }
          }
        ]
      }
     }`, option.WmtContracts[index].A, option.WmtContracts[index].B)

	tx, err := core.BuildWasmTx(acc.GetPrivateKey(), acc.GetAccountNumber(), acc.GetNonce(), option.ChainId, "", option.WmtContracts[index].Pair, execMsg, (*acc).GetHexAddress().String(), "")
	if err != nil {
		panic(err)
	}

	cli := client.Auth().(types.BaseClient)
	bytes, err := cli.GetCodec().MarshalBinaryLengthPrefixed(tx)

	wrapedTx := &cmwraptx.WrapCMTx{
		Tx:    bytes,
		Nonce: acc.GetNonce(),
	}

	acc.AddNonce()
	return wrapedTx
}

func swapAToBTxBuilder(client *client.CosmosClient, acc *account.Account, txMsg string, option *options.CWWMTOption, index int) *cmwraptx.WrapCMTx {

	var askAddr, offerAddr string
	if txMsg == "AtoB" {
		askAddr = option.WmtContracts[index].A
		offerAddr = option.WmtContracts[index].B
	} else if txMsg == "BtoA" {
		askAddr = option.WmtContracts[index].B
		offerAddr = option.WmtContracts[index].A
	} else {
		panic("wrong TxMsg in swapAToBTxBuilder")
	}
	reviverBase64 := swapBase(askAddr, offerAddr)
	execMsg := fmt.Sprintf(`{
      "send": { 
          "contract": "%s", 
          "amount": "10", 
          "msg": "%s"
          }
      }`, option.WmtContracts[index].Router, reviverBase64)

	tx, err := core.BuildWasmTx(acc.GetPrivateKey(), acc.GetAccountNumber(), acc.GetNonce(), option.ChainId, "", offerAddr, execMsg, (*acc).GetHexAddress().String(), "")
	if err != nil {
		panic(err)
	}

	cli := client.Auth().(types.BaseClient)
	bytes, err := cli.GetCodec().MarshalBinaryLengthPrefixed(tx)

	wrapedTx := &cmwraptx.WrapCMTx{
		Tx:    bytes,
		Nonce: acc.GetNonce(),
	}

	acc.AddNonce()
	return wrapedTx
}

func stakeTxBuilder(client *client.CosmosClient, acc *account.Account, option *options.CWWMTOption, index int) *cmwraptx.WrapCMTx {
	reviverBase64 := boundBase()
	execMsg := fmt.Sprintf(`{
                  "send":
                      { 
                        "contract": "%s", 
                        "amount": "1", 
                        "msg": "%s"
                       }
                 }`, option.WmtContracts[index].LPStaking, reviverBase64)

	tx, err := core.BuildWasmTx(acc.GetPrivateKey(), acc.GetAccountNumber(), acc.GetNonce(), option.ChainId, "", option.WmtContracts[index].ABLPToken, execMsg, (*acc).GetHexAddress().String(), "")
	if err != nil {
		panic(err)
	}

	cli := client.Auth().(types.BaseClient)
	bytes, err := cli.GetCodec().MarshalBinaryLengthPrefixed(tx)

	wrapedTx := &cmwraptx.WrapCMTx{
		Tx:    bytes,
		Nonce: acc.GetNonce(),
	}

	acc.AddNonce()
	return wrapedTx
}

func withdrawTxBuilder(client *client.CosmosClient, acc *account.Account, option *options.CWWMTOption, index int) *cmwraptx.WrapCMTx {
	execMsg := fmt.Sprintf(`{"withdraw": {}}`)

	tx, err := core.BuildWasmTx(acc.GetPrivateKey(), acc.GetAccountNumber(), acc.GetNonce(), option.ChainId, "", option.WmtContracts[index].LPStaking, execMsg, (*acc).GetHexAddress().String(), "")
	if err != nil {
		panic(err)
	}

	cli := client.Auth().(types.BaseClient)
	bytes, err := cli.GetCodec().MarshalBinaryLengthPrefixed(tx)

	wrapedTx := &cmwraptx.WrapCMTx{
		Tx:    bytes,
		Nonce: acc.GetNonce(),
	}

	acc.AddNonce()
	return wrapedTx
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
