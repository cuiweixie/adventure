package account

import (
	"crypto/ecdsa"
	"fmt"
	ethcmm "github.com/ethereum/go-ethereum/common"
	gosdk "github.com/okex/exchain-go-sdk"
	sdk "github.com/okex/exchain/libs/cosmos-sdk/types"
	"sync"

	"github.com/ethereum/go-ethereum/crypto"
)

type Account struct {
	lock          *sync.Mutex
	nonce         uint64
	accountNum    uint64
	privateKey    *ecdsa.PrivateKey
	hexAddres     *ethcmm.Address
	bech32Address *sdk.AccAddress
}

func NewAccount(privateKey string) (*Account, error) {
	priKey, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, err
	}

	pubkeyECDSA, ok := priKey.Public().(*ecdsa.PublicKey)
	if !ok {

		return nil, fmt.Errorf("convert into pubkey failed")
	}

	hexAddress := crypto.PubkeyToAddress(*pubkeyECDSA)
	bench32Addr := sdk.AccAddress(hexAddress.Bytes())

	account := &Account{
		lock:          &sync.Mutex{},
		nonce:         0,
		accountNum:    0,
		privateKey:    priKey,
		hexAddres:     &hexAddress,
		bech32Address: &bench32Addr,
	}

	return account, nil
}

func (a *Account) Lock() {
	a.lock.Lock()
}

func (a *Account) Unlock() {
	a.lock.Unlock()
}

func (a *Account) Init(cli *gosdk.Client) error {
	account, err := cli.Auth().QueryAccount(a.bech32Address.Bech32StringOptimized("ex"))
	if err != nil {
		return err
	}
	a.nonce = account.GetSequence()
	a.accountNum = account.GetAccountNumber()
	return nil
}

func (a *Account) AddNonce() {
	a.Lock()
	a.Unlock()
	a.nonce += 1
}

func (a *Account) GetNonce() uint64 {
	a.Lock()
	a.Unlock()
	return a.nonce
}

func (a *Account) GetAccountNumber() uint64 {
	return a.accountNum
}

func (a *Account) GetPrivateKey() *ecdsa.PrivateKey {
	return a.privateKey
}

func (a *Account) GetHexAddress() *ethcmm.Address {
	return a.hexAddres
}
