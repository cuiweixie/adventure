package transfer_okt

import (
	"crypto/ecdsa"
	"fmt"
	ethcmm "github.com/ethereum/go-ethereum/common"
	gosdk "github.com/okex/exchain-go-sdk"
	sdk "github.com/okex/exchain/libs/cosmos-sdk/types"
	"sync"

	"github.com/ethereum/go-ethereum/crypto"
)

type EthAccount struct {
	lock       *sync.Mutex
	nonce      uint64
	accountNum uint64
	queried    bool
	privateKey *ecdsa.PrivateKey
}

func generateAccounts(privkeys []string) (accounts []*EthAccount) {
	for _, p := range privkeys {
		privateKey, err := crypto.HexToECDSA(p)
		if err != nil {
			panic(err)
		}

		accounts = append(accounts, &EthAccount{new(sync.Mutex), 0, 0, false, privateKey})
	}
	return
}

func (a *EthAccount) Lock() {
	a.lock.Lock()
}

func (a *EthAccount) Unlock() {
	a.lock.Unlock()
}

func (a *EthAccount) SetNonce(cli *gosdk.Client) error {
	if a.queried {
		return nil
	}
	ethAddr := a.GetHexAddress()
	bench32Addr := sdk.AccAddress(ethAddr.Bytes())
	account, err := cli.Auth().QueryAccount(bench32Addr.Bech32StringOptimized("ex"))
	if err != nil {
		panic(err)
	}

	a.nonce = account.GetSequence()
	a.accountNum = account.GetAccountNumber()
	a.queried = true
	return nil
}

func (a *EthAccount) AddNonce() {
	a.Lock()
	defer a.Unlock()
	a.nonce += 1
}

func (a *EthAccount) GetNonce() uint64 {
	a.Lock()
	defer a.Unlock()
	return a.nonce
}

func (a *EthAccount) GetAccountNumber() uint64 {
	return a.accountNum
}

func (a *EthAccount) GetPrivateKey() *ecdsa.PrivateKey {
	return a.privateKey
}

func (a *EthAccount) GetHexAddress() ethcmm.Address {
	pubkeyECDSA, ok := a.privateKey.Public().(*ecdsa.PublicKey)
	if ok != true {
		panic(fmt.Errorf("convert into pubkey failed"))
	}
	return crypto.PubkeyToAddress(*pubkeyECDSA)
}
