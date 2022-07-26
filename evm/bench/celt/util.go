package celt

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"io/ioutil"
	"math/big"
	"time"
)

func getPrivateKey(key string) *ecdsa.PrivateKey {
	privateKey, err := crypto.HexToECDSA(key)
	panicerr(err)
	return privateKey
}

func transferOkt(key string, to common.Address, nonce uint64, value *big.Int) *types.Transaction {
	privateKey := getPrivateKey(key)

	tx, err := types.SignTx(types.NewTransaction(nonce, to, value, gasLimit, gasPrice, nil), signer, privateKey)
	panicerr(err)
	return tx
}

func SignTxWithNonce(privateKey *ecdsa.PrivateKey, to common.Address, payLoad []byte, nonce uint64) *types.Transaction {
	tx, err := types.SignTx(types.NewTransaction(nonce, to, new(big.Int), gasLimit, gasPrice, payLoad), signer, privateKey)
	panicerr(err)
	return tx
}

func panicerr(err error) {
	if err != nil {
		panic(err)
	}
}

func LoadContractList(file string) []CeltContract {
	data, err := ioutil.ReadFile(file)
	panicerr(err)
	cList := make([]CeltContract, 0)
	err = json.Unmarshal(data, &cList)
	panicerr(err)
	return cList
}

func keyToAcc(key string) *acc {
	privateKey := getPrivateKey(key)
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		panic("keyToAcc")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	return &acc{
		privateKey: key,
		ecdsaPriv:  privateKey,
		ethAddress: fromAddress,
	}
}

func SendTxs(client *ethclient.Client, txs []*types.Transaction) error {
	for index, v := range txs {
		//fmt.Println(v.Hash())
		cnt := 0
		var err error
		for cnt < 50 {
			if e := client.SendTransaction(context.Background(), v); e != nil {
				fmt.Println("index", index, e)
				err = e
				time.Sleep(time.Second * 5)
			} else {
				err = nil
				break
			}
			cnt++
		}
		if err != nil {
			return err
		}

		if index != 0 && index%200 == 0 {
			fmt.Println("send tx index", index)
		}
	}
	return nil
}

func loadCeltConfig(file string) *CeltConfig {
	data, err := ioutil.ReadFile(file)
	panicerr(err)
	c := new(CeltConfig)
	err = json.Unmarshal(data, c)
	panicerr(err)
	return c
}

func StringToBytes32(str string) [32]byte {
	value := [32]byte{}
	copy(value[:], str[:])
	return value
}
