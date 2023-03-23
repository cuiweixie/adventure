package transfer_okt

import (
	"fmt"
	gosdk "github.com/okex/exchain-go-sdk"
	"github.com/okex/exchain-go-sdk/types/tx"
	"github.com/okex/exchain-go-sdk/utils"
	"github.com/okex/exchain/libs/cosmos-sdk/crypto/keys"
	"github.com/okex/exchain/libs/cosmos-sdk/types"
	"sync"
)

type CosmosAccount struct {
	lock       *sync.Mutex
	account    types.Account
	privateKey keys.Info
}

func GenerateAccounts(privkeys []string) (accounts []*CosmosAccount) {
	for id, priStr := range privkeys {
		tx.Kb.ImportPrivKey(fmt.Sprintf("%d", id), priStr, "")
		fmt.Println(tx.Kb.Get(fmt.Sprintf("%d", id)))
		key, _ := utils.CreateAccountWithPrivateKey(priStr, "", "")
		accounts = append(accounts, &CosmosAccount{new(sync.Mutex), nil, key})
	}
	return
}

func (a *CosmosAccount) Lock() {
	a.lock.Lock()
}

func (a *CosmosAccount) Unlock() {
	a.lock.Unlock()
}

func (a *CosmosAccount) initAccount(client *gosdk.Client) error {
	account, err := client.Auth().QueryAccount(a.privateKey.GetAddress().String())
	if err != nil {
		return err
	}
	a.account = account

	return nil
}

func (a *CosmosAccount) GetNonce() uint64 {
	return a.account.GetSequence()
}

func (a *CosmosAccount) GetAccNumber() uint64 {
	return a.account.GetAccountNumber()
}

func (a *CosmosAccount) GetPrivateKey() keys.Info {
	return a.privateKey
}
