package common

import (
	"crypto/ecdsa"
	"fmt"
	ethcmm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/okex/exchain-go-sdk/utils"
	"github.com/okex/exchain/libs/cosmos-sdk/types"
)

func GetEthAddressFromPK(privateKey *ecdsa.PrivateKey) ethcmm.Address {
	pubkeyECDSA, ok := privateKey.Public().(*ecdsa.PublicKey)
	if ok != true {
		panic(fmt.Errorf("convert into pubkey failed"))
	}
	return crypto.PubkeyToAddress(*pubkeyECDSA)
}

func GetCosmosAddressFromPrivateKey(privateKey string) types.Address {
	info, _ := utils.CreateAccountWithPrivateKey(privateKey, "", "")
	return info.GetAddress()
}
